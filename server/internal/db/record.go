package db

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type RecordType string

const (
	RecordTypeFile     RecordType = "file"
	RecordTypeProcess  RecordType = "process"
	RecordTypeDocument RecordType = "document"
)

type RootRecord struct {
	ProcessID        uuid.UUID `bson:"process_id"`
	MessageType      `bson:"message_type"`
	RecordType       `bson:"record_type"`
	ContainedRecords []uuid.UUID `bson:"contained_records"`
}

type RootFileRecord struct {
	RootRecord `bson:"inline"`
	FileRecord `bson:"inline"`
}

type RootProcessRecord struct {
	RootRecord    `bson:"inline"`
	ProcessRecord `bson:"inline"`
}

type RootDocumentRecord struct {
	RootRecord     `bson:"inline"`
	DocumentRecord `bson:"inline"`
}

type RootRecords struct {
	Files     []FileRecord     `json:"files"`
	Processes []ProcessRecord  `json:"processes"`
	Documents []DocumentRecord `json:"documents"`
}

type NestedRecords struct {
	Files     []*FileRecord
	Processes []*ProcessRecord
	Documents []*DocumentRecord
}

type FileRecord struct {
	RecordID        uuid.UUID        `bson:"record_id" json:"recordId"`
	GeneralMetadata *GeneralMetadata `bson:"general_metadata" json:"generalMetadata"`
	ArchiveMetadata *ArchiveMetadata `bson:"archive_metadata" json:"archiveMetadata"`
	Lifetime        *Lifetime        `gorm:"foreignKey:LifetimeID;references:ID;constraint:OnDelete:SET NULL" json:"lifetime"`
	Type            string           `json:"type"`
	Subfiles        []FileRecord     `json:"subfiles"`
	Processes       []ProcessRecord  `json:"processes"`
	Documents       []DocumentRecord `json:"documents"`
}

type ProcessRecord struct {
	RecordID        uuid.UUID        `xml:"Identifikation>ID" bson:"record_id" json:"recordId"`
	GeneralMetadata *GeneralMetadata `bson:"general_metadata" json:"generalMetadata"`
	ArchiveMetadata *ArchiveMetadata `bson:"archive_metadata" json:"archiveMetadata"`
	Lifetime        *Lifetime        `gorm:"foreignKey:LifetimeID;references:ID;constraint:OnDelete:SET NULL" json:"lifetime"`
	Type            string           `json:"type" xml:"Typ"`
	Subprocesses    []ProcessRecord  `xml:"Teilvorgang" json:"subprocesses"`
	Documents       []DocumentRecord `xml:"Dokument" json:"documents"`
}

type DocumentRecord struct {
	XMLName         xml.Name         `bson:"-" json:"-"`
	RecordID        uuid.UUID        `xml:"Identifikation>ID" bson:"record_id" json:"recordId"`
	GeneralMetadata *GeneralMetadata `bson:"general_metadata" json:"generalMetadata"`
	Type            string           `json:"type" xml:"Typ"`
	IncomingDate    string           `xml:"Posteingangsdatum" bson:"incoming_date" json:"incomingDate"`
	OutgoingDate    string           `xml:"Postausgangsdatum" bson:"outgoing_date" json:"outgoingDate"`
	DocumentDate    string           `xml:"DatumDesSchreibens" bson:"document_date" json:"documentDate"`
	Versions        []Version        `xml:"Version" json:"versions"`
	Attachments     []DocumentRecord `xml:"Anlage" json:"attachments"`
}

type GeneralMetadata struct {
	XMLName              xml.Name              `xml:"AllgemeineMetadaten" bson:"-"  json:"-"`
	Subject              string                `xml:"Betreff" json:"subject"`
	RecordNumber         string                `xml:"Kennzeichen" bson:"record_number" json:"recordNumber"`
	FilePlan             *FilePlan             `xml:"Aktenplaneinheit" bson:"file_plan"  json:"filePlan"`
	ConfidentialityLevel *ConfidentialityLevel `xml:"Vertraulichkeitsstufe>code" json:"confidentialityLevel"`
	Medium               *Medium               `xml:"Medium>code" json:"medium"`
}

type ConfidentialityLevel string

const (
	ConfidentialityLevel001 ConfidentialityLevel = "001"
	ConfidentialityLevel002 ConfidentialityLevel = "002"
	ConfidentialityLevel003 ConfidentialityLevel = "003"
	ConfidentialityLevel004 ConfidentialityLevel = "004"
	ConfidentialityLevel005 ConfidentialityLevel = "005"
)

type Medium string

const (
	Medium001 Medium = "001"
	Medium002 Medium = "002"
	Medium003 Medium = "003"
)

type FilePlan struct {
	XMLName        xml.Name `xml:"Aktenplaneinheit" bson:"-" json:"-"`
	FilePlanNumber string   `xml:"Kennzeichen" bson:"file_plan_number" json:"filePlanNumber"`
}

type Lifetime struct {
	XMLName xml.Name `xml:"Laufzeit" bson:"-" json:"-"`
	Start   string   `xml:"Beginn" json:"start"`
	End     string   `xml:"Ende" json:"end"`
}

type ArchiveMetadata struct {
	XMLName             xml.Name `xml:"ArchivspezifischeMetadaten" bson:"-" json:"-"`
	AppraisalCode       string   `xml:"Aussonderungsart>Aussonderungsart>code" bson:"appraisal_code" json:"appraisalCode"`
	AppraisalRecommCode string   `xml:"Bewertungsvorschlag>code" bson:"appraisal_recomm_code" json:"appraisalRecommCode"`
}

type Version struct {
	VersionID string   `xml:"Nummer" bson:"version_id" json:"versionID"`
	Formats   []Format `xml:"Format" json:"formats"`
}

type Format struct {
	Code            string          `xml:"Name>code" json:"code"`
	OtherName       string          `xml:"SonstigerName" bson:"other_name" json:"otherName"`
	Version         string          `xml:"Version" json:"version"`
	PrimaryDocument PrimaryDocument `xml:"Primaerdokument" bson:"primary_document" json:"primaryDocument"`
}

type PrimaryDocument struct {
	Filename         string `xml:"Dateiname" json:"filename"`
	FilenameOriginal string `xml:"DateinameOriginal" bson:"filename_original" json:"filenameOriginal"`
	CreatorName      string `xml:"Ersteller" bson:"creator_name" json:"creatorName"`
	CreationTime     string `xml:"DatumUhrzeit" bson:"creation_time" json:"creationTime"`
}

type fileRecordObjectVersionDifferences struct {
	RecordID        uuid.UUID        `xml:"Identifikation>ID"`
	GeneralMetadata *GeneralMetadata `xml:"AllgemeineMetadaten"`
	ArchiveMetadata *ArchiveMetadata `xml:"ArchivspezifischeMetadaten"`
	Lifetime        *Lifetime        `xml:"Laufzeit"`
	Type            string           `xml:"Typ"`
	Processes       []ProcessRecord  `xml:"Akteninhalt>Vorgang"`
	Documents       []DocumentRecord `xml:"Akteninhalt>Dokument"`
	Subfiles        []FileRecord     `xml:"Akteninhalt>Teilakte"`
	SubfilesPre300  []FileRecord     `xml:"Teilakte"`
}

type agencyIdentificationVersionIndependent struct {
	Code       string `xml:"Behoerdenschluessel>code"`
	CodePre300 string `xml:"Kennung>code"`
	Prefix     string `xml:"Praefix>code"`
}

type archiveMetadataVersionIndependent struct {
	AppraisalCode       string `xml:"Aussonderungsart>Aussonderungsart>code"`
	AppraisalCodePre300 string `xml:"Aussonderungsart>code"`
	AppraisalRecommCode string `xml:"Bewertungsvorschlag>code"`
}

// UnmarshalXML corrects version specific differences of file record objects.
func (f *FileRecord) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var temp fileRecordObjectVersionDifferences
	err := d.DecodeElement(&temp, &start)
	if err != nil {
		return err
	}
	f.RecordID = temp.RecordID
	f.GeneralMetadata = temp.GeneralMetadata
	f.ArchiveMetadata = temp.ArchiveMetadata
	f.Lifetime = temp.Lifetime
	f.Type = temp.Type
	f.Processes = temp.Processes
	f.Documents = temp.Documents
	if temp.Subfiles != nil {
		f.Subfiles = temp.Subfiles
	} else if temp.SubfilesPre300 != nil {
		f.Subfiles = temp.SubfilesPre300
	}
	return nil
}

// UnmarshalXML corrects version specific differences of the agency identification.
func (agencyIdentification *AgencyIdentification) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var temp agencyIdentificationVersionIndependent
	err := d.DecodeElement(&temp, &start)
	if err != nil {
		return err
	}
	agencyIdentification.Prefix = temp.Prefix
	if temp.Code != "" {
		agencyIdentification.Code = temp.Code
	} else if temp.CodePre300 != "" {
		agencyIdentification.Code = temp.CodePre300
	}
	return nil
}

// UnmarshalXML corrects version specific differences of the archive metadata.
func (archiveMetadata *ArchiveMetadata) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var temp archiveMetadataVersionIndependent
	err := d.DecodeElement(&temp, &start)
	if err != nil {
		return err
	}
	archiveMetadata.AppraisalRecommCode = temp.AppraisalRecommCode
	if temp.AppraisalCode != "" {
		archiveMetadata.AppraisalCode = temp.AppraisalCode
	} else if temp.AppraisalCodePre300 != "" {
		archiveMetadata.AppraisalCode = temp.AppraisalCodePre300
	}
	return nil
}

// InsertRootRecords inserts records for files, processes, and documents for a
// given message into the database.
func InsertRootRecords(processID uuid.UUID, messageType MessageType, records RootRecords) {
	// Check for existing records
	coll := mongoDatabase.Collection("root_records")
	filter := bson.D{
		{"process_id", processID},
		{"message_type", messageType},
	}
	result := coll.FindOne(context.Background(), filter)
	if result.Err() != mongo.ErrNoDocuments {
		panic(fmt.Errorf("existing root records for %s message of submission process %s",
			messageType, processID))
	}
	// Create root record objects
	rootRecords := make([]interface{}, len(records.Files)+len(records.Processes)+len(records.Documents))
	for i, f := range records.Files {
		n := ExtractNestedRecords(&RootRecords{Files: []FileRecord{f}})
		rootRecords[i] = RootFileRecord{
			RootRecord: RootRecord{
				ProcessID:        processID,
				MessageType:      messageType,
				RecordType:       RecordTypeFile,
				ContainedRecords: getContainedRecordIds(n),
			},
			FileRecord: f,
		}
	}
	offset := len(records.Files)
	for i, p := range records.Processes {
		n := ExtractNestedRecords(&RootRecords{Processes: []ProcessRecord{p}})
		rootRecords[i+offset] = RootProcessRecord{
			RootRecord: RootRecord{
				ProcessID:        processID,
				MessageType:      messageType,
				RecordType:       RecordTypeProcess,
				ContainedRecords: getContainedRecordIds(n),
			},
			ProcessRecord: p,
		}
	}
	offset += len(records.Processes)
	for i, d := range records.Documents {
		n := ExtractNestedRecords(&RootRecords{Documents: []DocumentRecord{d}})
		rootRecords[i+offset] = RootDocumentRecord{
			RootRecord: RootRecord{
				ProcessID:        processID,
				MessageType:      messageType,
				RecordType:       RecordTypeDocument,
				ContainedRecords: getContainedRecordIds(n),
			},
			DocumentRecord: d,
		}
	}
	// Insert root records
	_, err := coll.InsertMany(context.Background(), rootRecords)
	if err != nil {
		panic(err)
	}
}

// FindRootRecords queries all root records of a given message and returns them
// as an array per record type.
func FindRootRecords(
	ctx context.Context,
	processID uuid.UUID,
	messageType MessageType,
) RootRecords {
	coll := mongoDatabase.Collection("root_records")
	// Find file records
	filter := bson.D{
		{"process_id", processID},
		{"message_type", messageType},
		{"record_type", RecordTypeFile},
	}
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		panic(err)
	}
	var r RootRecords
	err = cursor.All(ctx, &r.Files)
	if err != nil {
		panic(err)
	}
	// Find process records
	filter = bson.D{
		{"process_id", processID},
		{"message_type", messageType},
		{"record_type", RecordTypeProcess},
	}
	cursor, err = coll.Find(ctx, filter)
	if err != nil {
		panic(err)
	}
	err = cursor.All(ctx, &r.Processes)
	if err != nil {
		panic(err)
	}
	// Find document records
	filter = bson.D{
		{"process_id", processID},
		{"message_type", messageType},
		{"record_type", RecordTypeDocument},
	}
	cursor, err = coll.Find(ctx, filter)
	if err != nil {
		panic(err)
	}
	err = cursor.All(ctx, &r.Documents)
	if err != nil {
		panic(err)
	}
	return r
}

// FindRootRecord finds the root record that *contains* the given record, either
// as the root record itself or as nested record.
func FindRootRecord(
	ctx context.Context,
	processID uuid.UUID,
	messageType MessageType,
	recordID uuid.UUID,
) (RootRecords, bool) {
	coll := mongoDatabase.Collection("root_records")
	filter := bson.D{
		{"$and",
			bson.A{
				bson.D{{"process_id", processID}},
				bson.D{{"message_type", messageType}},
				bson.D{{"contained_records", bson.D{{"$all", bson.A{recordID}}}}},
			}},
	}
	var r RootRecords
	result := coll.FindOne(ctx, filter)
	raw, err := result.Raw()
	if err == mongo.ErrNoDocuments {
		return r, false
	} else if err != nil {
		panic(err)
	}
	recordType := raw.Lookup("record_type").StringValue()
	switch RecordType(recordType) {
	case RecordTypeFile:
		r.Files = make([]FileRecord, 1)
		err = result.Decode(&r.Files[0])
	case RecordTypeProcess:
		r.Processes = make([]ProcessRecord, 1)
		err = result.Decode(&r.Processes[0])
	case RecordTypeDocument:
		r.Documents = make([]DocumentRecord, 1)
		err = result.Decode(&r.Documents[0])
	default:
		panic("unknown record type: " + recordType)
	}
	if err != nil {
		panic(err)
	}
	return r, true
}

func DeleteRecordsForProcess(processID uuid.UUID) {
	coll := mongoDatabase.Collection("root_records")
	filter := bson.D{
		{"process_id", processID},
	}
	_, err := coll.DeleteMany(context.Background(), filter)
	if err != nil {
		panic(err)
	}
}

func DeleteRecordsForMessage(processID uuid.UUID, messageType MessageType) {
	coll := mongoDatabase.Collection("root_records")
	filter := bson.D{
		{"process_id", processID},
		{"message_type", messageType},
	}
	_, err := coll.DeleteMany(context.Background(), filter)
	if err != nil {
		panic(err)
	}
}

// ExtractNestedRecords returns all nested records from the given root records in
// a flat structure. It additionally keeps the nested child records.
func ExtractNestedRecords(r *RootRecords) NestedRecords {
	var n NestedRecords
	var appendDocumentRecords func(documents []DocumentRecord)
	appendDocumentRecords = func(documents []DocumentRecord) {
		for _, d := range documents {
			n.Documents = append(n.Documents, &d)
			appendDocumentRecords(d.Attachments)
		}
	}
	appendDocumentRecords(r.Documents)
	var appendProcessRecords func(processes []ProcessRecord)
	appendProcessRecords = func(processes []ProcessRecord) {
		for _, p := range processes {
			n.Processes = append(n.Processes, &p)
			appendProcessRecords(p.Subprocesses)
			appendDocumentRecords(p.Documents)
		}
	}
	appendProcessRecords(r.Processes)
	var appendFileRecords func(files []FileRecord)
	appendFileRecords = func(files []FileRecord) {
		for _, f := range files {
			n.Files = append(n.Files, &f)
			appendFileRecords(f.Subfiles)
			appendProcessRecords(f.Processes)
			appendDocumentRecords(f.Documents)
		}
	}
	appendFileRecords(r.Files)
	return n
}

// getContainedRecordIds returns the IDs of all nested records contained in the
// given root record.
func getContainedRecordIds(n NestedRecords) []uuid.UUID {
	ids := make([]uuid.UUID, len(n.Files)+len(n.Processes)+len(n.Documents))
	for i, f := range n.Files {
		ids[i] = f.RecordID
	}
	offset := len(n.Files)
	for i, p := range n.Processes {
		ids[i+offset] = p.RecordID
	}
	offset += len(n.Processes)
	for i, d := range n.Documents {
		ids[i+offset] = d.RecordID
	}
	return ids
}

func GetPrimaryDocuments(r *RootRecords) []PrimaryDocument {
	var d []PrimaryDocument
	for _, c := range r.Files {
		d = append(d, GetPrimaryDocumentsForFile(&c)...)
	}
	for _, c := range r.Processes {
		d = append(d, GetPrimaryDocumentsForProcess(&c)...)
	}
	for _, c := range r.Documents {
		d = append(d, GetPrimaryDocumentsForDocument(&c)...)
	}
	return d
}

func GetPrimaryDocumentsForFile(r *FileRecord) []PrimaryDocument {
	var d []PrimaryDocument
	for _, c := range r.Subfiles {
		d = append(d, GetPrimaryDocumentsForFile(&c)...)
	}
	for _, c := range r.Processes {
		d = append(d, GetPrimaryDocumentsForProcess(&c)...)
	}
	for _, c := range r.Documents {
		d = append(d, GetPrimaryDocumentsForDocument(&c)...)
	}
	return d
}

func GetPrimaryDocumentsForProcess(r *ProcessRecord) []PrimaryDocument {
	var d []PrimaryDocument
	for _, c := range r.Subprocesses {
		d = append(d, GetPrimaryDocumentsForProcess(&c)...)
	}
	for _, c := range r.Documents {
		d = append(d, GetPrimaryDocumentsForDocument(&c)...)
	}
	return d
}

func GetPrimaryDocumentsForDocument(r *DocumentRecord) []PrimaryDocument {
	var d []PrimaryDocument
	for _, version := range r.Versions {
		for _, format := range version.Formats {
			d = append(d, format.PrimaryDocument)
		}
	}
	for _, c := range r.Attachments {
		d = append(d, GetPrimaryDocumentsForDocument(&c)...)
	}
	return d
}
