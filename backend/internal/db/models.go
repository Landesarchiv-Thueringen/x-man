package db

import (
	"encoding/xml"
	"time"

	"gorm.io/gorm"
)

// Foreign keys need to be pointers so that the default value is nil.
// The same is true for nullable values.

type Code struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Code      *string        `xml:"code" json:"code"`
	Name      *string        `xml:"name" json:"name"`
}

type Process struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	XdomeaID  string         `json:"xdomeaID"`
	StoreDir  string         `json:"-"`
	Messages  []Message      `gorm:"many2many:process_messages;" json:"messages"`
}

type Message struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	CreatedAt         time.Time      `json:"-"`
	UpdatedAt         time.Time      `json:"-"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
	StoreDir          string         `json:"-"`
	MessagePath       string         `json:"-"`
	XdomeaVersion     string         `json:"xdomeaVersion"`
	MessageHeadID     *uint          `json:"-"`
	MessageHead       MessageHead    `gorm:"foreignKey:MessageHeadID;references:ID" json:"messageHead"`
	MessageTypeID     *uint          `json:"-"`
	MessageType       MessageType    `gorm:"foreignKey:MessageTypeID;references:ID" json:"messageType"`
	AppraisalComplete bool           `json:"appraisalComplete"`
	RecordObjects     []RecordObject `gorm:"many2many:message_record_objects;" json:"recordObjects"`
}

type MessageType struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Code      string         `json:"code"`
}

type Message0501 struct {
	XMLName       xml.Name       `gorm:"-" xml:"Aussonderung.Anbieteverzeichnis.0501" json:"-"`
	MessageHead   MessageHead    `xml:"Kopf" json:"messageHead"`
	RecordObjects []RecordObject `xml:"Schriftgutobjekt" json:"recordObjects"`
}

type Message0503 struct {
	XMLName       xml.Name       `gorm:"-" xml:"Aussonderung.Aussonderung.0503" json:"-"`
	MessageHead   MessageHead    `xml:"Kopf" json:"messageHead"`
	RecordObjects []RecordObject `xml:"Schriftgutobjekt" json:"recordObjects"`
}

type MessageHead struct {
	XMLName      xml.Name       `gorm:"-" xml:"Kopf" json:"-"`
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"-"`
	UpdatedAt    time.Time      `json:"-"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	ProcessID    string         `xml:"ProzessID" json:"processID"`
	CreationTime string         `xml:"Erstellungszeitpunkt" json:"creationTime"`
	SenderID     *uint          `json:"-"`
	Sender       Contact        `gorm:"foreignKey:SenderID;references:ID" xml:"Absender" json:"sender"`
	ReceiverID   *uint          `json:"-"`
	Receiver     Contact        `gorm:"foreignKey:ReceiverID;references:ID" xml:"Empfaenger" json:"receiver"`
}

type Contact struct {
	ID                     uint                 `gorm:"primaryKey" json:"id"`
	CreatedAt              time.Time            `json:"-"`
	UpdatedAt              time.Time            `json:"-"`
	DeletedAt              gorm.DeletedAt       `gorm:"index" json:"-"`
	AgencyIdentificationID *uint                `json:"-"`
	AgencyIdentification   AgencyIdentification `gorm:"foreignKey:AgencyIdentificationID;references:ID" xml:"Behoerdenkennung" json:"agencyIdentification"`
	InstitutionID          *uint                `json:"-"`
	Institution            Institution          `gorm:"foreignKey:InstitutionID;references:ID" xml:"Institution" json:"institution"`
}

type AgencyIdentification struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CodeID    *uint          `json:"-"`
	Code      Code           `gorm:"foreignKey:CodeID;references:ID" xml:"Behoerdenschluessel" json:"code"`
	PrefixID  *uint          `json:"-"`
	Prefix    Code           `gorm:"foreignKey:PrefixID;references:ID" xml:"Praefix" json:"prefix"`
}

type Institution struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"-"`
	UpdatedAt    time.Time      `json:"-"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Name         *string        `xml:"Name"  json:"name"`
	Abbreviation *string        `xml:"Kurzbezeichnung" json:"abbreviation"`
}

type RecordObject struct {
	XMLName            xml.Name         `gorm:"-" xml:"Schriftgutobjekt" json:"-"`
	ID                 uint             `gorm:"primaryKey" json:"id"`
	CreatedAt          time.Time        `json:"-"`
	UpdatedAt          time.Time        `json:"-"`
	DeletedAt          gorm.DeletedAt   `gorm:"index" json:"-"`
	FileRecordObjectID *uint            `json:"-"`
	FileRecordObject   FileRecordObject `gorm:"foreignKey:FileRecordObjectID;references:ID" xml:"Akte" json:"fileRecordObject"`
}

type FileRecordObject struct {
	XMLName           xml.Name              `gorm:"-" xml:"Akte" json:"-"`
	ID                uint                  `gorm:"primaryKey" json:"id"`
	CreatedAt         time.Time             `json:"-"`
	UpdatedAt         time.Time             `json:"-"`
	DeletedAt         gorm.DeletedAt        `gorm:"index" json:"-"`
	GeneralMetadataID *uint                 `json:"-"`
	GeneralMetadata   GeneralMetadata       `gorm:"foreignKey:GeneralMetadataID;references:ID" xml:"AllgemeineMetadaten" json:"generalMetadata"`
	ArchiveMetadataID *uint                 `json:"-"`
	ArchiveMetadata   ArchiveMetadata       `gorm:"foreignKey:ArchiveMetadataID;references:ID" xml:"ArchivspezifischeMetadaten" json:"archiveMetadata"`
	LifetimeID        *uint                 `json:"-"`
	Lifetime          Lifetime              `gorm:"foreignKey:LifetimeID;references:ID" json:"lifetime"`
	Type              *string               `json:"type" xml:"Typ"`
	Processes         []ProcessRecordObject `gorm:"many2many:file_processes;" xml:"Akteninhalt>Vorgang" json:"processes"`
}

type ProcessRecordObject struct {
	XMLName           xml.Name               `gorm:"-" xml:"Vorgang" json:"-"`
	ID                uint                   `gorm:"primaryKey" json:"id"`
	CreatedAt         time.Time              `json:"-"`
	UpdatedAt         time.Time              `json:"-"`
	DeletedAt         gorm.DeletedAt         `gorm:"index" json:"-"`
	GeneralMetadataID *uint                  `json:"-"`
	GeneralMetadata   GeneralMetadata        `gorm:"foreignKey:GeneralMetadataID;references:ID" xml:"AllgemeineMetadaten" json:"generalMetadata"`
	ArchiveMetadataID *uint                  `json:"-"`
	ArchiveMetadata   ArchiveMetadata        `gorm:"foreignKey:ArchiveMetadataID;references:ID" xml:"ArchivspezifischeMetadaten" json:"archiveMetadata"`
	LifetimeID        *uint                  `json:"-"`
	Lifetime          Lifetime               `gorm:"foreignKey:LifetimeID;references:ID" json:"lifetime"`
	Type              *string                `json:"type" xml:"Typ"`
	Documents         []DocumentRecordObject `gorm:"many2many:process_documents;" xml:"Dokument" json:"documents"`
}

type DocumentRecordObject struct {
	XMLName           xml.Name        `gorm:"-" xml:"Dokument" json:"-"`
	ID                uint            `gorm:"primaryKey" json:"id"`
	CreatedAt         time.Time       `json:"-"`
	UpdatedAt         time.Time       `json:"-"`
	DeletedAt         gorm.DeletedAt  `gorm:"index" json:"-"`
	GeneralMetadataID *uint           `json:"-"`
	GeneralMetadata   GeneralMetadata `gorm:"foreignKey:GeneralMetadataID;references:ID" xml:"AllgemeineMetadaten" json:"generalMetadata"`
	Type              *string         `json:"type" xml:"Typ"`
	IncomingDate      *string         `xml:"Posteingangsdatum" json:"incomingDate"`
	OutgoingDate      *string         `xml:"Postausgangsdatum" json:"outgoingDate"`
	DocumentDate      *string         `xml:"DatumDesSchreibens" json:"documentDate"`
}

type GeneralMetadata struct {
	XMLName    xml.Name       `gorm:"-" xml:"AllgemeineMetadaten" json:"-"`
	ID         uint           `gorm:"primaryKey" json:"id"`
	CreatedAt  time.Time      `json:"-"`
	UpdatedAt  time.Time      `json:"-"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	Subject    *string        `xml:"Betreff" json:"subject"`
	XdomeaID   *string        `xml:"Kennzeichen" json:"xdomeaID"`
	FilePlanID *uint          `json:"-"`
	FilePlan   FilePlan       `gorm:"foreignKey:FilePlanID;references:ID" xml:"Aktenplaneinheit" json:"filePlan"`
}

type FilePlan struct {
	XMLName   xml.Name       `gorm:"-" xml:"Aktenplaneinheit" json:"-"`
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	XdomeaID  *string        `xml:"Kennzeichen" json:"xdomeaID"`
}

type Lifetime struct {
	XMLName   xml.Name       `gorm:"-" xml:"Laufzeit" json:"-"`
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Start     *string        `xml:"Beginn" json:"start"`
	End       *string        `xml:"Ende" json:"end"`
}

type ArchiveMetadata struct {
	XMLName             xml.Name       `gorm:"-" xml:"ArchivspezifischeMetadaten" json:"-"`
	ID                  uint           `gorm:"primaryKey" json:"id"`
	CreatedAt           time.Time      `json:"-"`
	UpdatedAt           time.Time      `json:"-"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
	AppraisalCode       *string        `xml:"Aussonderungsart>Aussonderungsart>code" json:"appraisalCode"`
	AppraisalRecommCode *string        `xml:"Bewertungsvorschlag>code" json:"appraisalRecommCode"`
}

type RecordObjectAppraisal struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Code      string         `gorm:"unique" xml:"code" json:"code"`
	ShortDesc string         `json:"shortDesc"`
	Desc      string         `json:"desc"`
}
