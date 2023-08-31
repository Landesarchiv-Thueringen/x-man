package xdomea

import (
	"encoding/xml"
	"io/ioutil"
	"lath/xdomea/internal/db"
	"log"
	"os"
)

type Message0501 struct {
	XMLName       xml.Name       `xml:"Aussonderung.Anbieteverzeichnis.0501"`
	MessageHead   MessageHead    `xml:"Kopf"`
	RecordObjects []RecordObject `xml:"Schriftgutobjekt"`
}

type MessageHead struct {
	XMLName   xml.Name `xml:"Kopf"`
	ProcessID string   `xml:"ProzessID"`
}

type RecordObject struct {
	XMLName           xml.Name           `xml:"Schriftgutobjekt"`
	FileRecordObjects []FileRecordObject `xml:"Akte"`
}

type FileRecordObject struct {
	XMLName         xml.Name `xml:"Akte"`
	GeneralMetadata GeneralMetadata
}

type GeneralMetadata struct {
	XMLName xml.Name `xml:"AllgemeineMetadaten"`
	Subject string   `xml:"Betreff"`
}

func ParseMessage(message db.Message) {
	xmlFile, err := os.Open(message.MessagePath)
	if err != nil {
		log.Fatal(err)
	}
	defer xmlFile.Close()
	// read our opened xmlFile as a byte array.
	xmlBytes, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		log.Fatal(err)
	}
	var messageEl Message0501
	xml.Unmarshal(xmlBytes, &messageEl)
	log.Println(messageEl)
}
