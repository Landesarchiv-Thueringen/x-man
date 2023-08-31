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
		&Process{},
		&Message{},
		&MessageType{},
		&MessageHead{},
		&RecordObject{},
		&FileRecordObject{},
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
