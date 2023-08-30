package db

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Process struct {
	gorm.Model
	ID       uint      `gorm:"primaryKey"`
	Messages []Message `gorm:"many2many:process_messages;"`
}

type Message struct {
	gorm.Model
	ID            uint `gorm:"primaryKey"`
	MessageTypeID uint
	MessageType   MessageType `gorm:"foreignKey:MessageTypeID;references:ID"`
	StoreDir      string
	MessagePath   string
}

type MessageType struct {
	gorm.Model
	ID   uint `gorm:"primaryKey"`
	Code string
}

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
	// Migrate the schema
	db.AutoMigrate(&Process{}, &Message{}, &MessageType{})
	// type0501 := MessageType{Code: "0501"}
	// db.Create(&type0501)
	// message := Message{MessageType: type0501}
	// db.Create(&message)
	// messages := []Message{message}
	// process := Process{Messages: messages}
	// db.Create(&process)
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

func AddMessage(message Message) {
	result := db.Create(&message)
	log.Println(result)
}
