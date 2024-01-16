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

// SetVersionIndependentSubFiles appends the version specific sub files to the version independent sub files.
func (f *FileRecordObject) SetVersionIndependentSubFiles() {
	if f.SubFilesPreXdomeaVersion300 != nil {
		f.SubFileRecordObjects = append(f.SubFileRecordObjects, f.SubFilesPreXdomeaVersion300...)
	}
	if f.SubFilesFromXdomeaVersion300 != nil {
		f.SubFileRecordObjects = append(f.SubFileRecordObjects, f.SubFilesFromXdomeaVersion300...)
	}
	if f.SubFileRecordObjects != nil {
		for subFileIndex := range f.SubFileRecordObjects {
			f.SubFileRecordObjects[subFileIndex].SetVersionIndependentSubFiles()
		}
	}
}

type RecordObjectIter interface {
	GetChildren() []RecordObjectIter
	GetPrimaryDocuments() []PrimaryDocument
	SetMessageID(messageID uuid.UUID)
}

// GetChildren returns all list of all record objects contained in the file record object.
// The original child objects are returned instead of duplicates, allowing for persistent attribute changes.
func (f *FileRecordObject) GetChildren() []RecordObjectIter {
	recordObjects := []RecordObjectIter{}
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
func (p *ProcessRecordObject) GetChildren() []RecordObjectIter {
	recordObjects := []RecordObjectIter{}
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

func (d *DocumentRecordObject) SetMessageID(messageID uuid.UUID) {
	d.MessageID = messageID
}

type AppraisableRecordObject interface {
	GetAppraisal() (string, error)
	SetAppraisal(string) error
	GetID() uuid.UUID
	GetAppraisableObjects() []AppraisableRecordObject
}

func (r *RecordObject) GetAppraisableObjects() []AppraisableRecordObject {
	appraisableObjects := []AppraisableRecordObject{}
	if r.FileRecordObject != nil {
		appraisableObjects = append(appraisableObjects, r.FileRecordObject.GetAppraisableObjects()...)
	}
	// TODO: add process and document
	return appraisableObjects
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
