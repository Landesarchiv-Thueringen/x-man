package xdomea

import (
	"lath/xdomea/internal/db"
)

func Generate0502Message(message db.Message) {
	for _, r := range message.RecordObjects {
		if r.FileRecordObject != nil {
			GenerateAppraisedObject(*r.FileRecordObject)
		}
	}
}

func GenerateAppraisedObject(fileRecordObject db.FileRecordObject) db.AppraisedObject {
	var appraisedObject db.AppraisedObject
	if fileRecordObject.ArchiveMetadata != nil &&
		fileRecordObject.ArchiveMetadata.AppraisalCode != nil {
		appraisalCode := db.AppraisalCode{
			Code: *fileRecordObject.ArchiveMetadata.AppraisalCode,
		}
		objectAppraisal := db.ObjectAppraisal{
			AppraisalCode: appraisalCode,
		}
		appraisedObject = db.AppraisedObject{
			XdomeaID:        fileRecordObject.ID,
			ObjectAppraisal: objectAppraisal,
		}
	}
	return appraisedObject
}
