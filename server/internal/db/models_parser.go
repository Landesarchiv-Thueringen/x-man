package db

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Foreign keys need to be pointers so that the default value is nil.
// The same is true for Nullable values.

type XdomeaVersion struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	Code      string    `json:"code"`
	URI       string    `json:"uri"`
	XSDPath   string    `json:""`
}

func (v *XdomeaVersion) IsVersionPriorTo300() bool {
	return v.Code == "2.3.0" || v.Code == "2.4.0"
}

type Process struct {
	// ID is the process ID as parsed from an Xdomea message (ProzessID). It is
	// a UUID, but since that cannot be verified conveniently when read from an
	// XML file, we store it as string.
	ID               string            `gorm:"primaryKey;" json:"id"`
	CreatedAt        time.Time         `json:"receivedAt"`
	UpdatedAt        time.Time         `json:"-"`
	AgencyID         uint              `json:"-"`
	Agency           Agency            `gorm:"foreignKey:AgencyID;references:ID" json:"agency"`
	StoreDir         string            `json:"-"`
	Institution      *string           `json:"institution"`
	Note             *string           `json:"note"`
	Message0501ID    *uuid.UUID        `json:"message0501Id"`
	Message0501      *Message          `gorm:"foreignKey:Message0501ID;references:ID;constraint:OnDelete:SET NULL;" json:"message0501"`
	Message0502Path  *string           `json:"-"`
	Message0503ID    *uuid.UUID        `json:"message0503Id"`
	Message0503      *Message          `gorm:"foreignKey:Message0503ID;references:ID;constraint:OnDelete:SET NULL;" json:"message0503"`
	Message0504Path  *string           `json:"-"`
	Message0505ID    *uuid.UUID        `json:"message0505Id"`
	Message0505      *Message          `gorm:"foreignKey:Message0505ID;references:ID;constraint:OnDelete:SET NULL;" json:"message0505"`
	ProcessingErrors []ProcessingError `json:"processingErrors"`
	ProcessStateID   uint              `json:"-"`
	ProcessState     ProcessState      `gorm:"foreignKey:ProcessStateID;references:ID;constraint:OnDelete:SET NULL;" json:"processState"`
	Tasks            []Task            `json:"-"`
}

// BeforeDelete deletes associated rows of the deleted Process.
func (p *Process) BeforeDelete(tx *gorm.DB) (err error) {
	if p.ID == "" {
		return fmt.Errorf("failed to delete associations for Process")
	}
	process := Process{ID: p.ID}
	tx.Preload(clause.Associations).First(&process)
	if process.Message0501 != nil {
		tx.Delete(&process.Message0501)
	}
	if process.Message0503 != nil {
		aips := GetArchivePackages(p.ID)
		for _, aip := range aips {
			tx.Delete(&aip)
		}
		tx.Delete(&process.Message0503)
	}
	if process.Message0505 != nil {
		tx.Delete(&process.Message0505)
	}
	for _, e := range process.ProcessingErrors {
		tx.Delete(&e)
	}
	tx.Delete(&process.ProcessState)
	for _, t := range process.Tasks {
		tx.Delete(&t)
	}
	return
}

type ProcessState struct {
	ID                       uint        `gorm:"primaryKey" json:"-"`
	CreatedAt                time.Time   `json:"-"`
	UpdatedAt                time.Time   `json:"-"`
	Receive0501StepID        uint        `json:"-"`
	Receive0501              ProcessStep `gorm:"foreignKey:Receive0501StepID;references:ID;constraint:OnDelete:SET NULL" json:"receive0501"`
	AppraisalStepID          uint        `json:"-"`
	Appraisal                ProcessStep `gorm:"foreignKey:AppraisalStepID;references:ID;constraint:OnDelete:SET NULL" json:"appraisal"`
	Receive0505StepID        uint        `json:"-"`
	Receive0505              ProcessStep `gorm:"foreignKey:Receive0505StepID;references:ID;constraint:OnDelete:SET NULL" json:"receive0505"`
	Receive0503StepID        uint        `json:"-"`
	Receive0503              ProcessStep `gorm:"foreignKey:Receive0503StepID;references:ID;constraint:OnDelete:SET NULL" json:"receive0503"`
	FormatVerificationStepID uint        `json:"-"`
	FormatVerification       ProcessStep `gorm:"foreignKey:FormatVerificationStepID;references:ID;constraint:OnDelete:SET NULL" json:"formatVerification"`
	ArchivingStepID          uint        `json:"-"`
	Archiving                ProcessStep `gorm:"foreignKey:ArchivingStepID;references:ID;constraint:OnDelete:SET NULL" json:"archiving"`
}

// BeforeDelete deletes associated rows of the deleted ProcessState.
func (s *ProcessState) BeforeDelete(tx *gorm.DB) (err error) {
	if s.ID == 0 {
		return fmt.Errorf("failed to delete associations for ProcessState")
	}
	processState := ProcessState{ID: s.ID}
	tx.Preload(clause.Associations).First(&processState)
	tx.Delete(&processState.Receive0501)
	tx.Delete(&processState.Appraisal)
	tx.Delete(&processState.Receive0505)
	tx.Delete(&processState.Receive0503)
	tx.Delete(&processState.FormatVerification)
	tx.Delete(&processState.Archiving)
	return
}

type ProcessStep struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"updateTime"`
	// Complete is true if the step completed successfully.
	Complete bool `gorm:"default:false" json:"complete"`
	// CompletionTime is the time at which Complete was set to true.
	CompletionTime *time.Time `json:"completionTime"`
	// CompletedBy is the name of the user who performed the process step.
	CompletedBy *string `json:"completedBy"`
	// Message is a short notice that indicates the state of a not yet completed
	// process step.
	Message *string `json:"message"`
	// Tasks are all tasks associated with the process step.
	//
	// A task of the state "running" indicates that the process step is
	// currently in progress. Tasks of the state "failed" do *not* necessarily
	// indicate a failed process step (see ProcessingErrors). Also, a task of
	// the state "succeeded" is *not* a requirement for a complete process step.
	//
	// Not all process steps (completed or not) have tasks.
	Tasks []Task `gorm:"constraint:OnDelete:SET NULL" json:"tasks"`
	// ProcessingErrors are all processing errors associated with the process
	// step.
	//
	// An unresolved processing error indicates that the process step is
	// currently in a failed state.
	ProcessingErrors []ProcessingError `gorm:"constraint:OnDelete:SET NULL" json:"processingErrors"`
}

type Message struct {
	ID                    uuid.UUID              `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	CreatedAt             time.Time              `json:"-"`
	UpdatedAt             time.Time              `json:"-"`
	TransferDirPath       string                 `json:"-"`
	StoreDir              string                 `json:"-"`
	MessagePath           string                 `json:"-"`
	XdomeaVersion         string                 `json:"xdomeaVersion"`
	MessageHeadID         *uint                  `json:"-"`
	MessageHead           MessageHead            `gorm:"foreignKey:MessageHeadID;references:ID;constraint:OnDelete:SET NULL" json:"messageHead"`
	MessageTypeID         *uint                  `json:"-"`
	MessageType           MessageType            `gorm:"foreignKey:MessageTypeID;references:ID" json:"messageType"`
	SchemaValidation      bool                   `gorm:"default:true" json:"schemaValidation"`
	MaxRecordObjectDepth  uint                   `json:"maxRecordObjectDepth"`
	FileRecordObjects     []FileRecordObject     `gorm:"foreignKey:ParentMessageID" json:"fileRecordObjects"`
	ProcessRecordObjects  []ProcessRecordObject  `gorm:"foreignKey:ParentMessageID" json:"processRecordObjects"`
	DocumentRecordObjects []DocumentRecordObject `gorm:"foreignKey:ParentMessageID" json:"documentRecordObjects"`
	ProcessingErrors      []ProcessingError      `json:"processingErrors"`
	MessageJSON           string                 `gorm:"type:text" json:"-"`
}

// BeforeDelete deletes associated rows of the deleted Message.
func (m *Message) BeforeDelete(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		return fmt.Errorf("failed to delete associations for Message")
	}
	message := Message{ID: m.ID}
	tx.Preload(clause.Associations).First(&message)
	tx.Delete(&message.MessageHead)
	for _, o := range message.FileRecordObjects {
		tx.Delete(&o)
	}
	for _, o := range message.ProcessRecordObjects {
		tx.Delete(&o)
	}
	for _, o := range message.DocumentRecordObjects {
		tx.Delete(&o)
	}
	for _, e := range message.ProcessingErrors {
		tx.Delete(&e)
	}
	return
}

type Message0501 struct {
	XMLName               xml.Name               `gorm:"-" xml:"Aussonderung.Anbieteverzeichnis.0501" json:"-"`
	MessageHead           MessageHead            `xml:"Kopf" json:"messageHead"`
	FileRecordObjects     []FileRecordObject     `xml:"Schriftgutobjekt>Akte" json:"fileRecordObjects"`
	ProcessRecordObjects  []ProcessRecordObject  `xml:"Schriftgutobjekt>Vorgang" json:"processRecordObjects"`
	DocumentRecordObjects []DocumentRecordObject `xml:"Schriftgutobjekt>Dokument" json:"documentRecordObjects"`
}

type MessageBody0501 struct {
	XMLName xml.Name `gorm:"-" xml:"Aussonderung.Anbieteverzeichnis.0501" json:"-"`
}

type Message0503 struct {
	XMLName               xml.Name               `gorm:"-" xml:"Aussonderung.Aussonderung.0503" json:"-"`
	MessageHead           MessageHead            `xml:"Kopf" json:"messageHead"`
	FileRecordObjects     []FileRecordObject     `xml:"Schriftgutobjekt>Akte" json:"fileRecordObjects"`
	ProcessRecordObjects  []ProcessRecordObject  `xml:"Schriftgutobjekt>Vorgang" json:"processRecordObjects"`
	DocumentRecordObjects []DocumentRecordObject `xml:"Schriftgutobjekt>Dokument" json:"documentRecordObjects"`
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
	XMLName      xml.Name  `gorm:"-" xml:"Kopf" json:"-"`
	ID           uint      `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
	ProcessID    string    `xml:"ProzessID" json:"processID"`
	CreationTime string    `xml:"Erstellungszeitpunkt" json:"creationTime"`
	SenderID     *uint     `json:"-"`
	Sender       Contact   `gorm:"foreignKey:SenderID;references:ID;constraint:OnDelete:SET NULL" xml:"Absender" json:"sender"`
	ReceiverID   *uint     `json:"-"`
	Receiver     Contact   `gorm:"foreignKey:ReceiverID;references:ID;constraint:OnDelete:SET NULL" xml:"Empfaenger" json:"receiver"`
}

// BeforeDelete deletes associated rows of the deleted MessageHead.
func (m *MessageHead) BeforeDelete(tx *gorm.DB) (err error) {
	if m.ID == 0 {
		return fmt.Errorf("failed to delete associations for MessageHead")
	}
	message := MessageHead{ID: m.ID}
	tx.Preload(clause.Associations).First(&message)
	tx.Delete(&message.Sender)
	tx.Delete(&message.Receiver)
	return
}

type Contact struct {
	ID                     uint                  `gorm:"primaryKey" json:"id"`
	CreatedAt              time.Time             `json:"-"`
	UpdatedAt              time.Time             `json:"-"`
	AgencyIdentificationID *uint                 `json:"-"`
	AgencyIdentification   *AgencyIdentification `gorm:"constraint:OnDelete:SET NULL" xml:"Behoerdenkennung" json:"agencyIdentification"`
	InstitutionID          *uint                 `json:"-"`
	Institution            *Institution          `gorm:"constraint:OnDelete:SET NULL" xml:"Institution" json:"institution"`
}

// BeforeDelete deletes associated rows of the deleted Contact.
func (c *Contact) BeforeDelete(tx *gorm.DB) (err error) {
	if c.ID == 0 {
		return fmt.Errorf("failed to delete associations for Contact")
	}
	message := Contact{ID: c.ID}
	tx.Preload(clause.Associations).First(&message)
	tx.Delete(&message.AgencyIdentification)
	tx.Delete(&message.Institution)
	return
}

type AgencyIdentification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	Code      *string   `xml:"Behoerdenschluessel>code" json:"code"`
	Prefix    *string   `xml:"Praefix>code" json:"prefix"`
}

type AgencyIdentificationVersionIndependent struct {
	Code       *string `xml:"Behoerdenschluessel>code"`
	CodePre300 *string `xml:"Kennung>code"`
	Prefix     *string `xml:"Praefix>code"`
}

// UnmarshalXML corrects version specific differences of the agency identification.
func (agencyIdentification *AgencyIdentification) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var temp AgencyIdentificationVersionIndependent
	err := d.DecodeElement(&temp, &start)
	if err != nil {
		return err
	}
	agencyIdentification.Prefix = temp.Prefix
	if temp.Code != nil {
		agencyIdentification.Code = temp.Code
	} else if temp.CodePre300 != nil {
		agencyIdentification.Code = temp.CodePre300
	}
	return nil
}

type Institution struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
	Name         *string   `xml:"Name"  json:"name"`
	Abbreviation *string   `xml:"Kurzbezeichnung" json:"abbreviation"`
}

type FileRecordObject struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	XdomeaID  uuid.UUID `json:"xdomeaID"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	// MessageID is the message the FileRecordObject belongs to, either as the
	// Message's direct child or further down the tree.
	MessageID uuid.UUID `json:"messageID"`
	// FileRecordObject has either a Message or a FileRecordObject as parent.
	ParentMessageID       *uuid.UUID             `json:"-"`
	ParentFileRecordID    *uuid.UUID             `json:"-"`
	RecordObjectType      string                 `gorm:"default:file" json:"recordObjectType"`
	GeneralMetadataID     *uint                  `json:"-"`
	GeneralMetadata       *GeneralMetadata       `gorm:"foreignKey:GeneralMetadataID;references:ID;constraint:OnDelete:SET NULL" json:"generalMetadata"`
	ArchiveMetadataID     *uint                  `json:"-"`
	ArchiveMetadata       *ArchiveMetadata       `gorm:"foreignKey:ArchiveMetadataID;references:ID;constraint:OnDelete:SET NULL" json:"archiveMetadata"`
	LifetimeID            *uint                  `json:"-"`
	Lifetime              *Lifetime              `gorm:"foreignKey:LifetimeID;references:ID;constraint:OnDelete:SET NULL" json:"lifetime"`
	Type                  *string                `json:"type"`
	ProcessRecordObjects  []ProcessRecordObject  `gorm:"foreignKey:ParentFileRecordID" json:"processes"`
	SubFileRecordObjects  []FileRecordObject     `gorm:"foreignKey:ParentFileRecordID" json:"subfiles"`
	DocumentRecordObjects []DocumentRecordObject `gorm:"foreignKey:ParentFileRecordID" json:"documents"`
}

// BeforeDelete deletes associated rows of the deleted FileRecordObject.
func (f *FileRecordObject) BeforeDelete(tx *gorm.DB) (err error) {
	if f.ID == uuid.Nil {
		return fmt.Errorf("failed to delete associations for FileRecordObject")
	}
	fileRecord := FileRecordObject{ID: f.ID}
	tx.Preload(clause.Associations).First(&fileRecord)
	if fileRecord.GeneralMetadata != nil {
		tx.Delete(&fileRecord.GeneralMetadata)
	}
	if fileRecord.ArchiveMetadata != nil {
		tx.Delete(&fileRecord.ArchiveMetadata)
	}
	if fileRecord.Lifetime != nil {
		tx.Delete(&fileRecord.Lifetime)
	}
	for _, o := range fileRecord.ProcessRecordObjects {
		tx.Delete(&o)
	}
	for _, o := range fileRecord.SubFileRecordObjects {
		tx.Delete(&o)
	}
	for _, o := range fileRecord.DocumentRecordObjects {
		tx.Delete(&o)
	}
	return
}

type FileRecordObjectVersionDifferences struct {
	XdomeaID                   uuid.UUID              `xml:"Identifikation>ID"`
	GeneralMetadata            *GeneralMetadata       `xml:"AllgemeineMetadaten"`
	ArchiveMetadata            *ArchiveMetadata       `xml:"ArchivspezifischeMetadaten"`
	Lifetime                   *Lifetime              `xml:"Laufzeit"`
	Type                       *string                `xml:"Typ" json:"type"`
	Processes                  []ProcessRecordObject  `xml:"Akteninhalt>Vorgang"`
	DocumentRecordObjects      []DocumentRecordObject `xml:"Akteninhalt>Dokument"`
	SubFileRecordObjects       []FileRecordObject     `xml:"Akteninhalt>Teilakte"`
	SubFileRecordObjectsPre300 []FileRecordObject     `xml:"Teilakte" json:"-"`
}

// UnmarshalXML corrects version specific differences of file record objects.
func (fileRecordObject *FileRecordObject) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var temp FileRecordObjectVersionDifferences
	err := d.DecodeElement(&temp, &start)
	if err != nil {
		return err
	}
	fileRecordObject.XdomeaID = temp.XdomeaID
	fileRecordObject.GeneralMetadata = temp.GeneralMetadata
	fileRecordObject.ArchiveMetadata = temp.ArchiveMetadata
	fileRecordObject.Lifetime = temp.Lifetime
	fileRecordObject.Type = temp.Type
	fileRecordObject.ProcessRecordObjects = temp.Processes
	fileRecordObject.DocumentRecordObjects = temp.DocumentRecordObjects
	if temp.SubFileRecordObjects != nil {
		fileRecordObject.SubFileRecordObjects = temp.SubFileRecordObjects
	} else if temp.SubFileRecordObjectsPre300 != nil {
		fileRecordObject.SubFileRecordObjects = temp.SubFileRecordObjectsPre300
	}
	return nil
}

type ProcessRecordObject struct {
	ID                      uuid.UUID              `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	XdomeaID                uuid.UUID              `xml:"Identifikation>ID" json:"xdomeaID"`
	CreatedAt               time.Time              `json:"-"`
	UpdatedAt               time.Time              `json:"-"`
	MessageID               uuid.UUID              `json:"messageID"`
	ParentMessageID         *uuid.UUID             `json:"-"`
	ParentFileRecordID      *uuid.UUID             `json:"-"`
	ParentProcessRecordID   *uuid.UUID             `json:"-"`
	RecordObjectType        string                 `gorm:"default:process" json:"recordObjectType"`
	GeneralMetadataID       *uint                  `json:"-"`
	GeneralMetadata         *GeneralMetadata       `gorm:"foreignKey:GeneralMetadataID;references:ID;constraint:OnDelete:SET NULL" xml:"AllgemeineMetadaten" json:"generalMetadata"`
	ArchiveMetadataID       *uint                  `json:"-"`
	ArchiveMetadata         *ArchiveMetadata       `gorm:"foreignKey:ArchiveMetadataID;references:ID;constraint:OnDelete:SET NULL" xml:"ArchivspezifischeMetadaten" json:"archiveMetadata"`
	LifetimeID              *uint                  `json:"-"`
	Lifetime                *Lifetime              `gorm:"foreignKey:LifetimeID;references:ID;constraint:OnDelete:SET NULL" json:"lifetime"`
	Type                    *string                `json:"type" xml:"Typ"`
	SubProcessRecordObjects []ProcessRecordObject  `gorm:"foreignKey:ParentProcessRecordID" xml:"Teilvorgang" json:"subprocesses"`
	DocumentRecordObjects   []DocumentRecordObject `gorm:"foreignKey:ParentProcessRecordID" xml:"Dokument" json:"documents"`
}

// BeforeDelete deletes associated rows of the deleted ProcessRecordObject.
func (f *ProcessRecordObject) BeforeDelete(tx *gorm.DB) (err error) {
	if f.ID == uuid.Nil {
		return fmt.Errorf("failed to delete associations for ProcessRecordObject")
	}
	processRecord := ProcessRecordObject{ID: f.ID}
	tx.Preload(clause.Associations).First(&processRecord)
	if processRecord.GeneralMetadata != nil {
		tx.Delete(&processRecord.GeneralMetadata)
	}
	if processRecord.ArchiveMetadata != nil {
		tx.Delete(&processRecord.ArchiveMetadata)
	}
	if processRecord.Lifetime != nil {
		tx.Delete(&processRecord.Lifetime)
	}
	for _, o := range processRecord.SubProcessRecordObjects {
		tx.Delete(&o)
	}
	for _, o := range processRecord.DocumentRecordObjects {
		tx.Delete(&o)
	}
	return
}

type DocumentRecordObject struct {
	XMLName                xml.Name               `gorm:"-" json:"-"`
	ID                     uuid.UUID              `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	XdomeaID               uuid.UUID              `xml:"Identifikation>ID" json:"xdomeaID"`
	CreatedAt              time.Time              `json:"-"`
	UpdatedAt              time.Time              `json:"-"`
	MessageID              uuid.UUID              `json:"messageID"`
	ParentMessageID        *uuid.UUID             `json:"-"`
	ParentFileRecordID     *uuid.UUID             `json:"-"`
	ParentProcessRecordID  *uuid.UUID             `json:"-"`
	ParentDocumentRecordID *uuid.UUID             `json:"-"`
	RecordObjectType       string                 `gorm:"default:document" json:"recordObjectType"`
	GeneralMetadataID      *uint                  `json:"-"`
	GeneralMetadata        *GeneralMetadata       `gorm:"foreignKey:GeneralMetadataID;references:ID;constraint:OnDelete:SET NULL" xml:"AllgemeineMetadaten" json:"generalMetadata"`
	Type                   *string                `json:"type" xml:"Typ"`
	IncomingDate           *string                `xml:"Posteingangsdatum" json:"incomingDate"`
	OutgoingDate           *string                `xml:"Postausgangsdatum" json:"outgoingDate"`
	DocumentDate           *string                `xml:"DatumDesSchreibens" json:"documentDate"`
	Versions               []Version              `xml:"Version" json:"versions"`
	Attachments            []DocumentRecordObject `gorm:"foreignKey:ParentDocumentRecordID" xml:"Anlage" json:"attachments"`
}

// BeforeDelete deletes associated rows of the deleted DocumentRecordObject.
func (d *DocumentRecordObject) BeforeDelete(tx *gorm.DB) (err error) {
	if d.ID == uuid.Nil {
		return fmt.Errorf("failed to delete associations for DocumentRecordObject")
	}
	document := DocumentRecordObject{ID: d.ID}
	tx.Preload(clause.Associations).First(&document)
	if document.GeneralMetadata != nil {
		tx.Delete(&document.GeneralMetadata)
	}
	for _, v := range document.Versions {
		tx.Delete(&v)
	}
	for _, a := range document.Attachments {
		tx.Delete(&a)
	}
	return
}

type Version struct {
	ID                     uint      `gorm:"primaryKey" json:"id"`
	CreatedAt              time.Time `json:"-"`
	UpdatedAt              time.Time `json:"-"`
	DocumentRecordObjectID uuid.UUID `json:"-"`
	VersionID              string    `xml:"Nummer" json:"versionID"`
	Formats                []Format  `xml:"Format" json:"formats"`
}

// BeforeDelete deletes associated rows of the deleted Version.
func (v *Version) BeforeDelete(tx *gorm.DB) (err error) {
	if v.ID == 0 {
		return fmt.Errorf("failed to delete associations for Version")
	}
	version := Version{ID: v.ID}
	tx.Preload(clause.Associations).First(&version)
	for _, f := range version.Formats {
		tx.Delete(&f)
	}
	return
}

type Format struct {
	ID                uint            `gorm:"primaryKey" json:"id"`
	CreatedAt         time.Time       `json:"-"`
	UpdatedAt         time.Time       `json:"-"`
	VersionID         uint            `json:"-"`
	Code              string          `xml:"Name>code" json:"code"`
	OtherName         *string         `xml:"SonstigerName" json:"otherName"`
	Version           string          `xml:"Version" json:"version"`
	PrimaryDocumentID uint            `json:"-"`
	PrimaryDocument   PrimaryDocument `gorm:"foreignKey:PrimaryDocumentID;references:ID;constraint:OnDelete:SET NULL" xml:"Primaerdokument" json:"primaryDocument"`
}

// BeforeDelete deletes associated rows of the deleted Format.
func (v *Format) BeforeDelete(tx *gorm.DB) (err error) {
	if v.ID == 0 {
		return fmt.Errorf("failed to delete associations for Format")
	}
	format := Format{ID: v.ID}
	tx.Preload(clause.Associations).First(&format)
	tx.Delete(&format.PrimaryDocument)
	return
}

type PrimaryDocument struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	CreatedAt        time.Time `json:"-"`
	UpdatedAt        time.Time `json:"-"`
	FileName         string    `xml:"Dateiname" json:"fileName"`
	FileNameOriginal *string   `xml:"DateinameOriginal" json:"fileNameOriginal"`
	// FileSize is the file's size on disk in bytes.
	FileSize             uint64              `json:"fileSize"`
	CreatorName          *string             `xml:"Ersteller" json:"creatorName"`
	CreationTime         *string             `xml:"DatumUhrzeit" json:"creationTime"`
	FormatVerificationID *uint               `json:"-"`
	FormatVerification   *FormatVerification `gorm:"foreignKey:FormatVerificationID;references:ID;constraint:OnDelete:SET NULL" json:"formatVerification"`
}

// BeforeDelete deletes associated rows of the deleted PrimaryDocument.
func (p *PrimaryDocument) BeforeDelete(tx *gorm.DB) (err error) {
	if p.ID == 0 {
		return fmt.Errorf("failed to delete associations for PrimaryDocument")
	}
	primaryDocument := PrimaryDocument{ID: p.ID}
	tx.Preload(clause.Associations).First(&primaryDocument)
	if primaryDocument.FormatVerification != nil {
		tx.Delete(&primaryDocument.FormatVerification)
	}
	return
}

type GeneralMetadata struct {
	XMLName              xml.Name              `gorm:"-" xml:"AllgemeineMetadaten" json:"-"`
	ID                   uint                  `gorm:"primaryKey" json:"id"`
	CreatedAt            time.Time             `json:"-"`
	UpdatedAt            time.Time             `json:"-"`
	Subject              *string               `xml:"Betreff" json:"subject"`
	XdomeaID             *string               `xml:"Kennzeichen" json:"xdomeaID"`
	FilePlanID           *uint                 `json:"-"`
	FilePlan             *FilePlan             `gorm:"foreignKey:FilePlanID;references:ID;constraint:OnDelete:SET NULL" xml:"Aktenplaneinheit" json:"filePlan"`
	ConfidentialityID    *string               `xml:"Vertraulichkeitsstufe>code" json:"-"`
	ConfidentialityLevel *ConfidentialityLevel `gorm:"foreignKey:ConfidentialityID;references:ID" xml:"-" json:"confidentialityLevel"`
	MediumID             *string               `xml:"Medium>code" json:"-"`
	Medium               *Medium               `gorm:"foreignKey:MediumID;references:ID" xml:"-" json:"medium"`
}

// BeforeDelete deletes associated rows of the deleted GeneralMetadata.
func (g *GeneralMetadata) BeforeDelete(tx *gorm.DB) (err error) {
	if g.ID == 0 {
		return fmt.Errorf("failed to delete associations for GeneralMetadata")
	}
	generalMetadata := GeneralMetadata{ID: g.ID}
	tx.Preload(clause.Associations).First(&generalMetadata)
	if generalMetadata.FilePlan != nil {
		tx.Delete(&generalMetadata.FilePlan)
	}
	return
}

type FilePlan struct {
	XMLName   xml.Name  `gorm:"-" xml:"Aktenplaneinheit" json:"-"`
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	XdomeaID  *string   `xml:"Kennzeichen" json:"xdomeaID"`
}

type Lifetime struct {
	XMLName   xml.Name  `gorm:"-" xml:"Laufzeit" json:"-"`
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	Start     *string   `xml:"Beginn" json:"start"`
	End       *string   `xml:"Ende" json:"end"`
}

type ArchiveMetadata struct {
	XMLName             xml.Name  `gorm:"-" xml:"ArchivspezifischeMetadaten" json:"-"`
	ID                  uint      `gorm:"primaryKey" json:"id"`
	CreatedAt           time.Time `json:"-"`
	UpdatedAt           time.Time `json:"-"`
	AppraisalCode       *string   `xml:"Aussonderungsart>Aussonderungsart>code" json:"appraisalCode"`
	AppraisalRecommCode *string   `xml:"Bewertungsvorschlag>code" json:"appraisalRecommCode"`
}

type ArchiveMetadataVersionIndependent struct {
	AppraisalCode       *string `xml:"Aussonderungsart>Aussonderungsart>code"`
	AppraisalCodePre300 *string `xml:"Aussonderungsart>code"`
	AppraisalRecommCode *string `xml:"Bewertungsvorschlag>code"`
}

// UnmarshalXML corrects version specific differences of the archive metadata.
func (archiveMetadata *ArchiveMetadata) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var temp ArchiveMetadataVersionIndependent
	err := d.DecodeElement(&temp, &start)
	if err != nil {
		return err
	}
	archiveMetadata.AppraisalRecommCode = temp.AppraisalRecommCode
	if temp.AppraisalCode != nil {
		archiveMetadata.AppraisalCode = temp.AppraisalCode
	} else if temp.AppraisalCodePre300 != nil {
		archiveMetadata.AppraisalCode = temp.AppraisalCodePre300
	}
	return nil
}
