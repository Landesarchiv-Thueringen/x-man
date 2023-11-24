package db

import (
	"encoding/xml"
	"errors"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Foreign keys need to be pointers so that the default value is nil.
// The same is true for nullable values.

type XdomeaVersion struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Code      string         `json:"code"`
	URI       string         `json:"uri"`
	XSDPath   string         `json:""`
}

type Process struct {
	ID               uuid.UUID         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	CreatedAt        time.Time         `json:"receivedAt"`
	UpdatedAt        time.Time         `json:"-"`
	DeletedAt        gorm.DeletedAt    `gorm:"index" json:"-"`
	XdomeaID         string            `json:"xdomeaID"`
	StoreDir         string            `json:"-"`
	Institution      *string           `json:"institution"`
	Message0501ID    *uuid.UUID        `json:"-"`
	Message0501      *Message          `gorm:"foreignKey:Message0501ID;references:ID" json:"message0501"`
	Message0503ID    *uuid.UUID        `json:"-"`
	Message0503      *Message          `gorm:"foreignKey:Message0503ID;references:ID" json:"message0503"`
	Message0504Path  *string           `json:"-"`
	Message0505ID    *uuid.UUID        `json:"-"`
	Message0505      *Message          `gorm:"foreignKey:Message0505ID;references:ID" json:"message0505"`
	ProcessingErrors []ProcessingError `gorm:"many2many:process_errors;" json:"processingErrors"`
	ProcessStateID   uint              `json:"-"`
	ProcessState     ProcessState      `gorm:"foreignKey:ProcessStateID;references:ID" json:"processState"`
}

type ProcessState struct {
	ID                       uint           `gorm:"primaryKey" json:"-"`
	CreatedAt                time.Time      `json:"-"`
	UpdatedAt                time.Time      `json:"-"`
	DeletedAt                gorm.DeletedAt `gorm:"index" json:"-"`
	Receive0501StepID        uint           `json:"-"`
	Receive0501              ProcessStep    `gorm:"foreignKey:Receive0501StepID;references:ID" json:"receive0501"`
	AppraisalStepID          uint           `json:"-"`
	Appraisal                ProcessStep    `gorm:"foreignKey:AppraisalStepID;references:ID" json:"appraisal"`
	Receive0505StepID        uint           `json:"-"`
	Receive0505              ProcessStep    `gorm:"foreignKey:Receive0505StepID;references:ID" json:"receive0505"`
	Receive0503StepID        uint           `json:"-"`
	Receive0503              ProcessStep    `gorm:"foreignKey:Receive0503StepID;references:ID" json:"receive0503"`
	FormatVerificationStepID uint           `json:"-"`
	FormatVerification       ProcessStep    `gorm:"foreignKey:FormatVerificationStepID;references:ID" json:"formatVerification"`
	ArchivingStepID          uint           `json:"-"`
	Archiving                ProcessStep    `gorm:"foreignKey:ArchivingStepID;references:ID" json:"archiving"`
}

type ProcessStep struct {
	ID                 uint           `gorm:"primaryKey" json:"-"`
	CreatedAt          time.Time      `json:"-"`
	UpdatedAt          time.Time      `json:"-"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
	Complete           bool           `gorm:"default:false" json:"complete"`
	CompletionTime     time.Time      `json:"completionTime"`
	ItemCount          uint           `gorm:"default:0" json:"itemCount"`
	ItemCompletetCount uint           `gorm:"default:0" json:"itemCompletetCount"`
}

type Message struct {
	ID                     uuid.UUID      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	CreatedAt              time.Time      `json:"-"`
	UpdatedAt              time.Time      `json:"-"`
	DeletedAt              gorm.DeletedAt `gorm:"index" json:"-"`
	TransferDir            string         `json:"-"`
	TransferDirMessagePath string         `json:"-"`
	StoreDir               string         `json:"-"`
	MessagePath            string         `json:"-"`
	XdomeaVersion          string         `json:"xdomeaVersion"`
	MessageHeadID          *uint          `json:"-"`
	MessageHead            MessageHead    `gorm:"foreignKey:MessageHeadID;references:ID" json:"messageHead"`
	MessageTypeID          *uint          `json:"-"`
	MessageType            MessageType    `gorm:"foreignKey:MessageTypeID;references:ID" json:"messageType"`
	RecordObjects          []RecordObject `gorm:"many2many:message_record_objects;" json:"recordObjects"`
	AppraisalComplete      bool           `json:"appraisalComplete"`
	SchemaValidation       bool           `gorm:"default:true" json:"schemaValidation"`
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

type MessageBody0501 struct {
	XMLName xml.Name `gorm:"-" xml:"Aussonderung.Anbieteverzeichnis.0501" json:"-"`
}

type Message0503 struct {
	XMLName       xml.Name       `gorm:"-" xml:"Aussonderung.Aussonderung.0503" json:"-"`
	MessageHead   MessageHead    `xml:"Kopf" json:"messageHead"`
	RecordObjects []RecordObject `xml:"Schriftgutobjekt" json:"recordObjects"`
}

type MessageBody0503 struct {
	XMLName xml.Name `gorm:"-" xml:"Aussonderung.Aussonderung.0503" json:"-"`
}

type Message0505 struct {
	XMLName     xml.Name    `gorm:"-" xml:"Aussonderung.BewertungEmpfangBestaetigen.0505" json:"-"`
	MessageHead MessageHead `xml:"Kopf" json:"messageHead"`
}

type MessageBody0505 struct {
	XMLName xml.Name `gorm:"-" xml:"Aussonderung.BewertungEmpfangBestaetigen.0505" json:"-"`
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
	ID                     uint                  `gorm:"primaryKey" json:"id"`
	CreatedAt              time.Time             `json:"-"`
	UpdatedAt              time.Time             `json:"-"`
	DeletedAt              gorm.DeletedAt        `gorm:"index" json:"-"`
	AgencyIdentificationID *uint                 `json:"-"`
	AgencyIdentification   *AgencyIdentification `gorm:"foreignKey:AgencyIdentificationID;references:ID" xml:"Behoerdenkennung" json:"agencyIdentification"`
	InstitutionID          *uint                 `json:"-"`
	Institution            *Institution          `gorm:"foreignKey:InstitutionID;references:ID" xml:"Institution" json:"institution"`
}

type AgencyIdentification struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CodeID    *uint          `json:"-"`
	Code      *Code          `gorm:"foreignKey:CodeID;references:ID" xml:"Behoerdenschluessel" json:"code"`
	PrefixID  *uint          `json:"-"`
	Prefix    *Code          `gorm:"foreignKey:PrefixID;references:ID" xml:"Praefix" json:"prefix"`
}

type Code struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Code      *string        `xml:"code" json:"code"`
	Name      *string        `xml:"name" json:"name"`
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
	XMLName            xml.Name          `gorm:"-" xml:"Schriftgutobjekt" json:"-"`
	ID                 uint              `gorm:"primaryKey" json:"id"`
	CreatedAt          time.Time         `json:"-"`
	UpdatedAt          time.Time         `json:"-"`
	DeletedAt          gorm.DeletedAt    `gorm:"index" json:"-"`
	FileRecordObjectID *uuid.UUID        `json:"-"`
	FileRecordObject   *FileRecordObject `gorm:"foreignKey:FileRecordObjectID;references:ID" xml:"Akte" json:"fileRecordObject"`
}

type FileRecordObject struct {
	XMLName           xml.Name              `gorm:"-" xml:"Akte" json:"-"`
	ID                uuid.UUID             `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	XdomeaID          uuid.UUID             `xml:"Identifikation>ID" json:"xdomeaID"`
	CreatedAt         time.Time             `json:"-"`
	UpdatedAt         time.Time             `json:"-"`
	DeletedAt         gorm.DeletedAt        `gorm:"index" json:"-"`
	GeneralMetadataID *uint                 `json:"-"`
	GeneralMetadata   *GeneralMetadata      `gorm:"foreignKey:GeneralMetadataID;references:ID" xml:"AllgemeineMetadaten" json:"generalMetadata"`
	ArchiveMetadataID *uint                 `json:"-"`
	ArchiveMetadata   *ArchiveMetadata      `gorm:"foreignKey:ArchiveMetadataID;references:ID" xml:"ArchivspezifischeMetadaten" json:"archiveMetadata"`
	LifetimeID        *uint                 `json:"-"`
	Lifetime          Lifetime              `gorm:"foreignKey:LifetimeID;references:ID" json:"lifetime"`
	Type              *string               `json:"type" xml:"Typ"`
	Processes         []ProcessRecordObject `gorm:"many2many:file_processes;" xml:"Akteninhalt>Vorgang" json:"processes"`
	MessageID         uuid.UUID             `json:"messageID"`
	RecorcObjectType  string                `gorm:"default:file" json:"recordObjectType"`
}

type ProcessRecordObject struct {
	XMLName           xml.Name               `gorm:"-" xml:"Vorgang" json:"-"`
	ID                uuid.UUID              `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	XdomeaID          uuid.UUID              `xml:"Identifikation>ID" json:"xdomeaID"`
	CreatedAt         time.Time              `json:"-"`
	UpdatedAt         time.Time              `json:"-"`
	DeletedAt         gorm.DeletedAt         `gorm:"index" json:"-"`
	GeneralMetadataID *uint                  `json:"-"`
	GeneralMetadata   *GeneralMetadata       `gorm:"foreignKey:GeneralMetadataID;references:ID" xml:"AllgemeineMetadaten" json:"generalMetadata"`
	ArchiveMetadataID *uint                  `json:"-"`
	ArchiveMetadata   *ArchiveMetadata       `gorm:"foreignKey:ArchiveMetadataID;references:ID" xml:"ArchivspezifischeMetadaten" json:"archiveMetadata"`
	LifetimeID        *uint                  `json:"-"`
	Lifetime          Lifetime               `gorm:"foreignKey:LifetimeID;references:ID" json:"lifetime"`
	Type              *string                `json:"type" xml:"Typ"`
	Documents         []DocumentRecordObject `gorm:"many2many:process_documents;" xml:"Dokument" json:"documents"`
	MessageID         uuid.UUID              `json:"messageID"`
	RecorcObjectType  string                 `gorm:"default:process" json:"recordObjectType"`
}

type DocumentRecordObject struct {
	XMLName           xml.Name         `gorm:"-" xml:"Dokument" json:"-"`
	ID                uuid.UUID        `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	XdomeaID          uuid.UUID        `xml:"Identifikation>ID" json:"xdomeaID"`
	CreatedAt         time.Time        `json:"-"`
	UpdatedAt         time.Time        `json:"-"`
	DeletedAt         gorm.DeletedAt   `gorm:"index" json:"-"`
	GeneralMetadataID *uint            `json:"-"`
	GeneralMetadata   *GeneralMetadata `gorm:"foreignKey:GeneralMetadataID;references:ID" xml:"AllgemeineMetadaten" json:"generalMetadata"`
	Type              *string          `json:"type" xml:"Typ"`
	IncomingDate      *string          `xml:"Posteingangsdatum" json:"incomingDate"`
	OutgoingDate      *string          `xml:"Postausgangsdatum" json:"outgoingDate"`
	DocumentDate      *string          `xml:"DatumDesSchreibens" json:"documentDate"`
	MessageID         uuid.UUID        `json:"messageID"`
	RecorcObjectType  string           `gorm:"default:document" json:"recordObjectType"`
	Versions          []Version        `gorm:"many2many:document_versions;" xml:"Version" json:"versions"`
}

type Version struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	VersionID string         `xml:"Nummer" json:"versionID"`
	Formats   []Format       `gorm:"many2many:document_version_formats;" xml:"Format" json:"formats"`
}

type Format struct {
	ID                uint            `gorm:"primaryKey" json:"id"`
	CreatedAt         time.Time       `json:"-"`
	UpdatedAt         time.Time       `json:"-"`
	DeletedAt         gorm.DeletedAt  `gorm:"index" json:"-"`
	Code              string          `xml:"Name>code" json:"code"`
	OtherName         *string         `xml:"SonstigerName" json:"otherName"`
	Version           string          `xml:"Version" json:"version"`
	PrimaryDocumentID uint            `json:"-"`
	PrimaryDocument   PrimaryDocument `gorm:"foreignKey:PrimaryDocumentID;references:ID" xml:"Primaerdokument" json:"primaryDocument"`
}

type PrimaryDocument struct {
	ID                   uint                `gorm:"primaryKey" json:"id"`
	CreatedAt            time.Time           `json:"-"`
	UpdatedAt            time.Time           `json:"-"`
	DeletedAt            gorm.DeletedAt      `gorm:"index" json:"-"`
	FileName             string              `xml:"Dateiname" json:"fileName"`
	FileNameOriginal     *string             `xml:"DateinameOriginal" json:"fileNameOriginal"`
	CreatorName          *string             `xml:"Ersteller" json:"creatorName"`
	CreationTime         *string             `xml:"DatumUhrzeit" json:"creationTime"`
	FormatVerificationID *uint               `json:"-"`
	FormatVerification   *FormatVerification `gorm:"foreignKey:FormatVerificationID;references:ID" json:"formatVerification"`
}

type GeneralMetadata struct {
	XMLName             xml.Name       `gorm:"-" xml:"AllgemeineMetadaten" json:"-"`
	ID                  uint           `gorm:"primaryKey" json:"id"`
	CreatedAt           time.Time      `json:"-"`
	UpdatedAt           time.Time      `json:"-"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
	Subject             *string        `xml:"Betreff" json:"subject"`
	XdomeaID            *string        `xml:"Kennzeichen" json:"xdomeaID"`
	FilePlanID          *uint          `json:"-"`
	FilePlan            FilePlan       `gorm:"foreignKey:FilePlanID;references:ID" xml:"Aktenplaneinheit" json:"filePlan"`
	ConfidentialityCode *string        `xml:"Vertraulichkeitsstufe>code" json:"confidentialityCode"`
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

// code list entries

type RecordObjectAppraisal struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Code      string         `gorm:"unique" xml:"code" json:"code"`
	ShortDesc string         `json:"shortDesc"`
	Desc      string         `json:"desc"`
}

type RecordObjectConfidentiality struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Code      string         `gorm:"unique" xml:"code" json:"code"`
	Desc      string         `json:"desc"`
}

// interfaces and methods

func (message *Message) GetRemoteXmlPath(importDir string) string {
	return filepath.Join(importDir, filepath.Base(message.MessagePath))
}

type RecordObjectIter interface {
	GetChildren() []RecordObjectIter
	GetPrimaryDocuments() []PrimaryDocument
}

func (f *FileRecordObject) GetChildren() []RecordObjectIter {
	recordObjects := []RecordObjectIter{}
	for _, process := range f.Processes {
		recordObjects = append(recordObjects, &process)
		recordObjects = append(recordObjects, process.GetChildren()...)
	}
	return recordObjects
}

func (f *FileRecordObject) GetPrimaryDocuments() []PrimaryDocument {
	primaryDocuments := []PrimaryDocument{}
	for _, process := range f.Processes {
		primaryDocuments = append(primaryDocuments, process.GetPrimaryDocuments()...)
	}
	return primaryDocuments
}

func (p *ProcessRecordObject) GetChildren() []RecordObjectIter {
	recordObjects := []RecordObjectIter{}
	for _, document := range p.Documents {
		recordObjects = append(recordObjects, &document)
		recordObjects = append(recordObjects, document.GetChildren()...)
	}
	return recordObjects
}

func (p *ProcessRecordObject) GetPrimaryDocuments() []PrimaryDocument {
	primaryDocuments := []PrimaryDocument{}
	for _, document := range p.Documents {
		primaryDocuments = append(primaryDocuments, document.GetPrimaryDocuments()...)
	}
	return primaryDocuments
}

func (d *DocumentRecordObject) GetChildren() []RecordObjectIter {
	recordObjects := []RecordObjectIter{}
	return recordObjects
}

func (d *DocumentRecordObject) GetPrimaryDocuments() []PrimaryDocument {
	primaryDocuments := []PrimaryDocument{}
	for _, version := range d.Versions {
		for _, format := range version.Formats {
			primaryDocuments = append(primaryDocuments, format.PrimaryDocument)
		}
	}
	return primaryDocuments
}

type AppraisableRecordObject interface {
	GetAppraisal() (string, error)
	SetAppraisal(string) error
	GetID() uuid.UUID
	GetAppraisableObjects() []AppraisableRecordObject
}

func (f *FileRecordObject) GetAppraisal() (string, error) {
	if f.ArchiveMetadata != nil &&
		f.ArchiveMetadata.AppraisalCode != nil {
		return *f.ArchiveMetadata.AppraisalCode, nil
	}
	return "", errors.New("no appraisal existing")
}

func (f *FileRecordObject) SetAppraisal(appraisalCode string) error {
	if appraisalCode != "A" && appraisalCode != "V" && appraisalCode != "B" {
		return errors.New("unknown appraisal code")
	}
	if f.ArchiveMetadata == nil {
		archiveMetadata := ArchiveMetadata{
			AppraisalCode: &appraisalCode,
		}
		f.ArchiveMetadata = &archiveMetadata
	} else {
		f.ArchiveMetadata.AppraisalCode = &appraisalCode
	}
	return nil
}

func (f *FileRecordObject) GetAppraisableObjects() []AppraisableRecordObject {
	appraisableObjects := []AppraisableRecordObject{f}
	for _, p := range f.Processes {
		appraisableObjects = append(appraisableObjects, &p)
	}
	return appraisableObjects
}

func (f *FileRecordObject) GetID() uuid.UUID {
	return f.XdomeaID
}

func (f *FileRecordObject) GetTitle() string {
	title := "Akte"
	if f.GeneralMetadata != nil {
		if f.GeneralMetadata.XdomeaID != nil {
			title += " " + *f.GeneralMetadata.XdomeaID
		}
		if f.GeneralMetadata.Subject != nil {
			title += ": " + *f.GeneralMetadata.Subject
		}
	}
	return title
}

func (p *ProcessRecordObject) GetAppraisal() (string, error) {
	if p.ArchiveMetadata != nil &&
		p.ArchiveMetadata.AppraisalCode != nil {
		return *p.ArchiveMetadata.AppraisalCode, nil
	}
	return "", errors.New("no appraisal existing")
}

func (p *ProcessRecordObject) SetAppraisal(appraisalCode string) error {
	if appraisalCode != "A" && appraisalCode != "V" && appraisalCode != "B" {
		return errors.New("unknown appraisal code")
	}
	if p.ArchiveMetadata == nil {
		archiveMetadata := ArchiveMetadata{
			AppraisalCode: &appraisalCode,
		}
		p.ArchiveMetadata = &archiveMetadata
	} else {
		p.ArchiveMetadata.AppraisalCode = &appraisalCode
	}
	return nil
}

func (p *ProcessRecordObject) GetID() uuid.UUID {
	return p.XdomeaID
}

func (p *ProcessRecordObject) GetAppraisableObjects() []AppraisableRecordObject {
	appraisableObjects := []AppraisableRecordObject{p}
	return appraisableObjects
}

func (primaryDocument *PrimaryDocument) GetFileName() string {
	if primaryDocument.FileNameOriginal == nil {
		return primaryDocument.FileName
	}
	return *primaryDocument.FileNameOriginal
}

func (primaryDocument *PrimaryDocument) GetRemotePath(importDir string) string {
	return filepath.Join(importDir, primaryDocument.FileName)
}
