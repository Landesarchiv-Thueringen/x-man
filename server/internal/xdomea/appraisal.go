package xdomea

import (
	"fmt"
	"lath/xman/internal/db"
	"time"

	"github.com/google/uuid"
)

// AreAllRecordObjectsAppraised verifies whether every file, subfile, process, and subprocess has been appraised
// with either an 'A' (de: archivieren) or 'V' (de: vernichten).
func AreAllRecordObjectsAppraised(message db.Message) bool {
	for _, appraisableObject := range message.GetAppraisableObjects() {
		appraisal := db.GetAppraisal(message.MessageHead.ProcessID, appraisableObject.GetID())
		if appraisal.Decision != "A" && appraisal.Decision != "V" {
			return false
		}
	}
	return true
}

// SetAppraisalDecisionRecursive saves an appraisal decision for a record object
// to the database.
//
// It updates child objects if the given record object to the new appraisal
// decision if
//   - the child object has not yet been appraised, or
//   - the child object has been appraised with the same appraisal decision as
//     the given object had before.
//
// It clears the internal appraisal note of children that it updates.
//
// If the decision to set is "A" and given record object is a sub object, it
// makes sure that all ancestors are also set to "A".
func SetAppraisalDecisionRecursive(
	processID string,
	recordObjectID string,
	decision db.AppraisalDecisionOption,
) error {
	process, found := db.GetProcess(processID)
	if !found {
		return fmt.Errorf("process not found: %s", processID)
	} else if process.Message0501ID == nil {
		return fmt.Errorf("process \"%s\" has no 0501 message", processID)
	} else if process.ProcessState.Appraisal.Complete {
		return fmt.Errorf("appraisal already finished for process \"%s\"", processID)
	}
	parsedRecordObjectID, err := uuid.Parse(recordObjectID)
	if err != nil {
		return fmt.Errorf("failed to parse object ID \"%s\": %v", recordObjectID, err)
	}
	recordObject := db.GetAppraisableRecordObject(*process.Message0501ID, parsedRecordObjectID)
	if recordObject == nil {
		return fmt.Errorf("record object not found: %v", parsedRecordObjectID)
	}
	previousAppraisal := db.GetAppraisal(processID, recordObject.GetID())
	db.SetAppraisalDecision(processID, recordObject.GetID(), decision)
	if decision == db.AppraisalDecisionA {
		markAncestorsToBeArchived(processID, recordObject)
	}
	for _, subObject := range recordObject.GetAppraisableChildren() {
		a := db.GetAppraisal(processID, subObject.GetID())
		if a.Decision == "" || a.Decision == previousAppraisal.Decision {
			a.Decision = decision
			a.InternalNote = ""
			db.UpdateAppraisal(a)
		}
	}
	return nil
}

// SetAppraisalDecisionRecursive saves an internal appraisal note for a record
// object to the database.
func SetAppraisalInternalNote(
	processID string,
	recordObjectID string,
	internalNote string,
) error {
	process, found := db.GetProcess(processID)
	if !found {
		return fmt.Errorf("process not found: %s", processID)
	} else if process.Message0501ID == nil {
		return fmt.Errorf("process \"%s\" has no 0501 message", processID)
	} else if process.ProcessState.Appraisal.Complete {
		return fmt.Errorf("appraisal already finished for process \"%s\"", processID)
	}
	parsedRecordObjectID, err := uuid.Parse(recordObjectID)
	if err != nil {
		return fmt.Errorf("failed to parse object ID \"%s\": %v", recordObjectID, err)
	}
	recordObject := db.GetAppraisableRecordObject(*process.Message0501ID, parsedRecordObjectID)
	if recordObject == nil {
		return fmt.Errorf("record object not found: %v", parsedRecordObjectID)
	}
	db.SetAppraisalInternalNote(processID, recordObject.GetID(), internalNote)
	return nil
}

// SetAppraisals saves an appraisal decision and optional internal note for a
// number of record objects to the database.
//
// It checks whether any of the given objects are children of on another and
// omits the appraisal note for all child objects.
//
// If the decision to set is "A", it makes sure that for all sub objects, all
// ancestors are also set to "A".
func SetAppraisals(
	processID string,
	recordObjectIDs []string,
	decision db.AppraisalDecisionOption,
	internalNote string,
) error {
	process, found := db.GetProcess(processID)
	if !found {
		return fmt.Errorf("process not found: %s", processID)
	} else if process.Message0501ID == nil {
		return fmt.Errorf("process \"%s\" has no 0501 message", processID)
	} else if process.ProcessState.Appraisal.Complete {
		return fmt.Errorf("appraisal already finished for process \"%s\"", processID)
	}
	recordObjects := make([]db.AppraisableRecordObject, len(recordObjectIDs))
	for i, idString := range recordObjectIDs {
		recordObjectID, err := uuid.Parse(idString)
		if err != nil {
			return fmt.Errorf("failed to parse object ID \"%s\": %v", idString, err)
		}
		recordObject := db.GetAppraisableRecordObject(*process.Message0501ID, recordObjectID)
		if recordObject == nil {
			return fmt.Errorf("record object not found: %v", recordObjectID)
		}
		recordObjects[i] = recordObject
	}
	isSubAppraisal := map[int]bool{}
	// Mark all record objects as sub appraisals that have an ancestor of which
	// we are setting the appraisal.
	for i, recordObject := range recordObjects {
		if isSubAppraisal[i] {
			continue
		}
	SubObjectsLoop:
		for _, subObject := range recordObject.GetAppraisableChildren() {
			for j, o := range recordObjects {
				if subObject.GetID() == o.GetID() {
					isSubAppraisal[j] = true
					continue SubObjectsLoop
				}
			}
		}
	}
	for i, recordObject := range recordObjects {
		if isSubAppraisal[i] {
			db.SetAppraisal(processID, recordObject.GetID(), decision, "")
		} else {
			db.SetAppraisal(processID, recordObject.GetID(), decision, internalNote)
			if decision == db.AppraisalDecisionA {
				markAncestorsToBeArchived(processID, recordObject)
			}
		}
	}
	return nil
}

func markAncestorsToBeArchived(processID string, recordObject db.AppraisableRecordObject) {
	for parent := recordObject.GetAppraisableParent(); parent != nil; parent = parent.GetAppraisableParent() {
		a := db.GetAppraisal(processID, parent.GetID())
		if a.Decision != db.AppraisalDecisionA {
			a.Decision = db.AppraisalDecisionA
			a.InternalNote = ""
			db.UpdateAppraisal(a)
		}
	}
}

func FinalizeMessageAppraisal(message db.Message, completedBy string) db.Message {
	markUnappraisedRecordObjectsAsDiscardable(message)
	process, found := db.GetProcess(message.MessageHead.ProcessID)
	if !found {
		panic(fmt.Sprintf("process not found: %v", message.MessageHead.ProcessID))
	}
	completionTime := time.Now()
	db.UpdateProcessStep(process.ProcessState.Appraisal.ID, db.ProcessStep{
		Complete:       true,
		CompletionTime: &completionTime,
		CompletedBy:    &completedBy,
	})
	return message
}

func markUnappraisedRecordObjectsAsDiscardable(message db.Message) {
	for _, appraisableObject := range message.GetAppraisableObjects() {
		appraisal := db.GetAppraisal(message.MessageHead.ProcessID, appraisableObject.GetID())
		if appraisal.Decision != "A" && appraisal.Decision != "V" {
			db.SetAppraisalDecision(message.MessageHead.ProcessID, appraisableObject.GetID(), "V")
		}
	}
}
