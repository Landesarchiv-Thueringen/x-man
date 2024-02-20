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
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Code      string         `json:"code"`
	URI       string         `json:"uri"`
	XSDPath   string         `json:""`
}

func (v *XdomeaVersion) IsVersionPriorTo300() bool {
	return v.Code == "2.3.0" || v.Code == "2.4.0"
}

type Process struct {
	ID               uuid.UUID         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	CreatedAt        time.Time         `json:"receivedAt"`
	UpdatedAt        time.Time         `json:"-"`
	DeletedAt        gorm.DeletedAt    `gorm:"index" json:"-"`
	XdomeaID         string            `json:"xdomeaID"`
	AgencyID         uint              `json:"-"`
	Agency           Agency            `gorm:"foreignKey:AgencyID;references:ID" json:"agency"`
	StoreDir         string            `json:"-"`
	Institution      *string           `json:"institution"`
	Note             *string           `json:"note"`
	Message0501ID    *uuid.UUID        `json:"message0501Id"`
	Message0501      *Message          `gorm:"foreignKey:Message0501ID;references:ID" json:"message0501"`
	Message0502Path  *string           `json:"-"`
	Message0503ID    *uuid.UUID        `json:"message0503Id"`
	Message0503      *Message          `gorm:"foreignKey:Message0503ID;references:ID" json:"message0503"`
	Message0504Path  *string           `json:"-"`
	Message0505ID    *uuid.UUID        `json:"message0505Id"`
	Message0505      *Message          `gorm:"foreignKey:Message0505ID;references:ID" json:"message0505"`
	ProcessingErrors []ProcessingError `json:"processingErrors"`
	ProcessStateID   uint              `json:"-"`
	ProcessState     ProcessState      `gorm:"foreignKey:ProcessStateID;references:ID" json:"processState"`
	Tasks            []Task            `json:"-"`
}

// BeforeDelete deletes associated rows of the deleted Process.
func (p *Process) BeforeDelete(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		return fmt.Errorf("failed to delete associations for Process")
	}
	process := Process{ID: p.ID}
	tx.Preload(clause.Associations).First(&process)
	if process.Message0501 != nil {
		tx.Delete(&process.Message0501)
	}
	if process.Message0503 != nil {
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
	ID        uint           `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	// Complete is true if the step completed successfully.
	Complete bool `gorm:"default:false" json:"complete"`
	// CompletionTime is the time at which Complete was set to true.
	CompletionTime *time.Time `json:"completionTime"`
	// Tasks are all tasks associated with the process step.
	//
	// A task of the state "running" indicates that the process step is
	// currently in progress. Tasks of the state "failed" do *not* necessarily
	// indicate a failed process step (see ProcessingErrors). Also, a task of
	// the state "succeeded" is *not* a requirement for a complete process step.
	//
	// Not all process steps (completed or not) have tasks.
	Tasks []Task `json:"tasks"`
	// ProcessingErrors are all processing errors associated with the process
	// step.
	//
	// An unresolved processing error indicates that the process step is
	// currently in a failed state.
	ProcessingErrors []ProcessingError `json:"processingErrors"`
}

type Message struct {
	ID                     uuid.UUID              `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	CreatedAt              time.Time              `json:"-"`
	UpdatedAt              time.Time              `json:"-"`
	DeletedAt              gorm.DeletedAt         `gorm:"index" json:"-"`
	TransferDir            string                 `json:"-"`
	TransferDirMessagePath string                 `json:"-"`
	StoreDir               string                 `json:"-"`
	MessagePath            string                 `json:"-"`
	XdomeaVersion          string                 `json:"xdomeaVersion"`
	MessageHeadID          *uint                  `json:"-"`
	MessageHead            MessageHead            `gorm:"foreignKey:MessageHeadID;references:ID" json:"messageHead"`
	MessageTypeID          *uint                  `json:"-"`
	MessageType            MessageType            `gorm:"foreignKey:MessageTypeID;references:ID" json:"messageType"`
	AppraisalComplete      bool                   `json:"appraisalComplete"`
	SchemaValidation       bool                   `gorm:"default:true" json:"schemaValidation"`
	FileRecordObjects      []FileRecordObject     `gorm:"many2many:message_file_record_objects;" json:"fileRecordObjects"`
	ProcessRecordObjects   []ProcessRecordObject  `gorm:"many2many:message_process_record_objects;" json:"processRecordObjects"`
	DocumentRecordObjects  []DocumentRecordObject `gorm:"many2many:message_document_record_objects;" json:"documentRecordObjects"`
	ProcessingErrors       []ProcessingError      `json:"processingErrors"`
}

type MessageType struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Code      string         `json:"code"`
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
	Code      *string        `xml:"Behoerdenschluessel>code" json:"code"`
	Prefix    *string        `xml:"Praefix>code" json:"prefix"`
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
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"-"`
	UpdatedAt    time.Time      `json:"-"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Name         *string        `xml:"Name"  json:"name"`
	Abbreviation *string        `xml:"Kurzbezeichnung" json:"abbreviation"`
}

type FileRecordObject struct {
	ID                    uuid.UUID              `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	XdomeaID              uuid.UUID              `json:"xdomeaID"`
	CreatedAt             time.Time              `json:"-"`
	UpdatedAt             time.Time              `json:"-"`
	DeletedAt             gorm.DeletedAt         `gorm:"index" json:"-"`
	MessageID             uuid.UUID              `json:"messageID"`
	RecordObjectType      string                 `gorm:"default:file" json:"recordObjectType"`
	GeneralMetadataID     *uint                  `json:"-"`
	GeneralMetadata       *GeneralMetadata       `gorm:"foreignKey:GeneralMetadataID;references:ID" json:"generalMetadata"`
	ArchiveMetadataID     *uint                  `json:"-"`
	ArchiveMetadata       *ArchiveMetadata       `gorm:"foreignKey:ArchiveMetadataID;references:ID" json:"archiveMetadata"`
	LifetimeID            *uint                  `json:"-"`
	Lifetime              *Lifetime              `gorm:"foreignKey:LifetimeID;references:ID" json:"lifetime"`
	Type                  *string                `json:"type"`
	ProcessRecordObjects  []ProcessRecordObject  `gorm:"many2many:file_processes;" json:"processes"`
	SubFileRecordObjects  []FileRecordObject     `gorm:"many2many:file_subfiles;" json:"subfiles"`
	DocumentRecordObjects []DocumentRecordObject `gorm:"many2many:file_documents;" json:"documents"`
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
	DeletedAt               gorm.DeletedAt         `gorm:"index" json:"-"`
	MessageID               uuid.UUID              `json:"messageID"`
	RecordObjectType        string                 `gorm:"default:process" json:"recordObjectType"`
	GeneralMetadataID       *uint                  `json:"-"`
	GeneralMetadata         *GeneralMetadata       `gorm:"foreignKey:GeneralMetadataID;references:ID" xml:"AllgemeineMetadaten" json:"generalMetadata"`
	ArchiveMetadataID       *uint                  `json:"-"`
	ArchiveMetadata         *ArchiveMetadata       `gorm:"foreignKey:ArchiveMetadataID;references:ID" xml:"ArchivspezifischeMetadaten" json:"archiveMetadata"`
	LifetimeID              *uint                  `json:"-"`
	Lifetime                *Lifetime              `gorm:"foreignKey:LifetimeID;references:ID" json:"lifetime"`
	Type                    *string                `json:"type" xml:"Typ"`
	SubProcessRecordObjects []ProcessRecordObject  `gorm:"many2many:process_subprocesses;" xml:"Teilvorgang" json:"subprocesses"`
	DocumentRecordObjects   []DocumentRecordObject `gorm:"many2many:process_documents;" xml:"Dokument" json:"documents"`
}

type DocumentRecordObject struct {
	XMLName           xml.Name               `gorm:"-" json:"-"`
	ID                uuid.UUID              `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	XdomeaID          uuid.UUID              `xml:"Identifikation>ID" json:"xdomeaID"`
	CreatedAt         time.Time              `json:"-"`
	UpdatedAt         time.Time              `json:"-"`
	DeletedAt         gorm.DeletedAt         `gorm:"index" json:"-"`
	MessageID         uuid.UUID              `json:"messageID"`
	RecordObjectType  string                 `gorm:"default:document" json:"recordObjectType"`
	GeneralMetadataID *uint                  `json:"-"`
	GeneralMetadata   *GeneralMetadata       `gorm:"foreignKey:GeneralMetadataID;references:ID" xml:"AllgemeineMetadaten" json:"generalMetadata"`
	Type              *string                `json:"type" xml:"Typ"`
	IncomingDate      *string                `xml:"Posteingangsdatum" json:"incomingDate"`
	OutgoingDate      *string                `xml:"Postausgangsdatum" json:"outgoingDate"`
	DocumentDate      *string                `xml:"DatumDesSchreibens" json:"documentDate"`
	Versions          []Version              `gorm:"many2many:document_versions;" xml:"Version" json:"versions"`
	Attachments       []DocumentRecordObject `gorm:"many2many:document_attachments;" xml:"Anlage" json:"attachments"`
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
	XMLName              xml.Name              `gorm:"-" xml:"AllgemeineMetadaten" json:"-"`
	ID                   uint                  `gorm:"primaryKey" json:"id"`
	CreatedAt            time.Time             `json:"-"`
	UpdatedAt            time.Time             `json:"-"`
	DeletedAt            gorm.DeletedAt        `gorm:"index" json:"-"`
	Subject              *string               `xml:"Betreff" json:"subject"`
	XdomeaID             *string               `xml:"Kennzeichen" json:"xdomeaID"`
	FilePlanID           *uint                 `json:"-"`
	FilePlan             *FilePlan             `gorm:"foreignKey:FilePlanID;references:ID" xml:"Aktenplaneinheit" json:"filePlan"`
	ConfidentialityID    *string               `xml:"Vertraulichkeitsstufe>code" json:"-"`
	ConfidentialityLevel *ConfidentialityLevel `gorm:"foreignKey:ConfidentialityID;references:ID" xml:"-" json:"confidentialityLevel"`
	MediumID             *string               `xml:"Medium>code" json:"-"`
	Medium               *Medium               `gorm:"foreignKey:MediumID;references:ID" xml:"-" json:"medium"`
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
	XMLName               xml.Name       `gorm:"-" xml:"ArchivspezifischeMetadaten" json:"-"`
	ID                    uint           `gorm:"primaryKey" json:"id"`
	CreatedAt             time.Time      `json:"-"`
	UpdatedAt             time.Time      `json:"-"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
	AppraisalCode         *string        `xml:"Aussonderungsart>Aussonderungsart>code" json:"appraisalCode"`
	AppraisalRecommCode   *string        `xml:"Bewertungsvorschlag>code" json:"appraisalRecommCode"`
	InternalAppraisalNote *string        `json:"internalAppraisalNote"`
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
