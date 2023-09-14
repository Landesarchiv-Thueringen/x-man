package db

import (
	"errors"
	"log"

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
	)
}

func InitMessageTypes(messageTypes []*MessageType) {
	result := db.Create(messageTypes)
	if result.Error != nil {
		log.Fatal("Failed to initialize message types!")
	}
}

func GetMessageByID(id uint) (Message, error) {
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
		Preload("RecordObjects.FileRecordObject.Lifetime").
		Preload("RecordObjects.FileRecordObject.Processes").
		Preload("RecordObjects.FileRecordObject.Processes.GeneralMetadata").
		Preload("RecordObjects.FileRecordObject.Processes.GeneralMetadata.FilePlan").
		Preload("RecordObjects.FileRecordObject.Processes.Lifetime").
		Preload("RecordObjects.FileRecordObject.Processes.Documents").
		Preload("RecordObjects.FileRecordObject.Processes.Documents.GeneralMetadata").
		Preload("RecordObjects.FileRecordObject.Processes.Documents.GeneralMetadata.FilePlan").
		First(&message, id)
	return message, result.Error
}

func GetFileRecordObjectByID(id uint) (FileRecordObject, error) {
	var file FileRecordObject
	result := db.
		Preload("GeneralMetadata").
		Preload("GeneralMetadata.FilePlan").
		Preload("Lifetime").
		Preload("Processes").
		Preload("Processes.GeneralMetadata").
		Preload("Processes.GeneralMetadata.FilePlan").
		Preload("Processes.Lifetime").
		First(&file, id)
	return file, result.Error
}

func GetProcessRecordObjectByID(id uint) (ProcessRecordObject, error) {
	var process ProcessRecordObject
	result := db.
		Preload("GeneralMetadata").
		Preload("GeneralMetadata.FilePlan").
		Preload("Lifetime").
		Preload("Documents").
		Preload("Documents.GeneralMetadata").
		Preload("Documents.GeneralMetadata.FilePlan").
		First(&process, id)
	return process, result.Error
}

func GetDocumentRecordObjectByID(id uint) (DocumentRecordObject, error) {
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
	log.Println(messageType)
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
		Preload("RecordObjects").
		Preload("RecordObjects.FileRecordObject").
		Preload("RecordObjects.FileRecordObject.GeneralMetadata").
		Preload("RecordObjects.FileRecordObject.GeneralMetadata.FilePlan").
		Preload("RecordObjects.FileRecordObject.Lifetime").
		Where("message_type_id = ?", messageType.ID).
		Find(&messages)
	return messages, result.Error
}

func GetProcessByXdomeaID(xdomeaID string) (Process, error) {
	process := Process{XdomeaID: xdomeaID}
	result := db.Model(&Process{}).Preload("Messages").Where(&process).First(&process)
	return process, result.Error
}

func AddMessage(
	xdomeaID string,
	processStoreDir string,
	message Message,
) (Message, error) {
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
