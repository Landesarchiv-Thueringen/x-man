package db

import (
	"encoding/xml"

	"gorm.io/gorm"
)

type Process struct {
	gorm.Model
	ID       uint `gorm:"primaryKey"`
	XdomeaID string
	StoreDir string
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

type Message0501 struct {
	XMLName       xml.Name       `xml:"Aussonderung.Anbieteverzeichnis.0501"`
	MessageHead   MessageHead    `xml:"Kopf"`
	RecordObjects []RecordObject `xml:"Schriftgutobjekt"`
}

type MessageHead struct {
	gorm.Model
	XMLName   xml.Name `gorm:"-" xml:"Kopf"`
	ID        uint     `gorm:"primaryKey"`
	ProcessID string   `xml:"ProzessID"`
}

type RecordObject struct {
	gorm.Model
	XMLName            xml.Name `gorm:"-" xml:"Schriftgutobjekt"`
	ID                 uint     `gorm:"primaryKey"`
	FileRecordObjectID uint
	FileRecordObject   FileRecordObject `gorm:"foreignKey:FileRecordObjectID;references:ID" xml:"Akte"`
}

type FileRecordObject struct {
	gorm.Model
	XMLName           xml.Name `gorm:"-" xml:"Akte"`
	ID                uint     `gorm:"primaryKey"`
	GeneralMetadataID uint
	GeneralMetadata   GeneralMetadata `gorm:"foreignKey:GeneralMetadataID;references:ID" xml:"AllgemeineMetadaten"`
	Lifetime          Lifetime
}

type GeneralMetadata struct {
	gorm.Model
	XMLName  xml.Name `gorm:"-" xml:"AllgemeineMetadaten"`
	ID       uint     `gorm:"primaryKey"`
	Subject  string   `xml:"Betreff"`
	XdomeaID string   `xml:"Kennzeichen"`
	FilePlan FilePlan `gorm:"foreignKey:FilePlanID;references:ID" xml:"Aktenplaneinheit"`
}

type FilePlan struct {
	gorm.Model
	XMLName  xml.Name `gorm:"-" xml:"Aktenplaneinheit"`
	ID       uint     `gorm:"primaryKey"`
	XdomeaID string   `xml:"Kennzeichen"`
}

type Lifetime struct {
	gorm.Model
	XMLName xml.Name `gorm:"-" xml:"Laufzeit"`
	ID      uint     `gorm:"primaryKey"`
	Start   string   `xml:"Beginn"`
	End     string   `xml:"Ende"`
}
