package xdomea

import (
	"lath/xman/internal/db"
	"log"

	"github.com/google/uuid"
)

func SetAppraisalForFileRecorcdObjects(
	fileRecordObjectIDs []string,
	appraisalCode string,
	appraisalNote *string,
) ([]db.FileRecordObject, error) {
	updatedFileRecordObjects := []db.FileRecordObject{}
	for _, objectID := range fileRecordObjectIDs {
		id, err := uuid.Parse(objectID)
		if err != nil {
			log.Println(err)
			return updatedFileRecordObjects, err
		}
		fileRecordObject, err := db.SetFileRecordObjectAppraisal(id, appraisalCode, false)
		if err != nil {
			log.Println(err)
			return updatedFileRecordObjects, err
		}
		if appraisalNote != nil {
			fileRecordObject, err = db.SetFileRecordObjectAppraisalNote(id, *appraisalNote)
			if err != nil {
				log.Println(err)
				return updatedFileRecordObjects, err
			}
		}
		updatedFileRecordObjects = append(updatedFileRecordObjects, fileRecordObject)
	}
	return updatedFileRecordObjects, nil
}

func SetAppraisalForProcessRecorcdObjects(
	processRecordObjectIDs []string,
	appraisalCode string,
	appraisalNote *string,
) ([]db.ProcessRecordObject, error) {
	updatedProcessRecordObjects := []db.ProcessRecordObject{}
	for _, objectID := range processRecordObjectIDs {
		id, err := uuid.Parse(objectID)
		if err != nil {
			log.Println(err)
			return updatedProcessRecordObjects, err
		}
		processRecordObject, err := db.SetProcessRecordObjectAppraisal(id, appraisalCode)
		if err != nil {
			log.Println(err)
			return updatedProcessRecordObjects, err
		}
		if appraisalNote != nil {
			processRecordObject, err = db.SetProcessRecordObjectAppraisalNote(id, *appraisalNote)
			if err != nil {
				log.Println(err)
				return updatedProcessRecordObjects, err
			}
		}
		updatedProcessRecordObjects = append(updatedProcessRecordObjects, processRecordObject)
	}
	return updatedProcessRecordObjects, nil
}
