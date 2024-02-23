package xdomea

import (
	"errors"
	"fmt"
	"lath/xman/internal/db"
	"time"

	"github.com/google/uuid"
)

// AreAllRecordObjectsAppraised verifies whether every file, subfile, process, and subprocess has been appraised
// with either an 'A' (de: archivieren) or 'V' (de: vernichten).
func AreAllRecordObjectsAppraised(message db.Message) bool {
	for _, appraisableObject := range message.GetAppraisableObjects() {
		appraisalCode, found := appraisableObject.GetAppraisal()
		if !found || appraisalCode == "B" {
			return false
		}
	}
	return true
}

func SetAppraisalForFileRecordObjects(
	fileRecordObjectIDs []string,
	appraisalCode string,
	appraisalNote *string,
) ([]db.FileRecordObject, error) {
	updatedFileRecordObjects := []db.FileRecordObject{}
	for _, objectID := range fileRecordObjectIDs {
		id, err := uuid.Parse(objectID)
		if err != nil {
			return updatedFileRecordObjects, fmt.Errorf("failed to parse object ID: %v", err)
		}
		fileRecordObject, err := db.SetFileRecordObjectAppraisal(id, appraisalCode, false)
		if err != nil {
			return updatedFileRecordObjects, fmt.Errorf("failed to set appraisal: %v", err)
		}
		if appraisalNote != nil {
			fileRecordObject, err = db.SetFileRecordObjectAppraisalNote(id, *appraisalNote)
			if err != nil {
				return updatedFileRecordObjects, fmt.Errorf("failed to set appraisal note: %v", err)
			}
		}
		updatedFileRecordObjects = append(updatedFileRecordObjects, fileRecordObject)
	}
	return updatedFileRecordObjects, nil
}

func SetAppraisalForProcessRecordObjects(
	processRecordObjectIDs []string,
	appraisalCode string,
	appraisalNote *string,
) ([]db.ProcessRecordObject, error) {
	updatedProcessRecordObjects := []db.ProcessRecordObject{}
	for _, objectID := range processRecordObjectIDs {
		id, err := uuid.Parse(objectID)
		if err != nil {
			return updatedProcessRecordObjects, fmt.Errorf("failed to parse object ID: %v", err)
		}
		processRecordObject, found := db.GetProcessRecordObjectByID(id)
		if !found {
			return updatedProcessRecordObjects, fmt.Errorf("process record object not found: %v", id)
		}
		err = db.SetProcessRecordObjectAppraisal(&processRecordObject, appraisalCode)
		if err != nil {
			return updatedProcessRecordObjects, fmt.Errorf("failed to set appraisal: %v", err)
		}
		if appraisalNote != nil {
			db.SetProcessRecordObjectAppraisalNote(&processRecordObject, *appraisalNote)
		}
		updatedProcessRecordObjects = append(updatedProcessRecordObjects, processRecordObject)
	}
	return updatedProcessRecordObjects, nil
}

func FinalizeMessageAppraisal(message db.Message) db.Message {
	markUnappraisedRecordObjectsAsDiscardable(message)
	process, found := db.GetProcessByXdomeaID(message.MessageHead.ProcessID)
	if !found {
		panic(fmt.Sprintf("process not found: %v", message.MessageHead.ProcessID))
	}
	appraisalStep := process.ProcessState.Appraisal
	appraisalStep.Complete = true
	completionTime := time.Now()
	appraisalStep.CompletionTime = &completionTime
	db.UpdateProcessStep(appraisalStep)
	message.AppraisalComplete = true
	db.UpdateMessage(message)
	return message
}

func markUnappraisedRecordObjectsAsDiscardable(message db.Message) {
	for _, appraisableObject := range message.GetAppraisableObjects() {
		appraisalCode, found := appraisableObject.GetAppraisal()
		if !found || appraisalCode == "B" {
			if err := appraisableObject.SetAppraisal("V"); err != nil {
				panic(err)
			}
		}
	}
}

func TransferAppraisalNoteFrom0501To0503(process db.Process) error {
	if process.Message0501 == nil {
		return errors.New("0501 message doesn't exist")
	}
	fileRecordObjects0501, err := db.GetAllFileRecordObjects(process.Message0501.ID)
	if err != nil {
		return err
	}
	processRecordObjects0501, err := db.GetAllProcessRecordObjects(process.Message0501.ID)
	if err != nil {
		return err
	}
	fileRecordObjects0503, err := db.GetAllFileRecordObjects(process.Message0503.ID)
	if err != nil {
		return err
	}
	processRecordObjects0503, err := db.GetAllProcessRecordObjects(process.Message0503.ID)
	if err != nil {
		return err
	}
	for recordObjectID, file0503 := range fileRecordObjects0503 {
		file0501, ok := fileRecordObjects0501[recordObjectID]
		if !ok {
			return errors.New("file record object with ID " +
				recordObjectID.String() + " not found in 0501 message")
		}
		note, err := file0501.GetAppraisalNote()
		// if no appraisal note for the file record object exists, continue with next
		if err != nil {
			continue
		}
		err = file0503.SetAppraisalNote(note)
		if err != nil {
			return err
		}
	}
	for recordObjectID, process0503 := range processRecordObjects0503 {
		process0501, ok := processRecordObjects0501[recordObjectID]
		if !ok {
			return errors.New("process record object with ID " +
				recordObjectID.String() + " not found in 0501 message")
		}
		note, err := process0501.GetAppraisalNote()
		// if no appraisal note for the process record object exists, continue with next
		if err != nil {
			continue
		}
		process0503.SetAppraisalNote(note)
	}
	return nil
}
