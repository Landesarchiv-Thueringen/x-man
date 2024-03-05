package db

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetProcessingErrors() []ProcessingError {
	var processingErrors []ProcessingError
	result := db.
		Preload(clause.Associations).
		Preload("Message." + clause.Associations).
		Find(&processingErrors)
	if result.Error != nil {
		panic(result.Error)
	}
	return processingErrors
}

func GetAgencies() []Agency {
	var agencies []Agency
	result := db.Preload(clause.Associations).Find(&agencies)
	if result.Error != nil {
		panic(result.Error)
	}
	return agencies
}

func GetAgenciesForUser(userID []byte) []Agency {
	if len(userID) == 0 {
		panic("called GetAgenciesForUser with empty user ID")
	}
	var agencies []Agency
	result := db.
		Preload(clause.Associations).
		Where("id IN (SELECT agency_id FROM agency_users WHERE user_id = ?)", userID).
		Find(&agencies)
	if result.Error != nil {
		panic(result.Error)
	}
	return agencies
}

func GetAgenciesForCollection(collectionID uint) []Agency {
	var agencies []Agency
	result := db.
		Preload(clause.Associations).
		Where("collection_id = ?", collectionID).
		Find(&agencies)
	if result.Error != nil {
		panic(result.Error)
	}
	return agencies
}

func GetCollections() []Collection {
	var collections []Collection
	result := db.Preload(clause.Associations).Find(&collections)
	if result.Error != nil {
		panic(result.Error)
	}
	return collections
}

func GetTasks() []Task {
	var tasks []Task
	result := db.Preload(clause.Associations).Find(&tasks)
	if result.Error != nil {
		panic(result.Error)
	}
	return tasks
}

func GetSupportedXdomeaVersions() []XdomeaVersion {
	var xdomeaVersions []XdomeaVersion
	result := db.Find(&xdomeaVersions)
	if result.Error != nil {
		panic(result.Error)
	}
	return xdomeaVersions
}

func GetXdomeaVersionByCode(code string) (XdomeaVersion, error) {
	xdomeaVersion := XdomeaVersion{
		Code: code,
	}
	result := db.Where(&xdomeaVersion).First(&xdomeaVersion)
	return xdomeaVersion, result.Error
}

func GetProcesses() []Process {
	var processes []Process
	result := db.
		Preload("Agency").
		Preload("Message0501.MessageHead").
		Preload("Message0501.MessageType").
		Preload("Message0503.MessageHead").
		Preload("Message0503.MessageType").
		Preload("ProcessingErrors").
		Preload("ProcessingErrors.Agency").
		Preload("ProcessState.Receive0501." + clause.Associations).
		Preload("ProcessState.Appraisal." + clause.Associations).
		Preload("ProcessState.Receive0505." + clause.Associations).
		Preload("ProcessState.Receive0503." + clause.Associations).
		Preload("ProcessState.FormatVerification." + clause.Associations).
		Preload("ProcessState.Archiving." + clause.Associations).
		Find(&processes)
	if result.Error != nil {
		panic(result.Error)
	}
	return processes
}

func GetProcessesForUser(userID []byte) []Process {
	var processes []Process
	agencies := GetAgenciesForUser(userID)
	agencyIDs := make([]uint, len(agencies))
	for i, v := range agencies {
		agencyIDs[i] = v.ID
	}
	result := db.
		Where("agency_id IN ?", agencyIDs).
		Preload("Agency").
		Preload("Message0501.MessageHead").
		Preload("Message0501.MessageType").
		Preload("Message0503.MessageHead").
		Preload("Message0503.MessageType").
		Preload("ProcessingErrors").
		Preload("ProcessingErrors.Agency").
		Preload("ProcessState.Receive0501." + clause.Associations).
		Preload("ProcessState.Appraisal." + clause.Associations).
		Preload("ProcessState.Receive0505." + clause.Associations).
		Preload("ProcessState.Receive0503." + clause.Associations).
		Preload("ProcessState.FormatVerification." + clause.Associations).
		Preload("ProcessState.Archiving." + clause.Associations).
		Find(&processes)
	if result.Error != nil {
		panic(result.Error)
	}
	return processes
}

func GetMessageByID(id uuid.UUID) (Message, error) {
	var message Message
	result := db.First(&message, id)
	return message, result.Error
}

func GetCompleteMessageByID(id uuid.UUID) (Message, bool) {
	var message Message
	result := db.
		Preload(clause.Associations).
		Preload("MessageHead.Sender."+clause.Associations).
		Preload("MessageHead.Sender.AgencyIdentification."+clause.Associations).
		Preload("MessageHead.Receiver."+clause.Associations).
		Preload("MessageHead.Receiver.AgencyIdentification."+clause.Associations).
		Scopes(PreloadRecordObjects).
		Limit(1).
		Find(&message, id)
	if result.Error != nil {
		panic(result.Error)
	}
	return message, result.RowsAffected > 0
}

// GetProcessForMessage returns the process to which the given message belongs.
func GetProcessForMessage(message Message) Process {
	if message.ID == uuid.Nil {
		panic("called GetProcessForMessage with nil message.ID")
	}
	var process = Process{}
	result := db.Where("? in (message0501_id, message0503_id, message0505_id)", message.ID).First(&process)
	if result.Error != nil {
		panic(result.Error)
	}
	return process
}

// IsMessageAlreadyProcessed checks if a message exists, which was already processed,
// determined by the path in the transfer directory.
func IsMessageAlreadyProcessed(path string) bool {
	result := db.Where(&Message{TransferDirMessagePath: path}).Limit(1).Find(&Message{})
	if result.RowsAffected > 0 {
		return true
	}
	result = db.Where(&ProcessingError{TransferPath: &path, Resolved: false}).Limit(1).Find(&ProcessingError{})
	return result.RowsAffected > 0
}

func GetMessageTypeCode(id uuid.UUID) (string, error) {
	var message Message
	result := db.
		Preload("MessageType").
		First(&message, id)
	if result.Error != nil {
		return "", result.Error
	}
	return message.MessageType.Code, nil
}

func IsMessageAppraisalComplete(id uuid.UUID) (bool, error) {
	message, err := GetMessageByID(id)
	if err != nil {
		return false, err
	}
	return message.AppraisalComplete, err
}

func GetFileRecordObjectByID(id uuid.UUID) (FileRecordObject, error) {
	var file FileRecordObject
	result := db.
		Scopes(PreloadFileRecordObject("", 0, 0)).
		First(&file, id)
	return file, result.Error
}

func GetProcessRecordObjectByID(id uuid.UUID) (ProcessRecordObject, bool) {
	if id == uuid.Nil {
		panic("called GetProcessRecordObjectByID with ID nil")
	}
	var process ProcessRecordObject
	result := db.
		Scopes(PreloadProcessRecordObject("", 0, 0)).
		Limit(1).Find(&process, id)
	if result.Error != nil {
		panic(result.Error)
	}
	return process, result.RowsAffected == 1
}

func GetDocumentRecordObjectByID(id uuid.UUID) (DocumentRecordObject, error) {
	var document DocumentRecordObject
	result := db.
		Scopes(PreloadDocumentRecordObject("")).
		First(&document, id)
	return document, result.Error
}

func GetAllFileRecordObjects(messageID uuid.UUID) (map[uuid.UUID]FileRecordObject, error) {
	var fileRecordObjects []FileRecordObject
	result := db.
		Scopes(PreloadFileRecordObject("", 0, 0)).
		Where("message_id = ?", messageID.String()).
		Find(&fileRecordObjects)
	fileIndex := make(map[uuid.UUID]FileRecordObject)
	for _, f := range fileRecordObjects {
		fileIndex[f.XdomeaID] = f
	}
	return fileIndex, result.Error
}

func GetAllProcessRecordObjects(messageID uuid.UUID) (map[uuid.UUID]ProcessRecordObject, error) {
	var processRecordObjects []ProcessRecordObject
	result := db.
		Scopes(PreloadProcessRecordObject("", 0, 0)).
		Where("message_id = ?", messageID.String()).
		Find(&processRecordObjects)
	processIndex := make(map[uuid.UUID]ProcessRecordObject)
	for _, p := range processRecordObjects {
		processIndex[p.XdomeaID] = p
	}
	return processIndex, result.Error
}

func GetAllDocumentRecordObjects(messageID uuid.UUID) map[uuid.UUID]DocumentRecordObject {
	var documentRecordObjects []DocumentRecordObject
	result := db.
		Scopes(PreloadDocumentRecordObject("")).
		Where("message_id = ?", messageID.String()).
		Find(&documentRecordObjects)
	documentIndex := make(map[uuid.UUID]DocumentRecordObject)
	for _, d := range documentRecordObjects {
		documentIndex[d.XdomeaID] = d
	}
	if result.Error != nil {
		panic(result.Error)
	}
	return documentIndex
}

func GetAllPrimaryDocuments(messageID uuid.UUID) []PrimaryDocument {
	var primaryDocuments []PrimaryDocument
	var documents []DocumentRecordObject
	result := db.
		Preload("Versions.Formats.PrimaryDocument").
		Where("message_id = ?", messageID.String()).
		Find(&documents)
	if result.Error != nil {
		panic(result.Error)
	}
	for _, document := range documents {
		if document.Versions != nil {
			for _, version := range document.Versions {
				for _, format := range version.Formats {
					primaryDocuments = append(primaryDocuments, format.PrimaryDocument)
				}
			}
		}
	}
	return primaryDocuments
}

func GetAllPrimaryDocumentsWithFormatVerification(messageID uuid.UUID) []PrimaryDocument {
	var primaryDocuments []PrimaryDocument
	var documents []DocumentRecordObject
	result := db.
		Preload("Versions.Formats.PrimaryDocument.FormatVerification.Features.Values.Tools").
		Preload("Versions.Formats.PrimaryDocument.FormatVerification.FileIdentificationResults.Features").
		Preload("Versions.Formats.PrimaryDocument.FormatVerification.FileValidationResults.Features").
		Where("message_id = ?", messageID.String()).
		Find(&documents)
	if result.Error != nil {
		panic(result.Error)
	}
	for _, document := range documents {
		if document.Versions != nil {
			for _, version := range document.Versions {
				for _, format := range version.Formats {
					primaryDocuments = append(primaryDocuments, format.PrimaryDocument)
				}
			}
		}
	}
	for primaryDocumentIndex, primaryDocument := range primaryDocuments {
		if primaryDocument.FormatVerification == nil {
			continue
		}
		if len(primaryDocument.FormatVerification.Features) > 0 {
			summary := make(map[string]Feature)
			for _, feature := range primaryDocument.FormatVerification.Features {
				summary[feature.Key] = feature
			}
			primaryDocuments[primaryDocumentIndex].FormatVerification.Summary = summary
		}
		if len(primaryDocument.FormatVerification.FileIdentificationResults) > 0 {
			for toolID, tool := range primaryDocument.FormatVerification.FileIdentificationResults {
				features := make(map[string]string)
				for _, feature := range tool.Features {
					features[feature.Key] = feature.Value
				}
				primaryDocuments[primaryDocumentIndex].FormatVerification.
					FileIdentificationResults[toolID].ExtractedFeatures = &features
			}
		}
		if len(primaryDocument.FormatVerification.FileValidationResults) > 0 {
			for toolID, tool := range primaryDocument.FormatVerification.FileValidationResults {
				features := make(map[string]string)
				for _, feature := range tool.Features {
					features[feature.Key] = feature.Value
				}
				primaryDocuments[primaryDocumentIndex].FormatVerification.
					FileValidationResults[toolID].ExtractedFeatures = &features
			}
		}
	}
	return primaryDocuments
}

func GetMessageTypeByCode(code string) MessageType {
	messageType := MessageType{Code: code}
	result := db.Where(&messageType).First(&messageType)
	if result.Error != nil {
		panic(result.Error)
	}
	return messageType
}

func GetRecordObjectAppraisals() []RecordObjectAppraisal {
	var appraisals []RecordObjectAppraisal
	result := db.Find(&appraisals)
	if result.Error != nil {
		panic(result.Error)
	}
	return appraisals
}

func GetConfidentialityLevelCodelist() []ConfidentialityLevel {
	var codelist []ConfidentialityLevel
	result := db.Find(&codelist)
	if result.Error != nil {
		panic(result.Error)
	}
	return codelist
}

// GetAllTransferFilesOfProcess returns the transfer paths of all messages that
// belong to the given process.
func GetAllTransferFilesOfProcess(process Process) []string {
	p := Process{ID: process.ID}
	messages := make([]string, 0)
	result := db.Model(&Process{}).Preload(clause.Associations).First(&p)
	if result.Error != nil {
		panic(result.Error)
	}
	if p.Message0501 != nil {
		messages = append(messages, p.Message0501.TransferDirMessagePath)
	}
	if p.Message0502Path != nil {
		messages = append(messages, *p.Message0502Path)
	}
	if p.Message0503 != nil {
		messages = append(messages, p.Message0503.TransferDirMessagePath)
	}
	if p.Message0504Path != nil {
		messages = append(messages, *p.Message0504Path)
	}
	if p.Message0505 != nil {
		messages = append(messages, p.Message0505.TransferDirMessagePath)
	}
	return messages
}

func GetMessageOfProcessByCode(process Process, code string) (Message, error) {
	result := db.Model(&Process{}).
		Preload("Message0501.MessageType").
		Preload("Message0503.MessageType").
		Where(&process).
		First(&process)
	if result.Error != nil {
		return Message{}, fmt.Errorf("failed to read process {%s}: %w", process.ID, result.Error)
	}
	switch code {
	case "0501":
		if process.Message0501 == nil {
			return Message{}, errors.New("process {" + process.ID + "} has no 0501 message")
		} else {
			return *process.Message0501, nil
		}
	case "0503":
		if process.Message0503 == nil {
			return Message{}, errors.New("process {" + process.ID + "} has no 0503 message")
		} else {
			return *process.Message0503, nil
		}
	case "0505":
		if process.Message0505 == nil {
			return Message{}, errors.New("process {" + process.ID + "} has no 0505 message")
		} else {
			return *process.Message0505, nil
		}
	default:
		panic("unsupported message type with code: " + code)
	}
}

func GetMessagesByCode(code string) []Message {
	var messages []Message
	messageType := GetMessageTypeByCode(code)
	result := db.Model(&Message{}).
		Preload("MessageType").
		Preload("MessageHead.Sender.Institution").
		Preload("MessageHead.Sender.AgencyIdentification.Code").
		Preload("MessageHead.Sender.AgencyIdentification.Prefix").
		Preload("MessageHead.Receiver.Institution").
		Preload("MessageHead.Receiver.AgencyIdentification.Code").
		Preload("MessageHead.Receiver.AgencyIdentification.Prefix").
		Where("message_type_id = ?", messageType.ID).
		Find(&messages)
	if result.Error != nil {
		panic(result.Error)
	}
	return messages
}

func GetProcessStep(ID uint) (ProcessStep, error) {
	processStep := ProcessStep{ID: ID}
	result := db.First(&processStep)
	return processStep, result.Error
}

func GetProcess(processID string) (Process, bool) {
	if processID == "" {
		panic("called GetProcess with empty string")
	}
	process := Process{ID: processID}
	result := db.Model(&Process{}).
		Preload("Agency").
		Preload("Message0501.MessageHead").
		Preload("Message0501.MessageType").
		Preload("Message0503.MessageHead").
		Preload("Message0503.MessageType").
		Preload("ProcessingErrors").
		Preload("ProcessState.Receive0501." + clause.Associations).
		Preload("ProcessState.Appraisal." + clause.Associations).
		Preload("ProcessState.Receive0505." + clause.Associations).
		Preload("ProcessState.Receive0503." + clause.Associations).
		Preload("ProcessState.FormatVerification." + clause.Associations).
		Preload("ProcessState.Archiving." + clause.Associations).
		Where(&process).Limit(1).Find(&process)
	return process, result.RowsAffected == 1
}

func GetAppraisalByCode(code string) (RecordObjectAppraisal, bool) {
	if code == "" {
		panic("called GetAppraisalByCode with empty string")
	}
	appraisal := RecordObjectAppraisal{Code: code}
	result := db.Where(&appraisal).Limit(1).Find(&appraisal)
	if result.Error != nil {
		panic(result.Error)
	}
	return appraisal, result.RowsAffected == 1
}

func GetPrimaryFileStorePath(messageID uuid.UUID, primaryDocumentID uint) (string, error) {
	var message Message
	result := db.
		Preload("MessageType").
		First(&message, messageID)
	if result.Error != nil {
		return "", result.Error
	}
	var primaryDocument PrimaryDocument
	result = db.First(&primaryDocument, primaryDocumentID)
	if result.Error != nil {
		return "", result.Error
	}
	return filepath.Join(message.StoreDir, primaryDocument.FileName), nil
}

func PreloadRecordObjects(db *gorm.DB) *gorm.DB {
	return db.
		Scopes(PreloadFileRecordObject("FileRecordObjects.", 0, 4)).
		Scopes(PreloadProcessRecordObject("ProcessRecordObjects.", 0, 4)).
		Scopes(PreloadDocumentRecordObject("DocumentRecordObjects."))
}

// PreloadFileRecordObject populates all related data of the file record object.
// Record object children (de: Teilakten, Vorgänge, Dokumente) will be populated as well.
// The prefix is used to populate nested children. The Prefix should be empty if the file record object is the root element.
// Current depth should be initially 0.
// With max depth can the depth of child nesting be configured. A good value is 4 because this complies with xdomea.
func PreloadFileRecordObject(prefix string, currentDepth uint, maxDepth uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		preloadedDB := db.
			Scopes(PreloadGeneralMetadata(prefix)).
			Scopes(PreloadArchiveMetadata(prefix)).
			Preload(prefix + "Lifetime")
		if currentDepth < maxDepth {
			preloadedDB = preloadedDB.
				Scopes(PreloadFileRecordObject(prefix+"SubFileRecordObjects.", currentDepth+1, maxDepth)).
				Scopes(PreloadProcessRecordObject(prefix+"ProcessRecordObjects.", currentDepth+1, maxDepth)).
				Scopes(PreloadDocumentRecordObject(prefix + "DocumentRecordObjects."))
		}
		return preloadedDB
	}
}

// PreloadProcessRecordObject populates all related data of the process record object.
// Record object children (de: Vorgänge, Teilvorgänge, Dokumente) will be populated as well.
// The prefix is used to populate nested children. The Prefix should be empty if the file record object is the root element.
// Current depth should be initially 0.
// With max depth can the depth of child nesting be configured. A good value is 4 because this complies with xdomea.
func PreloadProcessRecordObject(prefix string, currentDepth uint, maxDepth uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		preloadedDB := db.
			Scopes(PreloadGeneralMetadata(prefix)).
			Scopes(PreloadArchiveMetadata(prefix)).
			Preload(prefix + "Lifetime")
		if currentDepth < maxDepth {
			preloadedDB = preloadedDB.
				Scopes(PreloadProcessRecordObject(prefix+"SubProcessRecordObjects.", currentDepth+1, maxDepth)).
				Scopes(PreloadDocumentRecordObject(prefix + "DocumentRecordObjects."))
		}
		return preloadedDB
	}
}

func PreloadDocumentRecordObject(prefix string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.
			Scopes(PreloadGeneralMetadata(prefix)).
			Preload(prefix + "Versions.Formats.PrimaryDocument").
			Scopes(PreloadGeneralMetadata(prefix + "Attachments.")).
			Preload(prefix + "Attachments.Versions.Formats.PrimaryDocument")
	}
}

func PreloadGeneralMetadata(prefix string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.
			Preload(prefix + "GeneralMetadata.FilePlan").
			Preload(prefix + "GeneralMetadata.ConfidentialityLevel").
			Preload(prefix + "GeneralMetadata.Medium")
	}
}

func PreloadArchiveMetadata(prefix string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.
			Preload(prefix + "ArchiveMetadata")
	}
}
