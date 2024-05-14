package xdomea

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"lath/xman/internal/auth"
	"lath/xman/internal/db"
	"lath/xman/internal/mail"
	"log"
	"os"
	"path"
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
			TransferPath:   transferDirMessagePath,
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
		// error while parsing message, can't be further processed
		HandleError(err)
		return
	}
	err = compareAgencyFields(agency, message, process)
	if err != nil {
		HandleError(err)
	}
	err = checkMaxRecordObjectDepth(agency, process, message)
	if err != nil {
		HandleError(err)
	}
	if messageType == "0503" {
		// get primary documents
		rootRecords := db.FindRootRecords(context.Background(), process.ProcessID, db.MessageType0503)
		primaryDocuments := GetPrimaryDocuments(&rootRecords)
		err = collectPrimaryDocumentsData(process, message, primaryDocuments)
		if err != nil {
			HandleError(err)
		}
		err = checkMessage0503Integrity(process, message)
		if err != nil {
			HandleError(err)
		}
		// start format verification
		if os.Getenv("BORG_ENDPOINT") != "" {
			err = VerifyFileFormats(process, message)
			if err != nil {
				HandleError(err)
			}
		}
	}
	// if no error occurred while processing the message
	if err == nil {
		// send the confirmation message that the 0501 message was received
		if messageType == "0501" {
			messagePath := Send0504Message(agency, message)
			db.UpdateProcessMessagePath(process.ProcessID, db.MessageType0504, messagePath)
		}
		// send e-mail notification to users
		for _, user := range agency.Users {
			address := auth.GetMailAddress(user)
			preferences := db.FindUserPreferences(context.Background(), user)
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
	processID uuid.UUID,
	messageType db.MessageType,
	processStoreDir string,
	storeDir string,
	transferDir string,
) (db.SubmissionProcess, db.Message, error) {
	messageName := GetMessageName(processID, messageType)
	messagePath := path.Join(storeDir, messageName)
	_, err := os.Stat(messagePath)
	if err != nil {
		panic("message doesn't exist: " + messagePath)
	}
	storagePaths := db.StoragePaths{
		TransferDirPath: transferDir,
		StoreDir:        storeDir,
		MessagePath:     messagePath,
	}
	err = checkMessageValidity(agency, messageType, storagePaths)
	if err != nil {
		return db.SubmissionProcess{}, db.Message{}, err
	}
	parsedMessage, err := parseMessage(messagePath, messageType)
	if err != nil {
		return db.SubmissionProcess{}, db.Message{}, err
	}
	message := db.Message{
		StoragePaths:   storagePaths,
		MessageType:    messageType,
		MessageHead:    parsedMessage.MessageHead,
		XdomeaVersion:  parsedMessage.XdomeaVersion.Code,
		MaxRecordDepth: getMaxRecordDepth(parsedMessage.RootRecords),
	}
	process := db.FindOrInsertProcess(processID, agency, processStoreDir)
	db.InsertMessage(message)
	markMessageReceived(message, process)
	if parsedMessage.RootRecords != nil {
		db.InsertRootRecords(processID, messageType, *parsedMessage.RootRecords)
	}
	return process, message, nil
}

// checkMessageValidity performs a xsd schema validation against the message XML file.
func checkMessageValidity(agency db.Agency, messageType db.MessageType, storagePaths db.StoragePaths) (err error) {
	xdomeaVersion, err := ExtractXdomeaVersion(messageType, storagePaths.MessagePath)
	if err != nil {
		return err
	}
	err = ValidateXdomeaXmlFile(storagePaths.MessagePath, xdomeaVersion)
	if err == nil {
		return
	}
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
			TransferPath:   storagePaths.MessagePath,
			Description:    "Schema-Validierung ungültig",
			AdditionalInfo: additionalInfo,
		}
	} else {
		return db.ProcessingError{
			Agency:       &agency,
			TransferPath: storagePaths.MessagePath,
			Description:  "Schema-Validierung ungültig",
		}
	}
}

// collectPrimaryDocumentsData checks the files referenced by the given primary
// documents and inserts corresponding primary-document-data entries into the
// database. If files are missing, it returns a processing error.
func collectPrimaryDocumentsData(
	process db.SubmissionProcess,
	message db.Message,
	primaryDocuments []db.PrimaryDocument,
) error {
	var missingDocuments []string
	var primaryDocumentsData []db.PrimaryDocumentData
	for _, d := range primaryDocuments {
		filePath := path.Join(message.StoreDir, d.Filename)
		s, err := os.Stat(filePath)
		if err != nil {
			missingDocuments = append(missingDocuments, d.Filename)
		} else {
			primaryDocumentsData = append(primaryDocumentsData, db.PrimaryDocumentData{
				ProcessID:       process.ProcessID,
				PrimaryDocument: d,
				FileSize:        s.Size(),
			})
		}
	}
	db.InsertPrimaryDocumentsData(primaryDocumentsData)
	if len(missingDocuments) > 0 {
		return db.ProcessingError{
			ProcessID:   process.ProcessID,
			Description: fmt.Sprintf("Primärdateien fehlen in Abgabe:\n  %v", strings.Join(missingDocuments, "\n  ")),
			MessageType: message.MessageType,
		}
	}
	return nil
}

func checkMessage0503Integrity(
	process db.SubmissionProcess,
	message db.Message,
) error {
	// check if 0501 message exists
	_, found := db.FindMessage(context.Background(), process.ProcessID, db.MessageType0501)
	// 0501 Message doesn't exist. No further message validation necessary.
	if !found {
		return nil
	}
	// check if appraisal of 0501 message is already complete
	if !process.ProcessState.Appraisal.Complete {
		errorMessage := "Die Abgabe wurde erhalten, bevor die Bewertung der Anbietung abgeschlossen wurde"
		return db.ProcessingError{
			ProcessID:   process.ProcessID,
			Description: errorMessage,
			MessageType: message.MessageType,
		}
	} else {
		return checkRecordObjectsOfMessage0503(process, message)
	}
}

// checkRecordObjectsOfMessage0503 compares a 0503 message with the appraisal of
// a 0501 message and returns an error if there are record objects missing in
// the 0503 message that were marked as to be archived in the appraisal or if
// there are any surplus objects included in the 0503 message.
func checkRecordObjectsOfMessage0503(
	process db.SubmissionProcess,
	message0503 db.Message,
) error {
	// Gather data
	appraisals := make(map[uuid.UUID]db.Appraisal)
	for _, a := range db.FindAppraisalsForProcess(context.Background(), process.ProcessID) {
		appraisals[a.RecordID] = a
	}

	rootRecords := db.FindRootRecords(context.Background(), process.ProcessID, db.MessageType0503)
	records := db.ExtractNestedRecords(&rootRecords)
	includedAppraisableRecords := make(map[uuid.UUID]interface{})
	for _, f := range records.Files {
		includedAppraisableRecords[f.RecordID] = &f
	}
	for _, p := range records.Processes {
		includedAppraisableRecords[p.RecordID] = &p
	}

	// Check for objects missing from the 0503 message
	var missingRecordObjects []string
	for id, a := range appraisals {
		if a.Decision == db.AppraisalDecisionA && includedAppraisableRecords[id] == nil {
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
			ProcessID:      process.ProcessID,
			Description:    errorMessage,
			AdditionalInfo: additionalInfo,
			MessageType:    db.MessageType0503,
		}
	}

	// Check for surplus objects in the 0503 message
	var surplusRecordObjects []string
	for id, o := range includedAppraisableRecords {
		a := appraisals[id]
		if a.Decision != db.AppraisalDecisionA {
			if _, isFile := o.(*db.FileRecord); isFile {
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
			ProcessID:      process.ProcessID,
			Description:    errorMessage,
			AdditionalInfo: additionalInfo,
			MessageType:    db.MessageType0503,
		}
	}

	return nil
}

// compareAgencyFields checks whether the message's metadata match the agency
// and creates a processing error if not.
//
// Only values that are set in `agency` are checked.
func compareAgencyFields(agency db.Agency, message db.Message, process db.SubmissionProcess) error {
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
			ProcessID:      process.ProcessID,
			MessageType:    message.MessageType,
			Type:           db.ProcessingErrorAgencyMismatch,
			Description:    "Behördenkennung der Nachricht stimmt nicht mit der konfigurierten abgebenden Stelle überein",
			AdditionalInfo: info,
		}
	}
	return nil
}

// checkMaxRecordObjectDepth checks if the configured maximal depth of record objects in the message
// comply with the configuration. The xdomea specification allows a maximal depth of 5.
func checkMaxRecordObjectDepth(agency db.Agency, process db.SubmissionProcess, message db.Message) error {
	maxDepthConfig := os.Getenv("XDOMEA_MAX_RECORD_OBJECT_DEPTH")
	// This configuration does not need to be set.
	if maxDepthConfig != "" {
		maxDepth, err := strconv.Atoi(maxDepthConfig)
		// This function is not the correct place to check configuration validity.
		if err == nil {
			if maxDepth < int(message.MaxRecordDepth) {
				additionalInfo := "konfigurierte maximale Stufigkeit: " +
					maxDepthConfig +
					"\nmaximale Stufigkeit in der Nachricht: " +
					strconv.FormatUint(uint64(message.MaxRecordDepth), 10)
				return db.ProcessingError{
					ProcessID:      process.ProcessID,
					MessageType:    message.MessageType,
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

func Send0506Message(process db.SubmissionProcess, message db.Message) {
	archivePackages := db.FindArchivePackagesForProcess(context.Background(), process.ProcessID)
	messageXml := Generate0506Message(message, archivePackages)
	messagePath := sendMessage(
		process.Agency,
		message.MessageHead.ProcessID,
		messageXml,
		Message0506MessageSuffix,
	)
	db.UpdateProcessMessagePath(process.ProcessID, db.MessageType0506, messagePath)
}

// sendMessage creates a xdomea message and copies it into the transfer directory.
// Returns the location of the message in the transfer directory.
func sendMessage(
	agency db.Agency,
	processID uuid.UUID,
	messageXml string,
	messageSuffix string,
) string {
	// Create temporary directory. The name of the directory ist the message ID.
	tempDir, err := os.MkdirTemp("", processID.String())
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)
	xmlName := processID.String() + messageSuffix + ".xml"
	messageName := processID.String() + messageSuffix + ".zip"
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

// getMaxRecordDepth returns the nesting level of the deepest nesting within the
// given root records.
func getMaxRecordDepth(rootRecords *db.RootRecords) uint {
	if rootRecords == nil {
		return 0
	}
	var getMaxDepthForDocuments func(documents []db.DocumentRecord) uint
	getMaxDepthForDocuments = func(documents []db.DocumentRecord) uint {
		if len(documents) == 0 {
			return 0
		}
		var depth uint = 1
		for _, d := range documents {
			depth = max(depth, 1+getMaxDepthForDocuments(d.Attachments))
		}
		return depth
	}
	var getMaxDepthForProcesses func(Processes []db.ProcessRecord) uint
	getMaxDepthForProcesses = func(processes []db.ProcessRecord) uint {
		if len(processes) == 0 {
			return 0
		}
		var depth uint = 1
		for _, p := range processes {
			depth = max(depth, 1+getMaxDepthForProcesses(p.Subprocesses))
			depth = max(depth, 1+getMaxDepthForDocuments(p.Documents))
		}
		return depth
	}
	var getMaxDepthForFiles func(Files []db.FileRecord) uint
	getMaxDepthForFiles = func(files []db.FileRecord) uint {
		if len(files) == 0 {
			return 0
		}
		var depth uint = 1
		for _, p := range files {
			depth = max(depth, 1+getMaxDepthForFiles(p.Subfiles))
			depth = max(depth, 1+getMaxDepthForProcesses(p.Processes))
			depth = max(depth, 1+getMaxDepthForDocuments(p.Documents))
		}
		return depth
	}
	return max(
		getMaxDepthForFiles(rootRecords.Files),
		getMaxDepthForProcesses(rootRecords.Processes),
		getMaxDepthForDocuments(rootRecords.Documents),
	)
}

func markMessageReceived(
	message db.Message,
	process db.SubmissionProcess,
) {
	var processStepType db.ProcessStepType
	var processStep db.ProcessStep
	switch message.MessageType {
	case "0501":
		processStepType = db.ProcessStepReceive0501
		processStep = process.ProcessState.Receive0501
	case "0503":
		processStepType = db.ProcessStepReceive0503
		processStep = process.ProcessState.Receive0503
	case "0505":
		processStepType = db.ProcessStepReceive0505
		processStep = process.ProcessState.Receive0505
	default:
		panic("unhandled message type: " + message.MessageType)
	}
	// Check if the process has already a message with the type of the given message.
	if processStep.Complete {
		panic("process already has message with type " + message.MessageType)
	}
	db.UpdateProcessStepCompletion(process.ProcessID, processStepType, true, "")
}
