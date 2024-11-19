package xdomea

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"lath/xman/internal/auth"
	"lath/xman/internal/db"
	"lath/xman/internal/errors"
	"lath/xman/internal/mail"
	"lath/xman/internal/verification"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/lestrrat-go/libxml2/xsd"
	"golang.org/x/text/encoding/charmap"
)

var transferFileExists = fmt.Errorf("transfer file exists")

func ProcessNewMessage(agency db.Agency, transferDirMessagePath string) {
	log.Println("Processing new message " + transferDirMessagePath)
	errorData := db.ProcessingError{
		Agency:       &agency,
		TransferPath: transferDirMessagePath,
	}
	// extract process ID from message filename
	processID := getProcessID(transferDirMessagePath)
	// extract message type from message filename
	messageType, err := getMessageTypeImpliedByPath(transferDirMessagePath)
	if err != nil {
		panic(err)
	}
	// copy message from transfer directory to a local temporary directory
	localMessagePath := CopyMessageFromTransferDirectory(agency, transferDirMessagePath)
	defer os.Remove(localMessagePath)
	// extract message to message storage
	processStoreDir, messageStoreDir := extractMessageToMessageStore(
		agency,
		transferDirMessagePath,
		localMessagePath,
		processID,
		messageType,
	)
	// save message
	process, message, rootRecords, err := addMessage(
		agency,
		processID,
		messageType,
		processStoreDir,
		messageStoreDir,
		transferDirMessagePath,
	)
	if err != nil {
		errors.AddProcessingErrorWithData(err, errorData)
		return
	}
	errorData.ProcessID = processID
	errorData.MessageType = messageType
	err = compareAgencyFields(agency, message, process)
	if err != nil {
		errors.AddProcessingErrorWithData(err, errorData)
	}
	err = checkMaxRecordObjectDepth(agency, process, message)
	if err != nil {
		errors.AddProcessingErrorWithData(err, errorData)
	}
	if rootRecords != nil && len(rootRecords.Documents) > 0 &&
		(messageType == db.MessageType0501 ||
			messageType == db.MessageType0503 && !process.ProcessState.Receive0501.Complete) {
		db.InsertWarning(db.Warning{
			CreatedAt:   time.Now(),
			Title:       "Aussonderung enthält nicht zugeordnete Dokumente",
			MessageType: messageType,
			ProcessID:   processID,
		})
	}
	if messageType == "0503" {
		primaryDocuments := GetPrimaryDocuments(rootRecords)
		primaryDocumentsData, err := collectPrimaryDocumentsData(
			process, message, primaryDocuments,
		)
		if err != nil {
			errors.AddProcessingErrorWithData(err, errorData)
		}
		errs := matchAgainst0501Message(process, message)
		for _, err := range errs {
			errors.AddProcessingErrorWithData(err, errorData)
		}
		// start format verification
		if os.Getenv("BORG_URL") != "" {
			verification.VerifyFileFormats(process, primaryDocumentsData)
		}
	}
	confirmMessageReceipt(agency, processID, messageType)
}

// confirmMessageReceipt sends the appropriate xdomea message (if any) and an
// e-mail notification to the post office and the archivist(s) in charge.
func confirmMessageReceipt(agency db.Agency, processID uuid.UUID, messageType db.MessageType) {
	message, ok := db.FindMessage(context.Background(), processID, messageType)
	if !ok {
		panic(fmt.Sprintf("failed to find message %s for process %s", messageType, processID))
	}
	errorData := db.ProcessingError{
		Agency:      &agency,
		ProcessID:   processID,
		MessageType: messageType,
	}
	// Send the confirmation message that the 0501 message was received.
	if messageType == db.MessageType0501 {
		err := Send0504Message(agency, message)
		if err != nil {
			if err == transferFileExists {
				// Ignore. This can occur when re-importing the message.
			} else {
				errorData.Title = "Fehler beim Senden der 0504-Nachricht"
				errors.AddProcessingErrorWithData(err, errorData)
			}
		}
	} else if messageType == db.MessageType0503 {
		err := Send0507Message(agency, message)
		if err != nil {
			if err == transferFileExists {
				// Ignore. This can occur when re-importing the message.
			} else {
				errorData.Title = "Fehler beim Senden der 0507-Nachricht"
				errors.AddProcessingErrorWithData(err, errorData)
			}
		}
	}
	// Forward e-mail to post office.
	errorData.Title = "Fehler bei der E-Mail-Weiterleitung zur Poststelle"
	if address := os.Getenv("POST_OFFICE_EMAIL"); address != "" {
		err := mail.SendMailNewMessagePostOffice(address, agency.Name, message)
		if err != nil {
			errors.AddProcessingErrorWithData(err, errorData)
		}
	}
	// Send e-mail notification to users.
	errorData.Title = "Fehler beim Versenden einer E-Mail-Benachrichtigung"
	for _, user := range agency.Users {
		address, err := auth.GetMailAddress(user)
		if err != nil {
			errors.AddProcessingErrorWithData(err, errorData)
			continue
		}
		preferences := db.FindUserPreferencesWithDefault(context.Background(), user)
		if preferences.MessageEmailNotifications {
			err = mail.SendMailNewMessageNotification(address, agency.Name, message)
			if err != nil {
				errors.AddProcessingErrorWithData(err, errorData)
			}
		}
	}
}

// extractMessage parses the given message file into a database entry and saves
// it to the database. It returns the saved entry.
//
// Returns the directories in message store for the process and the message.
func extractMessageToMessageStore(
	agency db.Agency,
	transferDirMessagePath string,
	localMessagePath string,
	processID uuid.UUID,
	messageType db.MessageType,
) (processStoreDir string, messageStoreDir string) {
	processStoreDir = path.Join("message_store", processID.String())
	// Create the message store directory if necessary.
	messageStoreDir = path.Join(processStoreDir, string(messageType))
	err := os.MkdirAll(messageStoreDir, 0700)
	if err != nil {
		panic(err)
	}
	// Open the message archive (zip).
	archive, err := zip.OpenReader(localMessagePath)
	if err != nil {
		panic(err)
	}
	defer archive.Close()
	var invalidFilenames []string
	for _, f := range archive.File {
		fileName := f.Name
		// file name is nor UTF-8 encoded
		if f.NonUTF8 {
			// try reading file name as IBM code page 437
			fileName, err = charmap.CodePage437.NewDecoder().String(f.Name)
			// file name could not be encoded as UTF-8
			if err != nil || !utf8.ValidString(fileName) {
				invalidFilenames = append(invalidFilenames, fileName)
			}
		}
		fileInArchive, err := f.Open()
		if err != nil {
			panic(err)
		}
		defer fileInArchive.Close()
		fileStorePath := path.Join(messageStoreDir, fileName)
		fileInStore, err := os.Create(fileStorePath)
		if err != nil {
			panic(err)
		}
		defer fileInStore.Close()
		_, err = io.Copy(fileInStore, fileInArchive)
		if err != nil {
			panic(err)
		}
	}
	// create a processing error for all invalid encoded file names
	if len(invalidFilenames) > 0 {
		errorInfo := "Die Dateinamen in der ZIP-Datei konnten nicht dekodiert werden.\n"
		errorInfo += "Der ZIP-Standard erlaubt nur UTF-8 und IBM Code Page 437 als Textkodierungen für Dateinamen.\n\n"
		errorInfo += "Fehlerhaft kodierte Dateinamen:\n"
		errorInfo += strings.Join(invalidFilenames, ",\n")
		procErr := db.ProcessingError{
			Title:        "ungültige Zeichenkodierung",
			Agency:       &agency,
			TransferPath: transferDirMessagePath,
			Info:         errorInfo,
		}
		errors.AddProcessingError(procErr)
	}
	return
}

// addMessage parses the message and saves it in the database.
//
// It returns panics with a processing error if reading the message failed due
// to errors within the message file.
func addMessage(
	agency db.Agency,
	processID uuid.UUID,
	messageType db.MessageType,
	processStoreDir string,
	storeDir string,
	transferDir string,
) (db.SubmissionProcess, db.Message, *db.RootRecords, error) {
	messageName := getMessageName(processID, messageType)
	messagePath := path.Join(storeDir, messageName)
	_, err := os.Stat(messagePath)
	if err != nil {
		panic("message doesn't exist: " + messagePath)
	}
	storagePaths := db.StoragePaths{
		TransferFile: transferDir,
		StoreDir:     storeDir,
		MessagePath:  messagePath,
	}
	err = checkMessageValidity(agency, messageType, storagePaths)
	if err != nil {
		e := errors.FromError("Schema-Validierung ungültig", err)
		return db.SubmissionProcess{}, db.Message{}, nil, &e
	}
	parsedMessage, err := parseMessage(messagePath, messageType)
	if err != nil {
		e := errors.FromError("Fehler beim Einlesen der Nachricht", err)
		return db.SubmissionProcess{}, db.Message{}, nil, &e
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
	return process, message, parsedMessage.RootRecords, nil
}

// checkMessageValidity performs a xsd schema validation against the message XML file.
func checkMessageValidity(agency db.Agency, messageType db.MessageType, storagePaths db.StoragePaths) error {
	xdomeaVersion, err := extractXdomeaVersion(messageType, storagePaths.MessagePath)
	if err != nil {
		return err
	}
	err = validateXdomeaXmlFile(storagePaths.MessagePath, xdomeaVersion)
	validationError, ok := err.(xsd.SchemaValidationError)
	if ok {
		var errorMessages []string
		for _, e := range validationError.Errors() {
			errorMessages = append(errorMessages, e.Error())
		}
		return fmt.Errorf("%s", strings.Join(errorMessages, "\n"))
	} else {
		return err
	}
}

// collectPrimaryDocumentsData checks the files referenced by the given primary
// documents and inserts corresponding primary-document-data entries into the
// database. If files are missing, it returns a processing error.
func collectPrimaryDocumentsData(
	process db.SubmissionProcess,
	message db.Message,
	primaryDocuments []db.PrimaryDocumentContext,
) ([]db.PrimaryDocumentData, error) {
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
				RecordID:        d.RecordID,
				PrimaryDocument: d.PrimaryDocument,
				FileSize:        s.Size(),
			})
		}
	}
	if len(primaryDocumentsData) > 0 {
		db.InsertPrimaryDocumentsData(primaryDocumentsData)
	}
	if len(missingDocuments) > 0 {
		return primaryDocumentsData, &db.ProcessingError{
			Title: "Primärdateien fehlen in Abgabe",
			Info:  strings.Join(missingDocuments, "\n  "),
		}
	}
	return primaryDocumentsData, nil
}

// matchAgainst0501Message compares a 0503 message with a previously received
// 0501 message and creates processing errors on mismatches.
func matchAgainst0501Message(
	process db.SubmissionProcess,
	message0503 db.Message,
) []error {
	// Check if 0501 message exists.
	message0501, found := db.FindMessage(context.Background(), process.ProcessID, db.MessageType0501)
	// 0501 Message doesn't exist. No further message validation necessary.
	if !found {
		return []error{}
	}
	// Check if appraisal of 0501 message is already complete.
	if !process.ProcessState.Appraisal.Complete {
		return []error{&db.ProcessingError{
			Title: "Die Abgabe wurde erhalten, bevor die Bewertung der Anbietung abgeschlossen wurde",
		}}
	}
	return checkRecordsOfMessage0503(message0501, message0503)
}

// checkRecordsOfMessage0503 compares a 0503 message with the appraisal of a
// 0501 message and returns an error if there are records missing in the 0503
// message that were marked as to be archived in the appraisal or if there are
// any surplus records included in the 0503 message.
//
// If a record is found be be missing or surplus, its child records will not be
// listed.
func checkRecordsOfMessage0503(
	message0501,
	message0503 db.Message,
) []error {
	discrepancies := FindDiscrepancies(message0501, message0503)
	var errs []error
	if n := len(discrepancies.MissingRecords); n > 0 {
		var s string
		if n == 1 {
			s = "fehlt 1 Schriftgutobjekt"
		} else {
			s = fmt.Sprintf("fehlen %d Schriftgutobjekte", n)
		}
		info := fmt.Sprintf(
			"In der Abgabe %s:\n    %v",
			s, strings.Join(discrepancies.MissingRecords, "\n    "))
		errs = append(errs, &db.ProcessingError{
			Title: "Die Abgabe ist nicht vollständig",
			Info:  info,
		})
	}
	if n := len(discrepancies.SurplusRecords); n > 0 {
		var s string
		if n == 1 {
			s = "1 Schriftgutobjekt, das nicht als zu archivieren bewertet wurde"
		} else {
			s = fmt.Sprintf(
				"%d Schriftgutobjekte, die nicht als zu archivieren bewertet wurden",
				n,
			)
		}
		info := fmt.Sprintf(
			"Die Abgabe enthält %s:\n    %v",
			s, strings.Join(discrepancies.SurplusRecords, "\n    "),
		)
		errs = append(errs, &db.ProcessingError{
			Title: "Die Abgabe enthält zusätzliche Schriftgutobjekte",
			Info:  info,
		})
	}
	return errs
}

// compareAgencyFields checks whether the message's metadata match the agency
// and returns a processing error if not.
//
// Only values that are set in `agency` are checked.
func compareAgencyFields(agency db.Agency, message db.Message, process db.SubmissionProcess) error {
	if agency.Prefix == "" && agency.Code == "" {
		return nil
	}
	a := message.MessageHead.Sender.AgencyIdentification
	if a == nil {
		return &db.ProcessingError{
			Title: "Behördenkennung der Nachricht stimmt nicht mit der konfigurierten abgebenden Stelle überein",
			Info:  "Die Nachricht gibt keine Behördenkennung an",
		}
	}
	if (agency.Prefix == "" || agency.Prefix == a.Prefix) && (agency.Code == "" || agency.Code == a.Code) {
		return nil
	}
	info := ""
	if a.Prefix != "" {
		info += fmt.Sprintf("Präfix der Nachricht: %s\n", a.Prefix)
	} else {
		info += fmt.Sprintf("Präfix der Nachricht: (kein Wert)\n")
	}
	if a.Code != "" {
		info += fmt.Sprintf("Behördenschlüssel der Nachricht: %s\n\n", a.Code)
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
	return &db.ProcessingError{
		Title: "Behördenkennung der Nachricht stimmt nicht mit der konfigurierten abgebenden Stelle überein",
		Info:  info,
	}
}

// checkMaxRecordObjectDepth checks if the configured maximal depth of record objects in the message
// comply with the configuration. The xdomea specification allows a maximal depth of 5.
func checkMaxRecordObjectDepth(agency db.Agency, process db.SubmissionProcess, message db.Message) error {
	maxDepthConfig := os.Getenv("MAX_RECORD_DEPTH")
	// This configuration does not need to be set.
	if maxDepthConfig == "" {
		return nil
	}
	maxDepth, err := strconv.Atoi(maxDepthConfig)
	if err != nil {
		panic("invalid value for MAX_RECORD_DEPTH: " + maxDepthConfig)
	}
	if maxDepth < message.MaxRecordDepth {
		info := fmt.Sprintf(
			"konfigurierte maximale Stufigkeit: %d\n"+
				"maximale Stufigkeit in der Nachricht: %d",
			maxDepth, message.MaxRecordDepth)
		return &db.ProcessingError{
			Title: "Stufigkeit zu hoch",
			Info:  info,
		}
	}
	return nil
}

func Send0502Message(agency db.Agency, message db.Message) error {
	messageXml := Generate0502Message(message)
	return sendMessage(
		agency,
		message.MessageHead.ProcessID,
		messageXml,
		Message0502MessageSuffix,
	)
}

func Send0504Message(agency db.Agency, message db.Message) error {
	messageXml := Generate0504Message(message)
	return sendMessage(
		agency,
		message.MessageHead.ProcessID,
		messageXml,
		Message0504MessageSuffix,
	)
}

func Send0506Message(process db.SubmissionProcess, message0503 db.Message) error {
	archivePackages := db.FindArchivePackagesForProcess(context.Background(), process.ProcessID)
	messageXml := Generate0506Message(message0503, archivePackages)
	return sendMessage(
		process.Agency,
		message0503.MessageHead.ProcessID,
		messageXml,
		Message0506MessageSuffix,
	)
}

func Send0507Message(agency db.Agency, message0503 db.Message) error {
	messageXml, ok := Generate0507Message(message0503)
	if !ok {
		return nil
	}
	return sendMessage(
		agency,
		message0503.MessageHead.ProcessID,
		messageXml,
		Message0507MessageSuffix,
	)
}

// sendMessage creates a xdomea message and copies it into the transfer directory.
// Returns the location of the message in the transfer directory.
func sendMessage(
	agency db.Agency,
	processID uuid.UUID,
	messageXml string,
	messageSuffix string,
) error {
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
	return CopyMessageToTransferDirectory(agency, processID, messagePath)
}

// getMaxRecordDepth returns the nesting level of the deepest nesting within the
// given root records.
func getMaxRecordDepth(rootRecords *db.RootRecords) int {
	if rootRecords == nil {
		return 0
	}
	var getMaxDepthForDocuments func(documents []db.DocumentRecord) int
	getMaxDepthForDocuments = func(documents []db.DocumentRecord) int {
		if len(documents) == 0 {
			return 0
		}
		var depth int = 1
		for _, d := range documents {
			depth = max(depth, 1+getMaxDepthForDocuments(d.Attachments))
		}
		return depth
	}
	var getMaxDepthForProcesses func(Processes []db.ProcessRecord) int
	getMaxDepthForProcesses = func(processes []db.ProcessRecord) int {
		if len(processes) == 0 {
			return 0
		}
		var depth int = 1
		for _, p := range processes {
			depth = max(depth, 1+getMaxDepthForProcesses(p.Subprocesses))
			depth = max(depth, 1+getMaxDepthForDocuments(p.Documents))
		}
		return depth
	}
	var getMaxDepthForFiles func(Files []db.FileRecord) int
	getMaxDepthForFiles = func(files []db.FileRecord) int {
		if len(files) == 0 {
			return 0
		}
		var depth int = 1
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
	db.MustUpdateProcessStepCompletion(process.ProcessID, processStepType, true, "")
}
