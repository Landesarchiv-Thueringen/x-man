package report

import (
	"lath/xman/internal/db"

	"github.com/google/uuid"
)

type RecordObjectType = string

const (
	File       RecordObjectType = "file"
	SubFile    RecordObjectType = "subFile"
	Process    RecordObjectType = "process"
	SubProcess RecordObjectType = "subProcess"
)

type ContentObject struct {
	XdomeaID         uuid.UUID           `json:"xdomeaID"`
	RecordObjectType RecordObjectType    `json:"recordObjectType"`
	GeneralMetadata  *db.GeneralMetadata `json:"generalMetadata"`
	ArchiveMetadata  *db.ArchiveMetadata `json:"archiveMetadata"`
	Lifetime         *db.Lifetime        `json:"lifetime"`
	ContentStats     AppraisalStats      `json:"contentStats"`
	Children         []ContentObject     `json:"children"`
}

func getContentObjects(message0501 db.Message, message0503 db.Message) (contentObjects []ContentObject) {
	contentObjects = make([]ContentObject, 0)
	var message db.Message
	if message0501.ID != uuid.Nil {
		message = message0501
	} else {
		message = message0503
	}
	for _, file := range message.FileRecordObjects {
		contentObjects = append(contentObjects, getContentObjectForFile(file, false))
	}
	for _, process := range message.ProcessRecordObjects {
		contentObjects = append(contentObjects, getContentObjectForProcess(process, false))
	}
	return
}

// getContentObjectForFile generates the ContentObjects for the given file and
// populates its children there are any deviating appraisals.
func getContentObjectForFile(
	file db.FileRecordObject,
	isSubObject bool,
) (contentObject ContentObject) {
	contentObject.XdomeaID = file.XdomeaID
	if isSubObject {
		contentObject.RecordObjectType = SubFile
	} else {
		contentObject.RecordObjectType = File
	}
	contentObject.GeneralMetadata = file.GeneralMetadata
	contentObject.ArchiveMetadata = file.ArchiveMetadata
	contentObject.Lifetime = file.Lifetime
	contentObject.ContentStats = getFileAppraisalStats(file)
	if fileHasDeviatingAppraisals(file) {
		contentObject.Children = make([]ContentObject, 0)
		for _, subFile := range file.SubFileRecordObjects {
			contentObject.Children = append(
				contentObject.Children,
				getContentObjectForFile(subFile, true),
			)
		}
		for _, process := range file.ProcessRecordObjects {
			contentObject.Children = append(
				contentObject.Children,
				getContentObjectForProcess(process, false),
			)
		}
	}
	return
}

// getContentObjectForProcess generates the ContentObjects for the given process
// and populates its children there are any deviating appraisals.
func getContentObjectForProcess(
	process db.ProcessRecordObject,
	isSubObject bool,
) (contentObject ContentObject) {
	contentObject.XdomeaID = process.XdomeaID
	if isSubObject {
		contentObject.RecordObjectType = SubProcess
	} else {
		contentObject.RecordObjectType = Process
	}
	contentObject.GeneralMetadata = process.GeneralMetadata
	contentObject.ArchiveMetadata = process.ArchiveMetadata
	contentObject.Lifetime = process.Lifetime
	contentObject.ContentStats = getProcessAppraisalStats(process)
	if processHasDeviatingAppraisals(process) {
		contentObject.Children = make([]ContentObject, 0)
		for _, subProcess := range process.SubProcessRecordObjects {
			contentObject.Children = append(
				contentObject.Children,
				getContentObjectForProcess(subProcess, true),
			)
		}
	}
	return
}

// fileHasDeviatingAppraisals returns true if the given file has any child object
// with an appraisal code that differs from the file's own appraisal code.
func fileHasDeviatingAppraisals(file db.FileRecordObject) bool {
	for _, subFile := range file.SubFileRecordObjects {
		if *subFile.ArchiveMetadata.AppraisalCode != *file.ArchiveMetadata.AppraisalCode ||
			fileHasDeviatingAppraisals(subFile) {
			return true
		}
	}
	for _, process := range file.ProcessRecordObjects {
		if *process.ArchiveMetadata.AppraisalCode != *file.ArchiveMetadata.AppraisalCode ||
			processHasDeviatingAppraisals(process) {
			return true
		}
	}
	return false
}

// processHasDeviatingAppraisals returns true if the given process has any child object
// with an appraisal code that differs from the process's own appraisal code.
func processHasDeviatingAppraisals(process db.ProcessRecordObject) bool {
	for _, subProcess := range process.SubProcessRecordObjects {
		if *subProcess.ArchiveMetadata.AppraisalCode != *process.ArchiveMetadata.AppraisalCode ||
			processHasDeviatingAppraisals(subProcess) {
			return true
		}
	}
	return false
}
