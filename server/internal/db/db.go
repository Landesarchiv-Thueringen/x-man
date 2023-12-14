package db

import (
	"errors"
	"log"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func Init() {
	dsn := `host=xman-database 
		user=xman
		password=test1234
		dbname=xman 
		port=5432 
		sslmode=disable 
		TimeZone=Europe/Berlin`
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database!")
	}
	db = database
}

func MigrationCompleted() bool {
	var serverState ServerState
	result := db.First(&serverState)
	return result.Error == nil
}

func SetMigrationCompleted() {
	serverState := ServerState{
		MigrationComplete: true,
	}
	result := db.Save(&serverState)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
}

func Migrate() {
	if db == nil {
		log.Fatal("database wasn't initialized")
	}
	// Migrate the complete schema.
	db.AutoMigrate(
		&ServerState{},
		&Agency{},
		&XdomeaVersion{},
		&Code{},
		&Process{},
		&ProcessState{},
		&ProcessStep{},
		&Message{},
		&MessageType{},
		&MessageHead{},
		&Contact{},
		&AgencyIdentification{},
		&Institution{},
		&RecordObject{},
		&FileRecordObject{},
		&ProcessRecordObject{},
		&DocumentRecordObject{},
		&GeneralMetadata{},
		&FilePlan{},
		&Lifetime{},
		&ArchiveMetadata{},
		&RecordObjectAppraisal{},
		&RecordObjectConfidentiality{},
		&Version{},
		&Format{},
		&PrimaryDocument{},
		&FormatVerification{},
		&ToolResponse{},
		&ExtractedFeature{},
		&Feature{},
		&FeatureValue{},
		&ToolConfidence{},
		&ProcessingError{},
	)
}

func InitMessageTypes(messageTypes []*MessageType) {
	result := db.Create(messageTypes)
	if result.Error != nil {
		log.Fatal("Failed to initialize message types!")
	}
}

func InitXdomeaVersions(versions []*XdomeaVersion) {
	result := db.Create(versions)
	if result.Error != nil {
		log.Fatal("Failed to initialize xdomea versions!")
	}
}

func InitRecordObjectAppraisals(appraisals []*RecordObjectAppraisal) {
	result := db.Create(appraisals)
	if result.Error != nil {
		log.Fatal("Failed to initialize record object appraisal values!")
	}
}

func InitRecordObjectConfidentialities(confidentialities []*RecordObjectConfidentiality) {
	result := db.Create(confidentialities)
	if result.Error != nil {
		log.Fatal("Failed to initialize record object confidentialitiy values!")
	}
}

func InitAgencies(agencies []Agency) {
	result := db.Create(agencies)
	if result.Error != nil {
		log.Fatal("Failed to initialize agency configuration!")
	}
}

func GetProcessingErrors() []ProcessingError {
	var processingErrors []ProcessingError
	result := db.Preload("Agency").Find(&processingErrors)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return processingErrors
}

func GetAgencies() ([]Agency, error) {
	var agencies []Agency
	result := db.Find(&agencies)
	return agencies, result.Error
}

func GetSupportedXdomeaVersions() []XdomeaVersion {
	var xdomeaVersions []XdomeaVersion
	result := db.Find(&xdomeaVersions)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return xdomeaVersions
}

func GetXdomeaVersionByCode(code string) (XdomeaVersion, error) {
	xdomeaVersion := XdomeaVersion{
		Code: code,
	}
	result := db.Where(&xdomeaVersion).First(&xdomeaVersion)
	return xdomeaVersion, result.Error
}

func GetProcesses() ([]Process, error) {
	var processes []Process
	result := db.
		Preload("Agency").
		Preload("Message0501.MessageHead").
		Preload("Message0501.MessageType").
		Preload("Message0503.MessageHead").
		Preload("Message0503.MessageType").
		Preload("ProcessingErrors").
		Preload("ProcessingErrors.Agency").
		Preload("ProcessState.Receive0501").
		Preload("ProcessState.Appraisal").
		Preload("ProcessState.Receive0505").
		Preload("ProcessState.Receive0503").
		Preload("ProcessState.FormatVerification").
		Preload("ProcessState.Archiving").
		Find(&processes)
	if result.Error != nil {
		return processes, result.Error
	}
	var processesWithoutErrors []Process
	for _, p := range processes {
		if len(p.ProcessingErrors) == 0 {
			processesWithoutErrors = append(processesWithoutErrors, p)
		}
	}
	return processesWithoutErrors, result.Error
}

func GetMessageByID(id uuid.UUID) (Message, error) {
	var message Message
	result := db.First(&message, id)
	return message, result.Error
}

func GetCompleteMessageByID(id uuid.UUID) (Message, error) {
	var message Message
	result := db.
		Preload("MessageType").
		Preload("MessageHead.Sender.Institution").
		Preload("MessageHead.Sender.AgencyIdentification").
		Preload("MessageHead.Sender.AgencyIdentification.Code").
		Preload("MessageHead.Sender.AgencyIdentification.Prefix").
		Preload("MessageHead.Receiver.Institution").
		Preload("MessageHead.Receiver.AgencyIdentification.Code").
		Preload("MessageHead.Receiver.AgencyIdentification.Prefix").
		Preload("RecordObjects.FileRecordObject.GeneralMetadata.FilePlan").
		Preload("RecordObjects.FileRecordObject.ArchiveMetadata").
		Preload("RecordObjects.FileRecordObject.Lifetime").
		Preload("RecordObjects.FileRecordObject.Processes.GeneralMetadata.FilePlan").
		Preload("RecordObjects.FileRecordObject.Processes.ArchiveMetadata").
		Preload("RecordObjects.FileRecordObject.Processes.Lifetime").
		Preload("RecordObjects.FileRecordObject.Processes.Documents.GeneralMetadata.FilePlan").
		First(&message, id)
	return message, result.Error
}

func GetMessageTypeCode(id uuid.UUID) (string, error) {
	var message Message
	result := db.
		Preload("MessageType").
		First(&message, id)
	if result.Error != nil {
		return "", result.Error
	}
	return message.MessageType.Code, nil
}

func IsMessageAppraisalComplete(id uuid.UUID) (bool, error) {
	message, err := GetMessageByID(id)
	if err != nil {
		return false, err
	}
	return message.AppraisalComplete, err
}

func GetRecordObjects(messageID uuid.UUID) ([]RecordObject, error) {
	var recordObjects []RecordObject
	// TODO: add process and document
	result := db.
		Preload("FileRecordObject.GeneralMetadata.FilePlan").
		Preload("FileRecordObject.ArchiveMetadata").
		Preload("FileRecordObject.Lifetime").
		Preload("FileRecordObject.Processes.GeneralMetadata.FilePlan").
		Preload("FileRecordObject.Processes.ArchiveMetadata").
		Preload("FileRecordObject.Processes.Lifetime").
		Preload("FileRecordObject.Processes.Documents.GeneralMetadata.FilePlan").
		Find(&recordObjects)
	return recordObjects, result.Error
}

func GetFileRecordObjectByID(id uuid.UUID) (FileRecordObject, error) {
	var file FileRecordObject
	result := db.
		Preload("GeneralMetadata.FilePlan").
		Preload("ArchiveMetadata").
		Preload("Lifetime").
		Preload("Processes.GeneralMetadata.FilePlan").
		Preload("Processes.ArchiveMetadata").
		Preload("Processes.Lifetime").
		Preload("Processes.Documents.GeneralMetadata.FilePlan").
		Preload("Processes.Documents.Versions.Formats.PrimaryDocument").
		First(&file, id)
	return file, result.Error
}

func GetProcessRecordObjectByID(id uuid.UUID) (ProcessRecordObject, error) {
	var process ProcessRecordObject
	result := db.
		Preload("GeneralMetadata.FilePlan").
		Preload("ArchiveMetadata").
		Preload("Lifetime").
		Preload("Documents.GeneralMetadata.FilePlan").
		Preload("Documents.Versions.Formats.PrimaryDocument").
		First(&process, id)
	return process, result.Error
}

func GetDocumentRecordObjectByID(id uuid.UUID) (DocumentRecordObject, error) {
	var document DocumentRecordObject
	result := db.
		Preload("GeneralMetadata.FilePlan").
		Preload("Versions.Formats.PrimaryDocument").
		First(&document, id)
	return document, result.Error
}

func GetAllFileRecordObjects(messageID uuid.UUID) (map[uuid.UUID]FileRecordObject, error) {
	var fileRecordObjects []FileRecordObject
	result := db.
		Preload("GeneralMetadata.FilePlan").
		Preload("ArchiveMetadata").
		Preload("Lifetime").
		Preload("Processes.GeneralMetadata.FilePlan").
		Preload("Processes.ArchiveMetadata").
		Preload("Processes.Lifetime").
		Preload("Processes.Documents.GeneralMetadata.FilePlan").
		Preload("Processes.Documents.Versions.Formats.PrimaryDocument").
		Where("message_id = ?", messageID.String()).
		Find(&fileRecordObjects)
	fileIndex := make(map[uuid.UUID]FileRecordObject)
	for _, f := range fileRecordObjects {
		fileIndex[f.XdomeaID] = f
	}
	return fileIndex, result.Error
}

func GetAllProcessRecordObjects(messageID uuid.UUID) (map[uuid.UUID]ProcessRecordObject, error) {
	var processRecordObjects []ProcessRecordObject
	result := db.
		Preload("GeneralMetadata.FilePlan").
		Preload("ArchiveMetadata").
		Preload("Lifetime").
		Preload("Documents.GeneralMetadata.FilePlan").
		Preload("Documents.Versions.Formats.PrimaryDocument").
		Where("message_id = ?", messageID.String()).
		Find(&processRecordObjects)
	processIndex := make(map[uuid.UUID]ProcessRecordObject)
	for _, p := range processRecordObjects {
		processIndex[p.XdomeaID] = p
	}
	return processIndex, result.Error
}

func GetAllDocumentRecordObjects(messageID uuid.UUID) (map[uuid.UUID]DocumentRecordObject, error) {
	var documentRecordObjects []DocumentRecordObject
	result := db.
		Preload("GeneralMetadata.FilePlan").
		Preload("Versions.Formats.PrimaryDocument").
		Where("message_id = ?", messageID.String()).
		Find(&documentRecordObjects)
	documentIndex := make(map[uuid.UUID]DocumentRecordObject)
	for _, d := range documentRecordObjects {
		documentIndex[d.XdomeaID] = d
	}
	return documentIndex, result.Error
}

func GetAllPrimaryDocuments(messageID uuid.UUID) ([]PrimaryDocument, error) {
	var primaryDocuments []PrimaryDocument
	var documents []DocumentRecordObject
	result := db.
		Preload("Versions.Formats.PrimaryDocument").
		Where("message_id = ?", messageID.String()).
		Find(&documents)
	if result.Error != nil {
		return primaryDocuments, result.Error
	}
	for _, document := range documents {
		if document.Versions != nil {
			for _, version := range document.Versions {
				for _, format := range version.Formats {
					primaryDocuments = append(primaryDocuments, format.PrimaryDocument)
				}
			}
		}
	}
	return primaryDocuments, nil
}

func GetAllPrimaryDocumentsWithFormatVerification(messageID uuid.UUID) ([]PrimaryDocument, error) {
	var primaryDocuments []PrimaryDocument
	var documents []DocumentRecordObject
	result := db.
		Preload("Versions.Formats.PrimaryDocument.FormatVerification.Features.Values.Tools").
		Preload("Versions.Formats.PrimaryDocument.FormatVerification.FileIdentificationResults.Features").
		Preload("Versions.Formats.PrimaryDocument.FormatVerification.FileValidationResults.Features").
		Where("message_id = ?", messageID.String()).
		Find(&documents)
	if result.Error != nil {
		return primaryDocuments, result.Error
	}
	for _, document := range documents {
		if document.Versions != nil {
			for _, version := range document.Versions {
				for _, format := range version.Formats {
					primaryDocuments = append(primaryDocuments, format.PrimaryDocument)
				}
			}
		}
	}
	for primaryDocumentIndex, primaryDocument := range primaryDocuments {
		if primaryDocument.FormatVerification == nil {
			continue
		}
		if len(primaryDocument.FormatVerification.Features) > 0 {
			summary := make(map[string]Feature)
			for _, feature := range primaryDocument.FormatVerification.Features {
				summary[feature.Key] = feature
			}
			primaryDocuments[primaryDocumentIndex].FormatVerification.Summary = summary
		}
		if len(primaryDocument.FormatVerification.FileIdentificationResults) > 0 {
			for toolID, tool := range primaryDocument.FormatVerification.FileIdentificationResults {
				features := make(map[string]string)
				for _, feature := range tool.Features {
					features[feature.Key] = feature.Value
				}
				primaryDocuments[primaryDocumentIndex].FormatVerification.
					FileIdentificationResults[toolID].ExtractedFeatures = &features
			}
		}
		if len(primaryDocument.FormatVerification.FileValidationResults) > 0 {
			for toolID, tool := range primaryDocument.FormatVerification.FileValidationResults {
				features := make(map[string]string)
				for _, feature := range tool.Features {
					features[feature.Key] = feature.Value
				}
				primaryDocuments[primaryDocumentIndex].FormatVerification.
					FileValidationResults[toolID].ExtractedFeatures = &features
			}
		}
	}
	return primaryDocuments, nil
}

func GetMessageTypeByCode(code string) MessageType {
	messageType := MessageType{Code: code}
	result := db.Where(&messageType).First(&messageType)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return messageType
}

func GetRecordObjectAppraisals() ([]RecordObjectAppraisal, error) {
	var appraisals []RecordObjectAppraisal
	result := db.Find(&appraisals)
	return appraisals, result.Error
}

func GetRecordObjectConfidentialities() ([]RecordObjectConfidentiality, error) {
	var confidentialities []RecordObjectConfidentiality
	result := db.Find(&confidentialities)
	return confidentialities, result.Error
}

func GetMessageOfProcessByCode(process Process, code string) (Message, error) {
	result := db.Model(&Process{}).
		Preload("Message0501.MessageType").
		Preload("Message0503.MessageType").
		Where(&process).
		First(&process)
	if result.Error != nil {
		log.Fatal("process not found")
	}
	switch code {
	case "0501":
		if process.Message0501 == nil {
			return Message{}, errors.New("process {" + process.XdomeaID + "} has no 0501 message")
		} else {
			return *process.Message0501, nil
		}
	case "0503":
		if process.Message0503 == nil {
			return Message{}, errors.New("process {" + process.XdomeaID + "} has no 0503 message")
		} else {
			return *process.Message0503, nil
		}
	case "0505":
		if process.Message0505 == nil {
			return Message{}, errors.New("process {" + process.XdomeaID + "} has no 0505 message")
		} else {
			return *process.Message0505, nil
		}
	default:
		errorMessage := "unsupported message type with code: " + code
		log.Fatal(errorMessage)
		return Message{}, errors.New(errorMessage)
	}
}

func GetMessagesByCode(code string) ([]Message, error) {
	var messages []Message
	messageType := GetMessageTypeByCode(code)
	result := db.Model(&Message{}).
		Preload("MessageType").
		Preload("MessageHead.Sender.Institution").
		Preload("MessageHead.Sender.AgencyIdentification.Code").
		Preload("MessageHead.Sender.AgencyIdentification.Prefix").
		Preload("MessageHead.Receiver.Institution").
		Preload("MessageHead.Receiver.AgencyIdentification.Code").
		Preload("MessageHead.Receiver.AgencyIdentification.Prefix").
		Where("message_type_id = ?", messageType.ID).
		Find(&messages)
	return messages, result.Error
}

func GetProcessByXdomeaID(xdomeaID string) (Process, error) {
	process := Process{XdomeaID: xdomeaID}
	// if first is used instead of find the error will get logged, that is not desired
	result := db.Model(&Process{}).
		Preload("Agency").
		Preload("Message0501.MessageHead").
		Preload("Message0501.MessageType").
		Preload("Message0503.MessageHead").
		Preload("Message0503.MessageType").
		Preload("ProcessingErrors").
		Preload("ProcessState.Receive0501").
		Preload("ProcessState.Appraisal").
		Preload("ProcessState.Receive0505").
		Preload("ProcessState.Receive0503").
		Preload("ProcessState.FormatVerification").
		Preload("ProcessState.Archiving").
		Where(&process).Limit(1).Find(&process)
	if result.RowsAffected == 0 {
		return process, gorm.ErrRecordNotFound
	}
	return process, result.Error
}

func GetAppraisalByCode(code string) (RecordObjectAppraisal, error) {
	appraisal := RecordObjectAppraisal{Code: code}
	result := db.Where(&appraisal).First(&appraisal)
	return appraisal, result.Error
}

func GetPrimaryFileStorePath(messageID uuid.UUID, primaryDocumentID uint) (string, error) {
	var message Message
	result := db.
		Preload("MessageType").
		First(&message, messageID)
	if result.Error != nil {
		return "", result.Error
	}
	var primaryDocument PrimaryDocument
	result = db.First(&primaryDocument, primaryDocumentID)
	if result.Error != nil {
		return "", result.Error
	}
	return filepath.Join(message.StoreDir, primaryDocument.FileName), nil
}

func AddProcess(
	agency Agency,
	xdomeaID string,
	processStoreDir string,
	institution *string,
) (Process, error) {
	var process Process
	processState, err := AddProcessState()
	if err != nil {
		return process, err
	}
	process = Process{
		Agency:       agency,
		XdomeaID:     xdomeaID,
		StoreDir:     processStoreDir,
		Institution:  institution,
		ProcessState: processState,
	}
	result := db.Save(&process)
	return process, result.Error
}

func AddProcessState() (ProcessState, error) {
	var processState ProcessState
	Receive0501 := ProcessStep{}
	result := db.Save(&Receive0501)
	if result.Error != nil {
		return processState, result.Error
	}
	Appraisal := ProcessStep{}
	result = db.Save(&Appraisal)
	if result.Error != nil {
		return processState, result.Error
	}
	Receive0505 := ProcessStep{}
	result = db.Save(&Receive0505)
	if result.Error != nil {
		return processState, result.Error
	}
	Receive0503 := ProcessStep{}
	result = db.Save(&Receive0503)
	if result.Error != nil {
		return processState, result.Error
	}
	FormatVerification := ProcessStep{}
	result = db.Save(&FormatVerification)
	if result.Error != nil {
		return processState, result.Error
	}
	Archiving := ProcessStep{}
	result = db.Save(&Archiving)
	if result.Error != nil {
		return processState, result.Error
	}
	processState = ProcessState{
		Receive0501:        Receive0501,
		Appraisal:          Appraisal,
		Receive0505:        Receive0505,
		Receive0503:        Receive0503,
		FormatVerification: FormatVerification,
		Archiving:          Archiving,
	}
	result = db.Save(&processState)
	return processState, result.Error
}

func AddMessage(
	agency Agency,
	xdomeaID string,
	processStoreDir string,
	message Message,
) (Process, Message, error) {
	var process Process
	// generate ID for message, propagate the ID to record object children
	// must be done before saving the message in database
	message.ID = uuid.New()
	setRecordObjectsMessageID(&message)
	result := db.Create(&message)
	// The Database failed to create the message.
	if result.Error != nil {
		return process, message, result.Error
	}
	process, err := GetProcessByXdomeaID(xdomeaID)
	// The process was not found. Create a new process.
	if err != nil {
		var institution *string
		// set institution if possible
		if message.MessageHead.Sender.Institution != nil {
			institution = message.MessageHead.Sender.Institution.Name
		}
		process, err = AddProcess(agency, xdomeaID, processStoreDir, institution)
		// The Database failed to create the process for the message.
		if err != nil {
			return process, message, err
		}
	} else {
		// Check if the process has already a message with the type of the given message.
		_, err = GetMessageOfProcessByCode(process, message.MessageType.Code)
		if err == nil {
			// The process has already a message with the type of the parameter.
			log.Fatal("process has already message with type")
		}
	}
	switch message.MessageType.Code {
	case "0501":
		process.Message0501 = &message
		processStep := process.ProcessState.Receive0501
		processStep.Complete = true
		processStep.CompletionTime = time.Now()
		err = UpdateProcessStep(processStep)
		if err != nil {
			log.Fatal(err)
		}
	case "0503":
		process.Message0503 = &message
		processStep := process.ProcessState.Receive0503
		processStep.Complete = true
		processStep.CompletionTime = time.Now()
		err = UpdateProcessStep(processStep)
		if err != nil {
			log.Fatal(err)
		}
	case "0505":
		process.Message0505 = &message
		processStep := process.ProcessState.Receive0505
		processStep.Complete = true
		processStep.CompletionTime = time.Now()
		err = UpdateProcessStep(processStep)
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("unhandled message type: " + message.MessageType.Code)
	}
	result = db.Save(&process)
	return process, message, result.Error
}

func setRecordObjectsMessageID(message *Message) {
	for _, r := range message.RecordObjects {
		if r.FileRecordObject != nil {
			setFileRecordObjectMessageID(message.ID, r.FileRecordObject)
		}
	}
}

func setFileRecordObjectMessageID(messageID uuid.UUID, fileRecordObject *FileRecordObject) {
	fileRecordObject.MessageID = messageID
	for i := range fileRecordObject.Processes {
		setProcessRecordObjectMessageID(messageID, &fileRecordObject.Processes[i])
	}
}

func setProcessRecordObjectMessageID(
	messageID uuid.UUID,
	processRecordObject *ProcessRecordObject,
) {
	processRecordObject.MessageID = messageID
	for i := range processRecordObject.Documents {
		setDocumentRecordObjectMessageID(messageID, &processRecordObject.Documents[i])
	}
}

func setDocumentRecordObjectMessageID(
	messageID uuid.UUID,
	documentRecordObject *DocumentRecordObject,
) {
	documentRecordObject.MessageID = messageID
}

func UpdateProcess(process Process) error {
	result := db.Save(&process)
	return result.Error
}

func UpdateMessage(message Message) error {
	result := db.Save(&message)
	return result.Error
}

func UpdatePrimaryDocument(primaryDocument PrimaryDocument) error {
	result := db.Save(&primaryDocument)
	return result.Error
}

func UpdateProcessStep(processStep ProcessStep) error {
	result := db.Save(&processStep)
	return result.Error
}

func SetFileRecordObjectAppraisal(
	id uuid.UUID,
	appraisalCode string,
	recursiv bool,
) (FileRecordObject, error) {
	fileRecordObject, err := GetFileRecordObjectByID(id)
	if err != nil {
		return fileRecordObject, err
	}
	// check if message appraisal is already completed, if true return error
	message, err := GetCompleteMessageByID(fileRecordObject.MessageID)
	if err != nil {
		return fileRecordObject, err
	}
	if message.AppraisalComplete {
		return fileRecordObject, errors.New("message appraisal already finished")
	}
	// set appraisal
	err = fileRecordObject.SetAppraisal(appraisalCode)
	if err != nil {
		return fileRecordObject, err
	}
	// set appraisal for child elements if recursiv appraisal was choosen
	if recursiv {
		for _, process := range fileRecordObject.Processes {
			_, err = SetProcessRecordObjectAppraisal(process.ID, appraisalCode)
			if err != nil {
				return fileRecordObject, err
			}
		}
	}
	// return updated file record object
	return GetFileRecordObjectByID(id)
}

func SetFileRecordObjectAppraisalNote(
	id uuid.UUID,
	appraisalNote string,
) (FileRecordObject, error) {
	fileRecordObject, err := GetFileRecordObjectByID(id)
	if err != nil {
		return fileRecordObject, err
	}
	// check if message appraisal is already completed, if true return error
	message, err := GetCompleteMessageByID(fileRecordObject.MessageID)
	if err != nil {
		return fileRecordObject, err
	}
	if message.AppraisalComplete {
		return fileRecordObject, errors.New("message appraisal already finished")
	}
	// set note
	err = fileRecordObject.SetAppraisalNote(appraisalNote)
	if err != nil {
		return fileRecordObject, err
	}
	// return updated file record object
	return GetFileRecordObjectByID(id)
}

func SetProcessRecordObjectAppraisal(
	id uuid.UUID,
	appraisalCode string,
) (ProcessRecordObject, error) {
	processRecordObject, err := GetProcessRecordObjectByID(id)
	if err != nil {
		return processRecordObject, err
	}
	// check if message appraisal is already completed, if true return error
	message, err := GetCompleteMessageByID(processRecordObject.MessageID)
	if err != nil {
		return processRecordObject, err
	}
	if message.AppraisalComplete {
		return processRecordObject, errors.New("message appraisal already finished")
	}
	// set appraisal
	err = processRecordObject.SetAppraisal(appraisalCode)
	if err != nil {
		return processRecordObject, err
	}
	// return updated process record object
	return GetProcessRecordObjectByID(id)
}

func SetProcessRecordObjectAppraisalNote(
	id uuid.UUID,
	appraisalNote string,
) (ProcessRecordObject, error) {
	processRecordObject, err := GetProcessRecordObjectByID(id)
	if err != nil {
		return processRecordObject, err
	}
	// check if message appraisal is already completed, if true return error
	message, err := GetCompleteMessageByID(processRecordObject.MessageID)
	if err != nil {
		return processRecordObject, err
	}
	if message.AppraisalComplete {
		return processRecordObject, errors.New("message appraisal already finished")
	}
	// set note
	err = processRecordObject.SetAppraisalNote(appraisalNote)
	if err != nil {
		return processRecordObject, err
	}
	// return updated process record object
	return GetProcessRecordObjectByID(id)
}

func AddProcessingError(e ProcessingError) {
	result := db.Save(&e)
	if result.Error != nil {
		// error handling not possible
		log.Fatal(result.Error)
	}
}

func AddProcessingErrorToProcess(process Process, e ProcessingError) {
	process.ProcessingErrors = append(process.ProcessingErrors, e)
	err := UpdateProcess(process)
	if err != nil {
		// error handling not possible
		log.Fatal(err)
	}
}
