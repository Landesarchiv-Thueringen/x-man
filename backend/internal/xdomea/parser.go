package xdomea

import (
	"encoding/xml"
	"io"
	"lath/xdomea/internal/db"
	"log"
	"os"
)

func ParseMessage(message db.Message) db.Message {
	xmlFile, err := os.Open(message.MessagePath)
	if err != nil {
		log.Fatal(err)
	}
	defer xmlFile.Close()
	// Read the opened xmlFile as a byte array.
	xmlBytes, err := io.ReadAll(xmlFile)
	if err != nil {
		log.Fatal(err)
	}
	switch message.MessageType.Code {
	case "0501":
		message = parse0501Message(message, xmlBytes)
	case "0503":
		message = parse0503Message(message, xmlBytes)
	default:
		log.Fatal("message type can't be parsed")
	}
	return message
}

func parse0501Message(message db.Message, xmlBytes []byte) db.Message {
	var m db.Message0501
	err := xml.Unmarshal(xmlBytes, &m)
	if err != nil {
		log.Fatal(err)
	}
	message.MessageHead = m.MessageHead
	message.RecordObjects = m.RecordObjects
	return message
}

func parse0503Message(message db.Message, xmlBytes []byte) db.Message {
	var m db.Message0503
	err := xml.Unmarshal(xmlBytes, &m)
	if err != nil {
		log.Fatal(err)
	}
	message.MessageHead = m.MessageHead
	message.RecordObjects = m.RecordObjects
	return message
}