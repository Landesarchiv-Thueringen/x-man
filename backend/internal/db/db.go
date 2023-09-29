package db

import (
	"errors"
	"log"

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

func GetMessageByID(id uuid.UUID) (Message, error) {
	var message Message
	result := db.
		Preload("MessageType").
		Preload("MessageHead").
		Preload("MessageHead.Sender").
		Preload("MessageHead.Sender.Institution").
		Preload("MessageHead.Sender.AgencyIdentification").
		Preload("MessageHead.Sender.AgencyIdentification.Code").
		Preload("MessageHead.Sender.AgencyIdentification.Prefix").
		Preload("MessageHead.Receiver").
		Preload("MessageHead.Receiver.Institution").
		Preload("MessageHead.Receiver.AgencyIdentification").
		Preload("MessageHead.Receiver.AgencyIdentification.Code").
		Preload("MessageHead.Receiver.AgencyIdentification.Prefix").
		Preload("RecordObjects").
		Preload("RecordObjects.FileRecordObject").
		Preload("RecordObjects.FileRecordObject.GeneralMetadata").
		Preload("RecordObjects.FileRecordObject.GeneralMetadata.FilePlan").
		Preload("RecordObjects.FileRecordObject.ArchiveMetadata").
		Preload("RecordObjects.FileRecordObject.Lifetime").
		Preload("RecordObjects.FileRecordObject.Processes").
		Preload("RecordObjects.FileRecordObject.Processes.GeneralMetadata").
		Preload("RecordObjects.FileRecordObject.Processes.GeneralMetadata.FilePlan").
		Preload("RecordObjects.FileRecordObject.Processes.ArchiveMetadata").
		Preload("RecordObjects.FileRecordObject.Processes.Lifetime").
		Preload("RecordObjects.FileRecordObject.Processes.Documents").
		Preload("RecordObjects.FileRecordObject.Processes.Documents.GeneralMetadata").
		Preload("RecordObjects.FileRecordObject.Processes.Documents.GeneralMetadata.FilePlan").
		First(&message, id)
	return message, result.Error
}

func GetFileRecordObjectByID(id uuid.UUID) (FileRecordObject, error) {
	var file FileRecordObject
	result := db.
		Preload("GeneralMetadata").
		Preload("GeneralMetadata.FilePlan").
		Preload("ArchiveMetadata").
		Preload("Lifetime").
		Preload("Processes").
		Preload("Processes.GeneralMetadata").
		Preload("Processes.GeneralMetadata.FilePlan").
		Preload("Processes.ArchiveMetadata").
		Preload("Processes.Lifetime").
		Preload("Processes.Documents").
		Preload("Processes.Documents.GeneralMetadata").
		Preload("Processes.Documents.GeneralMetadata.FilePlan").
		First(&file, id)
	return file, result.Error
}

func GetProcessRecordObjectByID(id uuid.UUID) (ProcessRecordObject, error) {
	var process ProcessRecordObject
	result := db.
		Preload("GeneralMetadata").
		Preload("GeneralMetadata.FilePlan").
		Preload("ArchiveMetadata").
		Preload("Lifetime").
		Preload("Documents").
		Preload("Documents.GeneralMetadata").
		Preload("Documents.GeneralMetadata.FilePlan").
		First(&process, id)
	return process, result.Error
}

func GetDocumentRecordObjectByID(id uuid.UUID) (DocumentRecordObject, error) {
	var document DocumentRecordObject
	result := db.
		Preload("GeneralMetadata").
		Preload("GeneralMetadata.FilePlan").
		First(&document, id)
	return document, result.Error
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
		Preload("Messages").
		Preload("Messages.MessageType").
		Where(&process).
		First(&process)
	if result.Error != nil {
		log.Fatal("process not found")
	}
	var message Message
	for _, m := range process.Messages {
		if m.MessageType.Code == code {
			return message, nil
		}
	}
	return Message{}, errors.New("no message with type found")
}

func GetMessagesByCode(code string) ([]Message, error) {
	var messages []Message
	messageType := GetMessageTypeByCode(code)
	result := db.Model(&Message{}).
		Preload("MessageType").
		Preload("MessageHead").
		Preload("MessageHead.Sender").
		Preload("MessageHead.Sender.Institution").
		Preload("MessageHead.Sender.AgencyIdentification").
		Preload("MessageHead.Sender.AgencyIdentification.Code").
		Preload("MessageHead.Sender.AgencyIdentification.Prefix").
		Preload("MessageHead.Receiver").
		Preload("MessageHead.Receiver.Institution").
		Preload("MessageHead.Receiver.AgencyIdentification").
		Preload("MessageHead.Receiver.AgencyIdentification.Code").
		Preload("MessageHead.Receiver.AgencyIdentification.Prefix").
		Where("message_type_id = ?", messageType.ID).
		Find(&messages)
	return messages, result.Error
}

func GetProcessByXdomeaID(xdomeaID string) (Process, error) {
	process := Process{XdomeaID: xdomeaID}
	result := db.Model(&Process{}).Preload("Messages").Where(&process).First(&process)
	return process, result.Error
}

func GetAppraisalByCode(code string) (RecordObjectAppraisal, error) {
	appraisal := RecordObjectAppraisal{Code: code}
	result := db.Where(&appraisal).First(&appraisal)
	return appraisal, result.Error
}

func AddMessage(
	xdomeaID string,
	processStoreDir string,
	message Message,
) (Message, error) {
	// generate ID for message, propagate the ID to record object children
	// must be done before saving the message in database
	message.ID = uuid.New()
	setRecordObjectsMessageID(&message)
	result := db.Create(&message)
	// The Database failed to create the message.
	if result.Error != nil {
		return message, result.Error
	}
	process, err := GetProcessByXdomeaID(xdomeaID)
	// The process was not found. Create a new process.
	if err != nil {
		process = Process{XdomeaID: xdomeaID, StoreDir: processStoreDir}
		result = db.Create(&process)
		// The Database failed to create the process for the message.
		if result.Error != nil {
			return message, result.Error
		}
	} else {
		// Check if the process has already a message with the type of the given message.
		_, err = GetMessageOfProcessByCode(process, message.MessageType.Code)
		if err == nil {
			// The process has already a message with the type of the parameter.
			log.Fatal("process has already message with type")
		}
	}
	process.Messages = append(process.Messages, message)
	result = db.Save(&process)
	return message, result.Error
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
	appraisal, err := GetAppraisalByCode(appraisalCode)
	if err != nil {
		return processRecordObject, err
	}
	processRecordObject.ArchiveMetadata.AppraisalCode = &appraisal.Code
	result := db.Save(&processRecordObject.ArchiveMetadata)
	if result.Error != nil {
		return processRecordObject, result.Error
	}
	// adsdasd
	processRecordObject, err = GetProcessRecordObjectByID(id)
	if err != nil {
		return processRecordObject, err
	}
	return processRecordObject, result.Error
}
