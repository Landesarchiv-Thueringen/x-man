package db

import (
	"errors"
	"path/filepath"

	"github.com/google/uuid"
)

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
