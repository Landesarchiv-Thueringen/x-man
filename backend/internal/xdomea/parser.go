package xdomea

import (
	"encoding/xml"
	"io/ioutil"
	"lath/xdomea/internal/db"
	"log"
	"os"

	"gorm.io/gorm"
)

type Message0501 struct {
	gorm.Model
	XMLName       xml.Name       `xml:"Aussonderung.Anbieteverzeichnis.0501"`
	MessageHead   MessageHead    `xml:"Kopf"`
	RecordObjects []RecordObject `xml:"Schriftgutobjekt"`
}

type MessageHead struct {
	gorm.Model
	XMLName   xml.Name `xml:"Kopf"`
	ProcessID string   `xml:"ProzessID"`
}

type RecordObject struct {
	gorm.Model
	XMLName           xml.Name           `xml:"Schriftgutobjekt"`
	FileRecordObjects []FileRecordObject `xml:"Akte"`
}

type FileRecordObject struct {
	gorm.Model
	XMLName         xml.Name `xml:"Akte"`
	GeneralMetadata GeneralMetadata
	Lifetime        Lifetime
}

type GeneralMetadata struct {
	gorm.Model
	XMLName  xml.Name `xml:"AllgemeineMetadaten"`
	Subject  string   `xml:"Betreff"`
	XdomeaID string   `xml:"Kennzeichen"`
	FilePlan FilePlan `xml:"Aktenplaneinheit"`
}

type FilePlan struct {
	gorm.Model
	XMLName  xml.Name `xml:"Aktenplaneinheit"`
	XdomeaID string   `xml:"Kennzeichen"`
}

type Lifetime struct {
	gorm.Model
	XMLName xml.Name `xml:"Laufzeit"`
	Start   string   `xml:"Beginn"`
	End     string   `xml:"Ende"`
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
	err = xml.Unmarshal(xmlBytes, &messageEl)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(messageEl)
}
