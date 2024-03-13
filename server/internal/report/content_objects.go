package report

import (
	"lath/xman/internal/db"
	"sort"

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
	Appraisal        db.Appraisal        `json:"appraisal"`
	Lifetime         *db.Lifetime        `json:"lifetime"`
	Children         []ContentObject     `json:"children"`
}

func getContentObjects(message db.Message) (contentObjects []ContentObject) {
	contentObjects = make([]ContentObject, 0)
	m := getAppraisalsMap(message.MessageHead.ProcessID)
	for _, file := range message.FileRecordObjects {
		contentObjects = append(contentObjects, getContentObjectForFile(file, false, m))
	}
	for _, process := range message.ProcessRecordObjects {
		contentObjects = append(contentObjects, getContentObjectForProcess(process, false, m))
	}
	sort.Slice(contentObjects, func(i, j int) bool {
		lhs := contentObjects[i]
		rhs := contentObjects[j]
		return *lhs.ArchiveMetadata.AppraisalCode < *rhs.ArchiveMetadata.AppraisalCode
	})
	return
}

// getContentObjectForFile generates the ContentObjects for the given file and
// populates its children there are any deviating appraisals.
func getContentObjectForFile(
	file db.FileRecordObject,
	isSubObject bool,
	m appraisalMap,
) (contentObject ContentObject) {
	contentObject.XdomeaID = file.XdomeaID
	if isSubObject {
		contentObject.RecordObjectType = SubFile
	} else {
		contentObject.RecordObjectType = File
	}
	contentObject.GeneralMetadata = file.GeneralMetadata
	contentObject.ArchiveMetadata = file.ArchiveMetadata
	contentObject.Appraisal = m[file.XdomeaID]
	contentObject.Lifetime = file.Lifetime
	return
}

// getContentObjectForProcess generates the ContentObjects for the given process
// and populates its children there are any deviating appraisals.
func getContentObjectForProcess(
	process db.ProcessRecordObject,
	isSubObject bool,
	m appraisalMap,
) (contentObject ContentObject) {
	contentObject.XdomeaID = process.XdomeaID
	if isSubObject {
		contentObject.RecordObjectType = SubProcess
	} else {
		contentObject.RecordObjectType = Process
	}
	contentObject.GeneralMetadata = process.GeneralMetadata
	contentObject.ArchiveMetadata = process.ArchiveMetadata
	contentObject.Appraisal = m[process.XdomeaID]
	contentObject.Lifetime = process.Lifetime
	return
}
