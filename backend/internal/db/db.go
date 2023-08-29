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
	// Migrate the schema
	database.AutoMigrate(&Process{}, &Message{}, &MessageType{})
	type0501 := MessageType{Code: "0501"}
	database.Create(&type0501)
	message := Message{MessageType: type0501}
	database.Create(&message)
	messages := []Message{message}
	process := Process{Messages: messages}
	database.Create(&process)
	// database.Create(&MessageType{ID: 2, Code: "0502"})
	// database.Create(&MessageType{ID: 3, Code: "0503"})
	// database.Create(&MessageType{ID: 4, Code: "0504"})
	// database.Create(&MessageType{ID: 5, Code: "0505"})
	// database.Create(&MessageType{ID: 6, Code: "0506"})
	// database.Create(&MessageType{ID: 7, Code: "0507"})
	db = database
}
