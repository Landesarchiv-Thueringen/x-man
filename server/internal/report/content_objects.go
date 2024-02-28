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
	if contentObject.ContentStats.HasDeviatingAppraisals {
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
	if contentObject.ContentStats.HasDeviatingAppraisals {
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
