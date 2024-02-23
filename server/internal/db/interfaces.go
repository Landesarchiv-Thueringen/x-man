package db

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
)

// interfaces and methods

func (process *Process) IsArchivable() bool {
	state := process.ProcessState
	return state.FormatVerification.Complete && !state.Archiving.Complete
}
func (message *Message) GetMessageFileName() string {
	return filepath.Base(message.MessagePath)
}

func (message *Message) GetRemoteXmlPath(importDir string) string {
	return filepath.Join(importDir, message.GetMessageFileName())
}

type RecordObject interface {
	GetChildren() []RecordObject
	GetPrimaryDocuments() []PrimaryDocument
	SetMessageID(messageID uuid.UUID)
}

// GetChildren returns all list of all record objects contained in the file record object.
// The original child objects are returned instead of duplicates, allowing for persistent attribute changes.
func (f *FileRecordObject) GetChildren() []RecordObject {
	recordObjects := []RecordObject{}
	for subfileIndex := range f.SubFileRecordObjects {
		recordObjects = append(recordObjects, &f.SubFileRecordObjects[subfileIndex])
		recordObjects = append(recordObjects, f.SubFileRecordObjects[subfileIndex].GetChildren()...)
	}
	for processIndex := range f.ProcessRecordObjects {
		recordObjects = append(recordObjects, &f.ProcessRecordObjects[processIndex])
		recordObjects = append(recordObjects, f.ProcessRecordObjects[processIndex].GetChildren()...)
	}
	return recordObjects
}

func (f *FileRecordObject) GetPrimaryDocuments() []PrimaryDocument {
	primaryDocuments := []PrimaryDocument{}
	for _, process := range f.ProcessRecordObjects {
		primaryDocuments = append(primaryDocuments, process.GetPrimaryDocuments()...)
	}
	return primaryDocuments
}

func (f *FileRecordObject) SetMessageID(messageID uuid.UUID) {
	f.MessageID = messageID
}

// GetChildren returns all list of all record objects contained in the process record object.
// The original child objects are returned instead of duplicates, allowing for persistent attribute changes.
func (p *ProcessRecordObject) GetChildren() []RecordObject {
	recordObjects := []RecordObject{}
	// de: Teilvorgänge
	for subprocessIndex := range p.SubProcessRecordObjects {
		recordObjects = append(recordObjects, &p.SubProcessRecordObjects[subprocessIndex])
		recordObjects = append(recordObjects, p.SubProcessRecordObjects[subprocessIndex].GetChildren()...)
	}
	// de: Dokumente
	for documentIndex := range p.DocumentRecordObjects {
		recordObjects = append(recordObjects, &p.DocumentRecordObjects[documentIndex])
		recordObjects = append(recordObjects, p.DocumentRecordObjects[documentIndex].GetChildren()...)
	}
	return recordObjects
}

func (p *ProcessRecordObject) GetPrimaryDocuments() []PrimaryDocument {
	primaryDocuments := []PrimaryDocument{}
	for _, document := range p.DocumentRecordObjects {
		primaryDocuments = append(primaryDocuments, document.GetPrimaryDocuments()...)
	}
	return primaryDocuments
}

func (p *ProcessRecordObject) SetMessageID(messageID uuid.UUID) {
	p.MessageID = messageID
}

// GetChildren returns all list of all attachments contained in the document record object.
// The original child objects are returned instead of duplicates, allowing for persistent attribute changes.
func (d *DocumentRecordObject) GetChildren() []RecordObject {
	recordObjects := []RecordObject{}
	for index := range d.Attachments {
		recordObjects = append(recordObjects, &d.Attachments[index])
		//recordObjects = append(recordObjects, d.Attachments[index].GetChildren()...)
	}
	return recordObjects
}

// GetPrimaryDocuments returns all primary documents of the document record object.
// All primary documents of attachments are returned as well.
func (d *DocumentRecordObject) GetPrimaryDocuments() []PrimaryDocument {
	primaryDocuments := []PrimaryDocument{}
	for _, version := range d.Versions {
		for _, format := range version.Formats {
			primaryDocuments = append(primaryDocuments, format.PrimaryDocument)
		}
	}
	for _, attachment := range d.Attachments {
		primaryDocuments = append(primaryDocuments, attachment.GetPrimaryDocuments()...)
	}
	return primaryDocuments
}

func (d *DocumentRecordObject) SetMessageID(messageID uuid.UUID) {
	d.MessageID = messageID
}

// GetRecordObjects retrieves all record objects from the root level of the message.
// The child record objects at the root level are not separately returned from their parent record object.
func (m *Message) GetRecordObjects() []RecordObject {
	var recordObjects []RecordObject
	for index := range m.FileRecordObjects {
		recordObjects = append(recordObjects, &m.FileRecordObjects[index])
	}
	for index := range m.ProcessRecordObjects {
		recordObjects = append(recordObjects, &m.ProcessRecordObjects[index])
	}
	for index := range m.DocumentRecordObjects {
		recordObjects = append(recordObjects, &m.DocumentRecordObjects[index])
	}
	return recordObjects
}

type AppraisableRecordObject interface {
	GetAppraisal() (string, bool)
	SetAppraisal(string) error
	GetID() uuid.UUID
	GetAppraisableObjects() []AppraisableRecordObject
}

func (f *FileRecordObject) GetAppraisal() (string, bool) {
	if f.ArchiveMetadata != nil &&
		f.ArchiveMetadata.AppraisalCode != nil {
		return *f.ArchiveMetadata.AppraisalCode, true
	}
	return "", false
}

func (f *FileRecordObject) SetAppraisal(appraisalCode string) error {
	appraisal, found := GetAppraisalByCode(appraisalCode)
	if !found {
		return fmt.Errorf("unknown appraisal code: %v", appraisalCode)
	}
	// archive metadata not created
	if f.ArchiveMetadata == nil {
		archiveMetadata := ArchiveMetadata{
			AppraisalCode: &appraisal.Code,
		}
		f.ArchiveMetadata = &archiveMetadata
	} else {
		f.ArchiveMetadata.AppraisalCode = &appraisal.Code
	}
	// save archive metadata
	result := db.Save(&f.ArchiveMetadata)
	if result.Error != nil {
		panic(result.Error)
	}
	return nil
}

func (f *FileRecordObject) GetAppraisalNote() (string, error) {
	if f.ArchiveMetadata != nil &&
		f.ArchiveMetadata.InternalAppraisalNote != nil {
		return *f.ArchiveMetadata.InternalAppraisalNote, nil
	}
	return "", errors.New("no appraisal note existing")
}

func (f *FileRecordObject) SetAppraisalNote(note string) error {
	// archive metadata not created
	if f.ArchiveMetadata == nil {
		archiveMetadata := ArchiveMetadata{
			InternalAppraisalNote: &note,
		}
		f.ArchiveMetadata = &archiveMetadata
	} else {
		f.ArchiveMetadata.InternalAppraisalNote = &note
	}
	// save archive metadata
	result := db.Save(&f.ArchiveMetadata)
	return result.Error
}

// GetAppraisableObjects returns a child record objects of file which are appraisable.
func (f *FileRecordObject) GetAppraisableObjects() []AppraisableRecordObject {
	appraisableObjects := []AppraisableRecordObject{f}
	// add all subfiles (de: Teilakten)
	for subfileIndex := range f.SubFileRecordObjects {
		appraisableObjects = append(appraisableObjects, &f.SubFileRecordObjects[subfileIndex])
	}
	// add all processes (de: Vorgänge)
	for processIndex := range f.ProcessRecordObjects {
		appraisableObjects = append(appraisableObjects, &f.ProcessRecordObjects[processIndex])
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

// GetCombinedLifetime returns a string representation of lifetime start and end.
func (f *FileRecordObject) GetCombinedLifetime() string {
	if f.Lifetime != nil {
		if f.Lifetime.Start != nil && f.Lifetime.End != nil {
			return *f.Lifetime.Start + " - " + *f.Lifetime.End
		} else if f.Lifetime.Start != nil {
			return *f.Lifetime.Start + " - "
		} else if f.Lifetime.End != nil {
			return " - " + *f.Lifetime.End
		}
	}
	return ""
}

func (p *ProcessRecordObject) GetTitle() string {
	title := "Vorgang"
	if p.GeneralMetadata != nil {
		if p.GeneralMetadata.XdomeaID != nil {
			title += " " + *p.GeneralMetadata.XdomeaID
		}
		if p.GeneralMetadata.Subject != nil {
			title += ": " + *p.GeneralMetadata.Subject
		}
	}
	return title
}

// GetCombinedLifetime returns a string representation of lifetime start and end.
func (p *ProcessRecordObject) GetCombinedLifetime() string {
	if p.Lifetime != nil {
		if p.Lifetime.Start != nil && p.Lifetime.End != nil {
			return *p.Lifetime.Start + " - " + *p.Lifetime.End
		} else if p.Lifetime.Start != nil {
			return *p.Lifetime.Start + " - "
		} else if p.Lifetime.End != nil {
			return " - " + *p.Lifetime.End
		}
	}
	return ""
}

func (p *ProcessRecordObject) GetAppraisal() (string, bool) {
	if p.ArchiveMetadata != nil &&
		p.ArchiveMetadata.AppraisalCode != nil {
		return *p.ArchiveMetadata.AppraisalCode, true
	}
	return "", false
}

func (p *ProcessRecordObject) SetAppraisal(appraisalCode string) error {
	appraisal, found := GetAppraisalByCode(appraisalCode)
	if !found {
		return fmt.Errorf("unknown appraisal code: %v", appraisalCode)
	}
	// archive metadata not created
	if p.ArchiveMetadata == nil {
		archiveMetadata := ArchiveMetadata{
			AppraisalCode: &appraisal.Code,
		}
		p.ArchiveMetadata = &archiveMetadata
	} else {
		p.ArchiveMetadata.AppraisalCode = &appraisal.Code
	}
	// save archive metadata
	result := db.Save(&p.ArchiveMetadata)
	if result.Error != nil {
		panic(result.Error)
	}
	return nil
}

func (p *ProcessRecordObject) GetAppraisalNote() (string, error) {
	if p.ArchiveMetadata != nil &&
		p.ArchiveMetadata.InternalAppraisalNote != nil {
		return *p.ArchiveMetadata.InternalAppraisalNote, nil
	}
	return "", errors.New("no appraisal note existing")
}

func (p *ProcessRecordObject) SetAppraisalNote(note string) {
	// archive metadata not created
	if p.ArchiveMetadata == nil {
		archiveMetadata := ArchiveMetadata{
			InternalAppraisalNote: &note,
		}
		p.ArchiveMetadata = &archiveMetadata
	} else {
		p.ArchiveMetadata.InternalAppraisalNote = &note
	}
	// save archive metadata
	result := db.Save(&p.ArchiveMetadata)
	if result.Error != nil {
		panic(result.Error)
	}
}

func (p *ProcessRecordObject) GetID() uuid.UUID {
	return p.XdomeaID
}

// GetAppraisableObjects returns a child record objects of process which are appraisable.
func (p *ProcessRecordObject) GetAppraisableObjects() []AppraisableRecordObject {
	appraisableObjects := []AppraisableRecordObject{p}
	// add all subprocesses (de: Teilvorgänge)
	for index := range p.SubProcessRecordObjects {
		appraisableObjects = append(appraisableObjects, &p.SubProcessRecordObjects[index])
	}
	return appraisableObjects
}

// GetAppraisableObjects returns all files, subfiles, processes and subprocesses of message.
func (m *Message) GetAppraisableObjects() []AppraisableRecordObject {
	var appraisableObjects []AppraisableRecordObject
	for index := range m.FileRecordObjects {
		appraisableObjects = append(appraisableObjects, m.FileRecordObjects[index].GetAppraisableObjects()...)
	}
	for index := range m.ProcessRecordObjects {
		appraisableObjects = append(appraisableObjects, m.FileRecordObjects[index].GetAppraisableObjects()...)
	}
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
