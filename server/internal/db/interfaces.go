package db

import (
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

// GetMaxChildDepth returns the maximal depth of all tree branches from this node.
// If the file record object has only documents as child the maximal child depth is 2.
// If the file record object has subfiles with processes and documents the maximal child depth is 4.
// If the file record object has subfiles with subprocessess ...
// Document attachments are not counted.
func (f *FileRecordObject) GetMaxChildDepth() uint {
	var depth uint = 1
	if len(f.DocumentRecordObjects) > 0 {
		depth = 2
	}
	if len(f.ProcessRecordObjects) > 0 {
		for _, process := range f.ProcessRecordObjects {
			depth = max(depth, process.GetMaxChildDepth()+1)
		}
	}
	if len(f.SubFileRecordObjects) > 0 {
		for _, subfiles := range f.SubFileRecordObjects {
			depth = max(depth, subfiles.GetMaxChildDepth()+1)
		}
	}
	return depth
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

// GetMaxChildDepth returns the maximal depth of all tree branches from this node.
// If the process record object has only documents as child the maximal child depth is 1.
// If the process record object has subprocesses with documents the maximal child depth is 2.
// If the process record object has subprocesses with subprocessess ...
// Document attachments are not counted.
func (p *ProcessRecordObject) GetMaxChildDepth() uint {
	var depth uint = 1
	if len(p.DocumentRecordObjects) > 0 {
		depth = 2
	}
	if len(p.SubProcessRecordObjects) > 0 {
		for _, subprocess := range p.SubProcessRecordObjects {
			depth = max(depth, subprocess.GetMaxChildDepth()+1)
		}
	}
	return depth
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

// GetMaxChildDepth returns the maximal depth of all tree branches from this node.
func (m *Message) GetMaxChildDepth() uint {
	var depth uint = 0
	if len(m.FileRecordObjects) > 0 {
		for _, fileRecordObject := range m.FileRecordObjects {
			depth = max(depth, fileRecordObject.GetMaxChildDepth())
		}
	}
	if len(m.ProcessRecordObjects) > 0 {
		for _, processRecordObject := range m.ProcessRecordObjects {
			depth = max(depth, processRecordObject.GetMaxChildDepth())
		}
	}
	if len(m.DocumentRecordObjects) > 0 {
		depth = max(depth, 1)
	}
	return depth
}

type AppraisableRecordObject interface {
	GetID() uuid.UUID
	GetAppraisableParent() AppraisableRecordObject
	GetAppraisableChildren() []AppraisableRecordObject
}

// GetAppraisableChildren returns a child record objects of file which are appraisable.
func (f *FileRecordObject) GetAppraisableChildren() []AppraisableRecordObject {
	if len(f.SubFileRecordObjects)+len(f.ProcessRecordObjects) == 0 {
		o := FileRecordObject{}
		result := db.
			Preload("SubFileRecordObjects").
			Preload("ProcessRecordObjects").
			First(&o, f.ID)
		if result.Error != nil {
			panic(result.Error)
		}
		f = &o
	}
	appraisableObjects := []AppraisableRecordObject{}
	// add all subfiles (de: Teilakten)
	for _, s := range f.SubFileRecordObjects {
		appraisableObjects = append(appraisableObjects, &s)
		appraisableObjects = append(appraisableObjects, s.GetAppraisableChildren()...)
	}
	// add all processes (de: Vorgänge)
	for _, s := range f.ProcessRecordObjects {
		appraisableObjects = append(appraisableObjects, &s)
		appraisableObjects = append(appraisableObjects, s.GetAppraisableChildren()...)
	}
	return appraisableObjects
}

func (f *FileRecordObject) GetID() uuid.UUID {
	return f.XdomeaID
}

func (f *FileRecordObject) GetAppraisableParent() AppraisableRecordObject {
	if f.ParentFileRecordID != nil {
		parent, found := GetFileRecordObjectByID(*f.ParentFileRecordID)
		if !found {
			panic(fmt.Sprintf("failed to get parent of file record object \"%v\"", f.ID))
		}
		return &parent
	} else {
		return nil
	}
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

func (p *ProcessRecordObject) GetID() uuid.UUID {
	return p.XdomeaID
}

func (p *ProcessRecordObject) GetAppraisableParent() AppraisableRecordObject {
	if p.ParentFileRecordID != nil {
		parent, found := GetFileRecordObjectByID(*p.ParentFileRecordID)
		if !found {
			panic(fmt.Sprintf("failed to get parent of process record object \"%v\"", p.ID))
		}
		return &parent
	} else if p.ParentProcessRecordID != nil {
		parent, found := GetProcessRecordObjectByID(*p.ParentProcessRecordID)
		if !found {
			panic(fmt.Sprintf("failed to get parent of process record object \"%v\"", p.ID))
		}
		return &parent
	} else {
		return nil
	}
}

// GetAppraisableChildren returns a child record objects of process which are appraisable.
func (p *ProcessRecordObject) GetAppraisableChildren() []AppraisableRecordObject {
	if len(p.SubProcessRecordObjects) == 0 {
		o := ProcessRecordObject{}
		result := db.
			Preload("SubProcessRecordObjects").
			First(&o, p.ID)
		if result.Error != nil {
			panic(result.Error)
		}
		p = &o
	}
	appraisableObjects := []AppraisableRecordObject{}
	// add all subprocesses (de: Teilvorgänge)
	for _, s := range p.SubProcessRecordObjects {
		appraisableObjects = append(appraisableObjects, &s)
		appraisableObjects = append(appraisableObjects, s.GetAppraisableChildren()...)
	}
	return appraisableObjects
}

// GetAppraisableObjects returns all files, subfiles, processes and subprocesses of message.
func (m *Message) GetAppraisableObjects() []AppraisableRecordObject {
	message := *m
	if len(m.FileRecordObjects) == 0 {
		message, _ = GetCompleteMessageByID(m.ID)
	}
	var appraisableObjects []AppraisableRecordObject
	for _, f := range message.FileRecordObjects {
		appraisableObjects = append(appraisableObjects, &f)
		appraisableObjects = append(appraisableObjects, f.GetAppraisableChildren()...)
	}
	for _, p := range message.ProcessRecordObjects {
		appraisableObjects = append(appraisableObjects, &p)
		appraisableObjects = append(appraisableObjects, p.GetAppraisableChildren()...)
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
