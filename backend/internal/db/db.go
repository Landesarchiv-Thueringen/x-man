package db

import (
	"errors"
	"log"
	"path/filepath"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func Init() {
	dsn := `host=localhost 
		user=lath_xdomea 
		password=temporary 
		dbname=lath_xdomea 
		port=5432 
		sslmode=disable 
		TimeZone=Europe/Berlin`
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database!")
	}
	db = database
}

func Migrate() {
	if db == nil {
		log.Fatal("database wasn't initialized")
	}
	// Migrate the complete schema.
	db.AutoMigrate(
		&XdomeaVersion{},
		&Code{},
		&Process{},
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

func GetProcessingErrors() []ProcessingError {
	var processingErrors []ProcessingError
	result := db.Find(&processingErrors)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return processingErrors
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
		Preload("Message0501.MessageHead").
		Preload("Message0501.MessageType").
		Preload("Message0503.MessageHead").
		Preload("Message0503.MessageType").
		Find(&processes)
	return processes, result.Error
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
	message, error := GetMessageByID(id)
	return message.AppraisalComplete, error
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

func GetAllPrimaryDocuments(messageID uuid.UUID) ([]PrimaryDocument, error) {
	var primaryDocuments []PrimaryDocument
	var documents []DocumentRecordObject
	result := db.Preload("Versions.Formats.PrimaryDocument").
		Where("message_id = ?", messageID.String()).
		Find(&documents)
	if result.Error != nil {
		return primaryDocuments, result.Error
	}
	for _, document := range documents {
		if document.Versions != nil {
			for _, version := range *document.Versions {
				for _, format := range version.Formats {
					primaryDocuments = append(primaryDocuments, format.PrimaryDocument)
				}
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
	result := db.Model(&Process{}).Where(&process).Limit(1).Find(&process)
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

func AddMessage(
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
		// set institution if possible
		var institution string
		if message.MessageHead.Sender.Institution != nil &&
			message.MessageHead.Sender.Institution.Name != nil {
			institution = *message.MessageHead.Sender.Institution.Name
			process = Process{
				XdomeaID:    xdomeaID,
				StoreDir:    processStoreDir,
				Institution: institution,
			}
		} else {
			process = Process{
				XdomeaID: xdomeaID,
				StoreDir: processStoreDir,
			}
		}
		result = db.Create(&process)
		// The Database failed to create the process for the message.
		if result.Error != nil {
			return process, message, result.Error
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
	case "0503":
		process.Message0503 = &message
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

func SetFileRecordObjectAppraisal(
	id uuid.UUID,
	appraisalCode string,
) (FileRecordObject, error) {
	fileRecordObject, err := GetFileRecordObjectByID(id)
	if err != nil {
		return fileRecordObject, err
	}
	message, err := GetCompleteMessageByID(fileRecordObject.MessageID)
	if err != nil {
		return fileRecordObject, err
	}
	if message.AppraisalComplete {
		return fileRecordObject, errors.New("message appraisal already finished")
	}
	appraisal, err := GetAppraisalByCode(appraisalCode)
	if err != nil {
		return fileRecordObject, err
	}
	fileRecordObject.ArchiveMetadata.AppraisalCode = &appraisal.Code
	result := db.Save(&fileRecordObject.ArchiveMetadata)
	if result.Error != nil {
		return fileRecordObject, result.Error
	}
	for _, process := range fileRecordObject.Processes {
		_, err = SetProcessRecordObjectAppraisal(process.ID, appraisalCode)
		if err != nil {
			return fileRecordObject, err
		}
	}
	fileRecordObject, err = GetFileRecordObjectByID(id)
	if err != nil {
		return fileRecordObject, err
	}
	return fileRecordObject, nil
}

func SetProcessRecordObjectAppraisal(
	id uuid.UUID,
	appraisalCode string,
) (ProcessRecordObject, error) {
	processRecordObject, err := GetProcessRecordObjectByID(id)
	if err != nil {
		return processRecordObject, err
	}
	message, err := GetCompleteMessageByID(processRecordObject.MessageID)
	if err != nil {
		return processRecordObject, err
	}
	if message.AppraisalComplete {
		return processRecordObject, errors.New("message appraisal already finished")
	}
	appraisal, err := GetAppraisalByCode(appraisalCode)
	if err != nil {
		return processRecordObject, err
	}
	processRecordObject.ArchiveMetadata.AppraisalCode = &appraisal.Code
	result := db.Save(&processRecordObject.ArchiveMetadata)
	if result.Error != nil {
		return processRecordObject, result.Error
	}
	processRecordObject, err = GetProcessRecordObjectByID(id)
	if err != nil {
		return processRecordObject, err
	}
	return processRecordObject, result.Error
}

func AddProcessingError(error ProcessingError) {
	result := db.Save(&error)
	if result.Error != nil {
		// error handling not possible
		log.Fatal(result.Error)
	}
}
