package xdomea

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"lath/xman/internal/db"
	"lath/xman/internal/format"
	"log"
	"os"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/lestrrat-go/libxml2/xsd"
)

func ProcessNewMessage(agency db.Agency, localPath string, transferDirPath string) {

}

// TODO: description
//
// It returns an error when a reading the message resulted in any processing
// error.
func AddMessage(
	agency db.Agency,
	xdomeaID string,
	messageType db.MessageType,
	processStoreDir string,
	messageStoreDir string,
	transferDirMessagePath string,
) (db.Process, db.Message, error) {
	var process db.Process
	var message db.Message
	messageName := GetMessageName(xdomeaID, messageType)
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
	// Store message in database with parsed message metadata.
	process, message, err = db.AddMessage(agency, xdomeaID, processStoreDir, message)
	if err != nil {
		return process, message, err
	}
	if err := compareAgencyFields(agency, message, process); err != nil {
		return process, message, err
	}
	if messageType.Code == "0503" {
		// get primary documents
		primaryDocuments := db.GetAllPrimaryDocuments(message.ID)
		err = checkMessage0503Integrity(process, message, primaryDocuments)
		if err != nil {
			return process, message, err
		}
		// start format verification
		err = format.VerifyFileFormats(process, message)
		if err != nil {
			return process, message, err
		}
	}
	// if no error occurred while processing the message
	if err == nil {
		// store the confirmation message that the 0501 message was received
		if messageType.Code == "0501" {
			messagePath := Send0504Message(agency, message)
			process.Message0504Path = &messagePath
			db.UpdateProcess(process)
		}
	}
	return process, message, nil
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

func sendMessage(
	agency db.Agency,
	messageID string,
	messageXml string,
	messageSuffix string,
) string {
	// Create temporary directory. The name of the directory ist the message ID.
	tempDir, err := os.MkdirTemp("", messageID)
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)
	xmlName := messageID + messageSuffix + ".xml"
	messageName := messageID + messageSuffix + ".zip"
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
	message0501, found := db.GetMessageOfProcessByCode(process, "0501")
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
		return checkRecordObjectsOfMessage0503(process, message0501, message0503)
	}
}

func checkRecordObjectsOfMessage0503(
	process db.Process,
	message0501 db.Message,
	message0503 db.Message,
) error {
	message0503Incomplete := false
	additionalInfo := ""
	err := checkFileRecordObjectsOfMessage0503(
		message0501.ID,
		message0503.ID,
		&additionalInfo,
	)
	if err != nil {
		message0503Incomplete = true
	}
	err = checkProcessRecordObjectsOfMessage0503(
		message0501.ID,
		message0503.ID,
		&additionalInfo,
	)
	if err != nil {
		message0503Incomplete = true
	}
	if message0503Incomplete {
		errorMessage := "Die Abgabe ist nicht vollständig"
		return db.ProcessingError{
			Process:        &process,
			Description:    errorMessage,
			AdditionalInfo: additionalInfo,
			Message:        &message0503,
		}
	}
	return nil
}

func checkFileRecordObjectsOfMessage0503(
	message0501ID uuid.UUID,
	message0503ID uuid.UUID,
	additionalInfo *string,
) error {
	message0503Incomplete := false
	fileIndex0501, err := db.GetAllFileRecordObjects(message0501ID)
	if err != nil {
		panic(err)
	}
	fileIndex0503, err := db.GetAllFileRecordObjects(message0503ID)
	if err != nil {
		panic(err)
	}
	for id0501, file0501 := range fileIndex0501 {
		// missing appraisal metadata for 0501 message, should not happen
		if file0501.ArchiveMetadata == nil || file0501.ArchiveMetadata.AppraisalCode == nil {
			continue
		}
		_, file0503Exists := fileIndex0503[id0501]
		if *file0501.ArchiveMetadata.AppraisalCode == "A" && !file0503Exists {
			errorMessage :=
				"0503 integrity check failed: missing file record object [" + file0501.ID.String() + "]"
			*additionalInfo += "Akte [" + file0501.ID.String() + "] fehlt in Abgabe"
			log.Println(errorMessage)
			message0503Incomplete = true
		}
	}
	if message0503Incomplete {
		return errors.New("0503 message incomplete: file record objects missing")
	}
	return nil
}

func checkProcessRecordObjectsOfMessage0503(
	message0501ID uuid.UUID,
	message0503ID uuid.UUID,
	additionalInfo *string,
) error {
	message0503Incomplete := false
	processIndex0501, err := db.GetAllProcessRecordObjects(message0501ID)
	if err != nil {
		panic(err)
	}
	processIndex0503, err := db.GetAllProcessRecordObjects(message0503ID)
	if err != nil {
		panic(err)
	}
	for id0501, process0501 := range processIndex0501 {
		// missing appraisal metadata for 0501 message, should not happen
		if process0501.ArchiveMetadata == nil || process0501.ArchiveMetadata.AppraisalCode == nil {
			continue
		}
		_, process0503Exists := processIndex0503[id0501]
		if *process0501.ArchiveMetadata.AppraisalCode == "A" && !process0503Exists {
			errorMessage := "0503 integrity check failed: missing process record object [" +
				process0501.ID.String() + "]"
			*additionalInfo += "Vorgang [" + process0501.ID.String() + "] fehlt in Abgabe"
			log.Println(errorMessage)
			message0503Incomplete = true
		}
	}
	if message0503Incomplete {
		return errors.New("0503 message incomplete: process record objects missing")
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
