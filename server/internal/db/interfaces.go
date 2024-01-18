package db

import (
	"errors"
	"path/filepath"

	"github.com/google/uuid"
)

// interfaces and methods

func (process *Process) IsArchivable() bool {
	state := process.ProcessState
	return state.FormatVerification.Complete && !state.Archiving.Complete
}

func (message *Message) GetRemoteXmlPath(importDir string) string {
	return filepath.Join(importDir, filepath.Base(message.MessagePath))
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
	for processIndex := range f.Processes {
		recordObjects = append(recordObjects, &f.Processes[processIndex])
		recordObjects = append(recordObjects, f.Processes[processIndex].GetChildren()...)
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
	for documentIndex := range p.Documents {
		recordObjects = append(recordObjects, &p.Documents[documentIndex])
		recordObjects = append(recordObjects, p.Documents[documentIndex].GetChildren()...)
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

func (p *ProcessRecordObject) SetMessageID(messageID uuid.UUID) {
	p.MessageID = messageID
}

// GetChildren Returns an empty list.
// Document record objects do not have any other record objects as their children.
// This might change in future xdomea versions.
func (d *DocumentRecordObject) GetChildren() []RecordObject {
	recordObjects := []RecordObject{}
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
	appraisal, err := GetAppraisalByCode(appraisalCode)
	// unknown appraisal code
	if err != nil {
		return errors.New("unknown appraisal code")
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
	return result.Error
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
	for processIndex := range f.Processes {
		appraisableObjects = append(appraisableObjects, &f.Processes[processIndex])
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

func (p *ProcessRecordObject) GetAppraisal() (string, error) {
	if p.ArchiveMetadata != nil &&
		p.ArchiveMetadata.AppraisalCode != nil {
		return *p.ArchiveMetadata.AppraisalCode, nil
	}
	return "", errors.New("no appraisal existing")
}

func (p *ProcessRecordObject) SetAppraisal(appraisalCode string) error {
	appraisal, err := GetAppraisalByCode(appraisalCode)
	// unknown appraisal code
	if err != nil {
		return errors.New("unknown appraisal code")
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
	return result.Error
}

func (p *ProcessRecordObject) GetAppraisalNote() (string, error) {
	if p.ArchiveMetadata != nil &&
		p.ArchiveMetadata.InternalAppraisalNote != nil {
		return *p.ArchiveMetadata.InternalAppraisalNote, nil
	}
	return "", errors.New("no appraisal note existing")
}

func (p *ProcessRecordObject) SetAppraisalNote(note string) error {
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
	return result.Error
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
