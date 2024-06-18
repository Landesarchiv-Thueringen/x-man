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

	"github.com/google/uuid"
	"github.com/lestrrat-go/libxml2/xsd"
)

func ProcessNewMessage(agency db.Agency, transferDirMessagePath string) {
	log.Println("Processing new message " + transferDirMessagePath)
	errorData := db.ProcessingError{
		Agency:       &agency,
		TransferPath: transferDirMessagePath,
	}
	// extract process ID from message filename
	processID, err := getProcessID(transferDirMessagePath)
	if err != nil {
		panic(err)
	}
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
	process, message, err := addMessage(
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
	if messageType == "0503" {
		// get primary documents
		rootRecords := db.FindRootRecords(context.Background(), process.ProcessID, messageType)
		primaryDocuments := db.GetPrimaryDocuments(&rootRecords)
		err = collectPrimaryDocumentsData(process, message, primaryDocuments)
		if err != nil {
			errors.AddProcessingErrorWithData(err, errorData)
		}
		err = matchAgainst0501Message(process, message)
		if err != nil {
			errors.AddProcessingErrorWithData(err, errorData)
		}
		// start format verification
		if os.Getenv("BORG_URL") != "" {
			verification.VerifyFileFormats(process, message)
		}
	}
	// if no error occurred while processing the message
	if err == nil {
		// send the confirmation message that the 0501 message was received
		if messageType == "0501" {
			messagePath := Send0504Message(agency, message)
			db.MustUpdateProcessMessagePath(process.ProcessID, db.MessageType0504, messagePath)
		}
		// send e-mail notification to users
		for _, user := range agency.Users {
			address, err := auth.GetMailAddress(user)
			if err != nil {
				errors.AddProcessingErrorWithData(err, db.ProcessingError{
					Title:     "Fehler beim Versenden einer E-Mail-Benachrichtigung",
					ProcessID: processID,
					Agency:    &agency,
				})
			} else {
				preferences := db.FindUserPreferencesWithDefault(context.Background(), user)
				if preferences.MessageEmailNotifications {
					mail.SendMailNewMessage(address, agency.Name, message)
				}
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
	for _, f := range archive.File {
		fileInArchive, err := f.Open()
		if err != nil {
			panic(err)
		}
		defer fileInArchive.Close()
		fileStorePath := path.Join(messageStoreDir, f.Name)
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
) (db.SubmissionProcess, db.Message, error) {
	messageName := getMessageName(processID, messageType)
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
		e := errors.FromError("Schema-Validierung ungültig", err)
		return db.SubmissionProcess{}, db.Message{}, &e
	}
	parsedMessage, err := parseMessage(messagePath, messageType)
	if err != nil {
		e := errors.FromError("Fehler beim Einlesen der Nachricht", err)
		return db.SubmissionProcess{}, db.Message{}, &e
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
func checkMessageValidity(agency db.Agency, messageType db.MessageType, storagePaths db.StoragePaths) error {
	xdomeaVersion, err := extractXdomeaVersion(messageType, storagePaths.MessagePath)
	if err != nil {
		return err
	}
	err = validateXdomeaXmlFile(storagePaths.MessagePath, xdomeaVersion)
	if err != nil {
		return err
	}
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
		return &db.ProcessingError{
			Title: "Primärdateien fehlen in Abgabe",
			Info:  strings.Join(missingDocuments, "\n  "),
		}
	}
	return nil
}

// matchAgainst0501Message compares a 0503 message with a previously received
// 0501 message and creates processing errors on mismatches.
func matchAgainst0501Message(
	process db.SubmissionProcess,
	message db.Message,
) error {
	// Check if 0501 message exists.
	_, found := db.FindMessage(context.Background(), process.ProcessID, db.MessageType0501)
	// 0501 Message doesn't exist. No further message validation necessary.
	if !found {
		return nil
	}
	// Check if appraisal of 0501 message is already complete.
	if !process.ProcessState.Appraisal.Complete {
		return &db.ProcessingError{
			Title: "Die Abgabe wurde erhalten, bevor die Bewertung der Anbietung abgeschlossen wurde",
		}
	}
	return checkRecordsOfMessage0503(process, message)
}

// checkRecordsOfMessage0503 compares a 0503 message with the appraisal of
// a 0501 message and returns an error if there are record objects missing in
// the 0503 message that were marked as to be archived in the appraisal or if
// there are any surplus objects included in the 0503 message.
func checkRecordsOfMessage0503(
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
		info := fmt.Sprintf(
			"In der Abgabe fehlen %d Schriftgutobjekte:\n    %v",
			len(missingRecordObjects),
			strings.Join(missingRecordObjects, "\n    "))
		return &db.ProcessingError{
			Title: "Die Abgabe ist nicht vollständig",
			Info:  info,
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
		info := fmt.Sprintf(
			"Die Abgabe enthält %d Schriftgutobjekte, die nicht als zu archivieren bewertet wurden:\n    %v",
			len(surplusRecordObjects),
			strings.Join(surplusRecordObjects, "\n    "))
		return &db.ProcessingError{
			Title: "Die Abgabe enthält zusätzliche Schriftgutobjekte",
			Info:  info,
		}
	}
	return nil
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
	maxDepthConfig := os.Getenv("XDOMEA_MAX_RECORD_OBJECT_DEPTH")
	// This configuration does not need to be set.
	if maxDepthConfig == "" {
		return nil
	}
	maxDepth, err := strconv.Atoi(maxDepthConfig)
	if err != nil {
		panic("failed to read XDOMEA_MAX_RECORD_OBJECT_DEPTH")
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
	db.MustUpdateProcessMessagePath(process.ProcessID, db.MessageType0506, messagePath)
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
