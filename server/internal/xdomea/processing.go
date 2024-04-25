package xdomea

import (
	"archive/zip"
	"fmt"
	"io"
	"lath/xman/internal/auth"
	"lath/xman/internal/db"
	"lath/xman/internal/format"
	"lath/xman/internal/mail"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/lestrrat-go/libxml2/xsd"
)

// ProcessNewMessage
func ProcessNewMessage(agency db.Agency, transferDirMessagePath string) {
	defer HandlePanic(fmt.Sprintf("ProcessNewMessage %s", transferDirMessagePath))
	// extract process ID from message filename
	processID := GetMessageID(transferDirMessagePath)
	// extract message type from message filename
	messageType, err := GetMessageTypeImpliedByPath(transferDirMessagePath)
	// copy message from transfer directory to a local temporary directory
	localMessagePath := CopyMessageFromTransferDirectory(agency, transferDirMessagePath)
	defer os.Remove(localMessagePath)
	// extract message to message storage
	processStoreDir, messageStoreDir, err := extractMessageToMessageStore(
		agency,
		transferDirMessagePath,
		localMessagePath,
		processID,
		messageType,
	)
	// error happened while extracting the message
	if err != nil {
		HandleError(db.ProcessingError{
			Agency:         &agency,
			TransferPath:   &transferDirMessagePath,
			Description:    "Fehler beim Entpacken der Nachricht",
			AdditionalInfo: err.Error(),
		})
	}
	// save message
	process, message, err := AddMessage(
		agency,
		processID,
		messageType,
		processStoreDir,
		messageStoreDir,
		transferDirMessagePath,
	)
	if err != nil {
		HandleError(err)
	}
	err = compareAgencyFields(agency, message, process)
	if err != nil {
		HandleError(err)
	}
	err = checkMaxRecordObjectDepth(agency, process, message)
	if err != nil {
		HandleError(err)
	}
	if messageType.Code == "0503" {
		// get primary documents
		primaryDocuments := db.GetAllPrimaryDocuments(message.ID)
		err = checkMessage0503Integrity(process, message, primaryDocuments)
		if err != nil {
			HandleError(err)
		}
		recordFileSizes(message, primaryDocuments)
		// start format verification
		if os.Getenv("BORG_ENDPOINT") != "" {
			err = format.VerifyFileFormats(process, message)
			if err != nil {
				HandleError(err)
			}
		}
	}
	// if no error occurred while processing the message
	if err == nil {
		// send the confirmation message that the 0501 message was received
		if messageType.Code == "0501" {
			messagePath := Send0504Message(agency, message)
			db.UpdateProcess(process.ID, db.Process{
				Message0504Path: &messagePath,
			})
		}
		// send e-mail notification to users
		for _, user := range agency.Users {
			address := auth.GetMailAddress(user.ID)
			preferences := db.GetUserInformation(user.ID).Preferences
			if preferences.MessageEmailNotifications {
				mail.SendMailNewMessage(address, agency.Name, message)
			}
		}
	}
}

// AddMessage parses the message and saves it in the database.
//
// It returns an error when a reading the message resulted in any processing error.
func AddMessage(
	agency db.Agency,
	processID string,
	messageType db.MessageType,
	processStoreDir string,
	messageStoreDir string,
	transferDirMessagePath string,
) (db.Process, db.Message, error) {
	var process db.Process
	var message db.Message
	messageName := GetMessageName(processID, messageType)
	messagePath := path.Join(messageStoreDir, messageName)
	_, err := os.Stat(messagePath)
	if err != nil {
		panic("message doesn't exist: " + messagePath)
	}
	message = db.Message{
		MessageType:     messageType,
		TransferDirPath: transferDirMessagePath,
		StoreDir:        messageStoreDir,
		MessagePath:     messagePath,
	}
	err = checkMessageValidity(agency, &message, transferDirMessagePath)
	if err != nil {
		return process, message, err
	}
	// parse message
	message, err = ParseMessage(message)
	if err != nil {
		return process, message, err
	}
	// count the maximal record object depth within the message
	message.MaxRecordObjectDepth = message.GetMaxChildDepth()
	// Store message in database with parsed message metadata.
	return db.AddMessage(agency, processID, processStoreDir, message)
}

// checkMessageValidity performs a xsd schema validation against the message XML file.
func checkMessageValidity(agency db.Agency, message *db.Message, transferDirMessagePath string) error {
	xdomeaVersion, err := ExtractVersionFromMessage(*message)
	if err != nil {
		return err
	}
	messageIsValid, err := ValidateXdomeaXmlFile(message.MessagePath, xdomeaVersion)
	message.SchemaValidation = messageIsValid
	if err != nil {
		if !messageIsValid {
			validationError, ok := err.(xsd.SchemaValidationError)
			if ok {
				var errorMessages []string
				for _, e := range validationError.Errors() {
					errorMessages = append(errorMessages, e.Error())
					log.Printf("XML schema error: %s", e.Error())
				}
				additionalInfo := strings.Join(errorMessages, "\n")
				return db.ProcessingError{
					Agency:         &agency,
					TransferPath:   &transferDirMessagePath,
					Description:    "Schema-Validierung ungültig",
					AdditionalInfo: additionalInfo,
				}
			} else {
				return db.ProcessingError{
					Agency:       &agency,
					TransferPath: &transferDirMessagePath,
					Description:  "Schema-Validierung ungültig",
				}
			}
		}
		return err
	}
	return nil
}

func checkMessage0503Integrity(
	process db.Process,
	message0503 db.Message,
	primaryDocuments []db.PrimaryDocument,
) error {
	// check if all primary document files exist
	for _, primaryDocument := range primaryDocuments {
		filePath := path.Join(message0503.StoreDir, primaryDocument.FileName)
		_, err := os.Stat(filePath)
		if err != nil {
			return db.ProcessingError{
				Process:     &process,
				Description: "Primärdatei fehlt in Abgabe",
				Message:     &message0503,
			}
		}
	}
	// check if 0501 message exists
	_, found := db.GetMessageOfProcessByCode(process, "0501")
	// 0501 Message doesn't exist. No further message validation necessary.
	if !found {
		return nil
	}
	// check if appraisal of 0501 message is already complete
	if !process.ProcessState.Appraisal.Complete {
		errorMessage := "Die Abgabe wurde erhalten, bevor die Bewertung der Anbietung abgeschlossen wurde"
		return db.ProcessingError{
			Process:     &process,
			Description: errorMessage,
			Message:     &message0503,
		}
	} else {
		return checkRecordObjectsOfMessage0503(process, message0503)
	}
}

// recordFileSizes reads the primary documents' file sizes on disk and saves the
// numbers to the database record.
func recordFileSizes(message db.Message, documents []db.PrimaryDocument) {
	for _, d := range documents {
		filePath := filepath.Join(message.StoreDir, d.FileName)
		s, err := os.Stat(filePath)
		if err != nil {
			panic(err)
		}
		size := s.Size()
		db.UpdatePrimaryDocument(d.ID, db.PrimaryDocument{
			FileSize: uint64(size),
		})
	}
}

// checkRecordObjectsOfMessage0503 compares a 0503 message with the appraisal of
// a 0501 message and returns an error if there are record objects missing in
// the 0503 message that were marked as to be archived in the appraisal or if
// there are any surplus objects included in the 0503 message.
func checkRecordObjectsOfMessage0503(
	process db.Process,
	message0503 db.Message,
) error {
	// Gather data
	appraisals := make(map[uuid.UUID]db.Appraisal)
	for _, a := range db.GetAppraisalsForProcess(process.ID) {
		appraisals[a.RecordObjectID] = a
	}

	includedRecordObjects := make(map[uuid.UUID]db.AppraisableRecordObject)
	for _, f := range db.GetAllFileRecordObjects(message0503.ID) {
		includedRecordObjects[f.XdomeaID] = &f
	}
	for _, p := range db.GetAllProcessRecordObjects(message0503.ID) {
		includedRecordObjects[p.XdomeaID] = &p
	}

	// Check for objects missing from the 0503 message
	var missingRecordObjects []string
	for id, a := range appraisals {
		if a.Decision == db.AppraisalDecisionA && includedRecordObjects[id] == nil {
			missingRecordObjects = append(missingRecordObjects, id.String())
		}
	}
	if len(missingRecordObjects) > 0 {
		errorMessage := "Die Abgabe ist nicht vollständig"
		additionalInfo := fmt.Sprintf(
			"In der Abgabe fehlen %d Schriftgutobjekte:\n    %v",
			len(missingRecordObjects),
			strings.Join(missingRecordObjects, "\n    "))
		return db.ProcessingError{
			Process:        &process,
			Description:    errorMessage,
			AdditionalInfo: additionalInfo,
			Message:        &message0503,
		}
	}

	// Check for surplus objects in the 0503 message
	var surplusRecordObjects []string
	for id, o := range includedRecordObjects {
		a := appraisals[id]
		if a.Decision != db.AppraisalDecisionA {
			if _, isFile := o.(*db.FileRecordObject); isFile {
				surplusRecordObjects = append(surplusRecordObjects, fmt.Sprintf("Akte [%s]", id.String()))
			} else {
				surplusRecordObjects = append(surplusRecordObjects, fmt.Sprintf("Vorgang [%s]", id.String()))
			}
		}
	}
	if len(surplusRecordObjects) > 0 {
		errorMessage := "Die Abgabe enthält zusätzliche Schriftgutobjekte"
		additionalInfo := fmt.Sprintf(
			"Die Abgabe enthält %d Schriftgutobjekte, die nicht als zu archivieren bewertet wurden:\n    %v",
			len(surplusRecordObjects),
			strings.Join(surplusRecordObjects, "\n    "))
		return db.ProcessingError{
			Process:        &process,
			Description:    errorMessage,
			AdditionalInfo: additionalInfo,
			Message:        &message0503,
		}
	}

	return nil
}

// compareAgencyFields checks whether the message's metadata match the agency
// and creates a processing error if not.
//
// Only values that are set in `agency` are checked.
func compareAgencyFields(agency db.Agency, message db.Message, process db.Process) error {
	a := message.MessageHead.Sender.AgencyIdentification
	if a == nil ||
		(agency.Prefix != "" && a.Prefix == nil) ||
		(agency.Code != "" && a.Code == nil) ||
		(a.Prefix != nil && agency.Prefix != *a.Prefix) ||
		(a.Code != nil && agency.Code != *a.Code) {
		info := ""
		if a != nil && a.Prefix != nil {
			info += fmt.Sprintf("Präfix der Nachricht: %s\n", *a.Prefix)
		} else {
			info += fmt.Sprintf("Präfix der Nachricht: (kein Wert)\n")
		}
		if a != nil && a.Code != nil {
			info += fmt.Sprintf("Behördenschlüssel der Nachricht: %s\n\n", *a.Code)
		} else {
			info += fmt.Sprintf("Behördenschlüssel der Nachricht: (kein Wert)\n\n")
		}
		if agency.Prefix != "" {
			info += fmt.Sprintf("Präfix der konfigurierten abgebenden Stelle: %s\n", agency.Prefix)
		} else {
			info += fmt.Sprintf("Präfix der konfigurierten abgebenden Stelle: (kein Wert)\n")
		}
		if agency.Code != "" {
			info += fmt.Sprintf("Behördenschlüssel der konfigurierten abgebenden Stelle: %s", agency.Code)
		} else {
			info += fmt.Sprintf("Behördenschlüssel der konfigurierten abgebenden Stelle: (kein Wert)")
		}
		return db.ProcessingError{
			Process:        &process,
			Message:        &message,
			Type:           db.ProcessingErrorAgencyMismatch,
			Description:    "Behördenkennung der Nachricht stimmt nicht mit der konfigurierten abgebenden Stelle überein",
			AdditionalInfo: info,
		}
	}
	return nil
}

// checkMaxRecordObjectDepth checks if the configured maximal depth of record objects in the message
// comply with the configuration. The xdomea specification allows a maximal depth of 5.
func checkMaxRecordObjectDepth(agency db.Agency, process db.Process, message db.Message) error {
	maxDepthConfig := os.Getenv("XDOMEA_MAX_RECORD_OBJECT_DEPTH")
	// This configuration does not need to be set.
	if maxDepthConfig != "" {
		maxDepth, err := strconv.Atoi(maxDepthConfig)
		// This function is not the correct place to check configuration validity.
		if err == nil {
			if maxDepth < int(message.MaxRecordObjectDepth) {
				additionalInfo := "konfigurierte maximale Stufigkeit: " +
					maxDepthConfig +
					"\nmaximale Stufigkeit in der Nachricht: " +
					strconv.FormatUint(uint64(message.MaxRecordObjectDepth), 10)
				return db.ProcessingError{
					Process:        &process,
					Message:        &message,
					Description:    "Stufigkeit zu hoch",
					AdditionalInfo: additionalInfo,
				}
			}
		}
	}
	return nil
}

func Send0502Message(agency db.Agency, message db.Message) string {
	messageXml := Generate0502Message(message)
	return sendMessage(
		agency,
		message.MessageHead.ProcessID,
		messageXml,
		Message0502MessageSuffix,
	)
}

func Send0504Message(agency db.Agency, message db.Message) string {
	messageXml := Generate0504Message(message)
	return sendMessage(
		agency,
		message.MessageHead.ProcessID,
		messageXml,
		Message0504MessageSuffix,
	)
}

func Send0506Message(process db.Process, message db.Message) {
	archivePackages := db.GetArchivePackagesWithAssociations(process.ID)
	messageXml := Generate0506Message(message, archivePackages)
	messagePath := sendMessage(
		process.Agency,
		message.MessageHead.ProcessID,
		messageXml,
		Message0506MessageSuffix,
	)
	db.UpdateProcess(process.ID, db.Process{Message0506Path: &messagePath})
}

// sendMessage creates a xdomea message and copies it into the transfer directory.
// Returns the location of the message in the transfer directory.
func sendMessage(
	agency db.Agency,
	processID string,
	messageXml string,
	messageSuffix string,
) string {
	// Create temporary directory. The name of the directory ist the message ID.
	tempDir, err := os.MkdirTemp("", processID)
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)
	xmlName := processID + messageSuffix + ".xml"
	messageName := processID + messageSuffix + ".zip"
	messagePath := path.Join(tempDir, messageName)
	messageArchive, err := os.Create(messagePath)
	if err != nil {
		panic(err)
	}
	defer messageArchive.Close()
	zipWriter := zip.NewWriter(messageArchive)
	defer zipWriter.Close()
	zipEntry, err := zipWriter.Create(xmlName)
	if err != nil {
		panic(err)
	}
	xmlStringReader := strings.NewReader(messageXml)
	_, err = io.Copy(zipEntry, xmlStringReader)
	if err != nil {
		panic(err)
	}
	// important close zip writer and message archive so it can be written on disk
	zipWriter.Close()
	messageArchive.Close()
	return CopyMessageToTransferDirectory(agency, messagePath)
}
