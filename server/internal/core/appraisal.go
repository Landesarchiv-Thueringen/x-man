package core

import (
	"context"
	"fmt"
	"lath/xman/internal/db"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AppraisableRecordRelations struct {
	Parent       *string // nil for root-level records
	Children     []string
	HasDocuments bool
	Type         db.RecordType
}

type AppraisableRecordsMap map[string]AppraisableRecordRelations

func AppraisableRecords(r *db.RootRecords) AppraisableRecordsMap {
	m := make(AppraisableRecordsMap)
	var appendFileRecords func(parent *string, files []db.FileRecord) (childIDs []string)
	var appendProcessRecords func(parent *string, processes []db.ProcessRecord) (childIDs []string)
	appendFileRecords = func(parent *string, files []db.FileRecord) (childIDs []string) {
		for _, f := range files {
			childIDs = append(childIDs, f.RecordID)
			innerChildIDs := appendFileRecords(&f.RecordID, f.Subfiles)
			innerChildIDs = append(innerChildIDs, appendProcessRecords(&f.RecordID, f.Processes)...)
			m[f.RecordID] = AppraisableRecordRelations{
				Parent:       parent,
				Children:     innerChildIDs,
				HasDocuments: len(f.Documents) > 0,
				Type:         db.RecordTypeFile,
			}
		}
		return
	}
	appendProcessRecords = func(parent *string, processes []db.ProcessRecord) (childIDs []string) {
		for _, p := range processes {
			childIDs = append(childIDs, p.RecordID)
			innerChildIDs := appendProcessRecords(&p.RecordID, p.Subprocesses)
			m[p.RecordID] = AppraisableRecordRelations{
				Parent:       parent,
				Children:     innerChildIDs,
				HasDocuments: len(p.Documents) > 0,
				Type:         db.RecordTypeProcess,
			}
		}
		return
	}
	appendFileRecords(nil, r.Files)
	appendProcessRecords(nil, r.Processes)
	return m
}

// AreAllRecordObjectsAppraised verifies whether every file, subfile, process, and subprocess has been appraised
// with either an 'A' (de: archivieren) or 'V' (de: vernichten).
func AreAllRecordObjectsAppraised(ctx context.Context, processID string) bool {
	rootRecords := db.FindAllRootRecords(ctx, processID, db.MessageType0501)
	m := AppraisableRecords(&rootRecords)
	for id := range m {
		a, _ := db.FindAppraisal(processID, id)
		if a.Decision != "A" && a.Decision != "V" {
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
// If the decision to set is "A" and given record is a sub record, it makes sure
// that all ancestors are also set to "A".
//
// For any other decision, if the given record is a sub record, in case that
// with the given appraisal all siblings have assumed the same decision, it
// updates the parent to match the decision, repeating the process for further
// ancestors.
func SetAppraisalDecisionRecursive(
	processID string,
	recordID string,
	decision db.AppraisalDecisionOption,
) error {
	process, found := db.FindProcess(context.Background(), processID)
	if !found {
		return fmt.Errorf("process not found: %s", processID)
	} else if process.ProcessState.Appraisal.Complete {
		return fmt.Errorf("appraisal already finished for process \"%s\"", processID)
	}
	rootRecords, ok := db.FindRootRecord(context.Background(), processID, db.MessageType0501, recordID)
	if !ok {
		return fmt.Errorf("record object not found: %v", recordID)
	}
	m := AppraisableRecords(&rootRecords)
	previousAppraisal, _ := db.FindAppraisal(processID, recordID)
	db.UpsertAppraisalDecision(processID, recordID, decision)
	if decision == db.AppraisalDecisionA {
		markAncestorsToBeArchived(processID, m, recordID)
	} else {
		matchParentForEqualSiblings(processID, m, recordID, decision)
	}
	propagateAppraisalDecisionDown(processID, recordID, m, decision, previousAppraisal)
	updateAppraisalProcessStep(processID)
	return nil
}

// propagateAppraisalDecisionDown recursively propagates an appraisal decision
// as described in SetAppraisalDecisionRecursive.
func propagateAppraisalDecisionDown(
	processID string,
	recordID string,
	m AppraisableRecordsMap,
	decision db.AppraisalDecisionOption,
	previousAppraisal db.Appraisal,
) {
	for _, subRecordID := range m[recordID].Children {
		a, _ := db.FindAppraisal(processID, subRecordID)
		if a.Decision == "" || a.Decision == previousAppraisal.Decision {
			db.UpsertAppraisal(processID, subRecordID, decision, "")
			propagateAppraisalDecisionDown(processID, subRecordID, m, decision, previousAppraisal)
		}
	}
}

// SetAppraisalDecisionRecursive saves an internal appraisal note for a record
// object to the database.
func SetAppraisalInternalNote(
	processID string,
	recordID string,
	internalNote string,
) error {
	process, found := db.FindProcess(context.Background(), processID)
	if !found {
		return fmt.Errorf("process not found: %s", processID)
	} else if !process.ProcessState.Receive0501.Complete {
		return fmt.Errorf("process \"%s\" has no 0501 message", processID)
	} else if process.ProcessState.Appraisal.Complete {
		return fmt.Errorf("appraisal already finished for process \"%s\"", processID)
	}
	db.UpsertAppraisalNote(processID, recordID, internalNote)
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
	recordIDs []string,
	decision db.AppraisalDecisionOption,
	internalNote string,
) error {
	process, found := db.FindProcess(context.Background(), processID)
	if !found {
		return fmt.Errorf("process not found: %s", processID)
	} else if !process.ProcessState.Receive0501.Complete {
		return fmt.Errorf("process \"%s\" has no 0501 message", processID)
	} else if process.ProcessState.Appraisal.Complete {
		return fmt.Errorf("appraisal already finished for process \"%s\"", processID)
	}
	rootRecords := db.FindAllRootRecords(context.Background(), processID, db.MessageType0501)
	m := AppraisableRecords(&rootRecords)
	isSubAppraisal := map[int]bool{}
	// Mark all record objects as sub appraisals that have an ancestor of which
	// we are setting the appraisal.
	for i, id := range recordIDs {
		if isSubAppraisal[i] {
			continue
		}
		r := m[id]
	SubRecordsLoop:
		for _, subRecordID := range r.Children {
			for j, id := range recordIDs {
				if subRecordID == id {
					isSubAppraisal[j] = true
					continue SubRecordsLoop
				}
			}
		}
	}
	for i, id := range recordIDs {
		if isSubAppraisal[i] {
			db.UpsertAppraisal(processID, id, decision, "")
		} else {
			db.UpsertAppraisal(processID, id, decision, internalNote)
			if decision == db.AppraisalDecisionA {
				markAncestorsToBeArchived(processID, m, id)
			} else {
				matchParentForEqualSiblings(processID, m, id, decision)
			}
		}
	}
	updateAppraisalProcessStep(processID)
	return nil
}

// matchParentForEqualSiblings checks if all siblings of the given recordObject
// have the same appraisal decision and if so, set the same decision for its
// parent. If the parent has been modified, the process is repeated for the
// parent.
func matchParentForEqualSiblings(
	processID string,
	m AppraisableRecordsMap,
	id string,
	decision db.AppraisalDecisionOption,
) {
	parent := m[id].Parent
	if parent != nil {
		// If the record has documents as siblings, these documents
		// automatically assume the parent's appraisal decision since documents
		// themselves are not appraisable. Therefore, there we can never update
		// the parent's appraisal in this case.
		if m[*parent].HasDocuments {
			return
		}
		parentAppraisal, _ := db.FindAppraisal(processID, *parent)
		if parentAppraisal.Decision != decision {
			for _, sibling := range m[*parent].Children {
				a, _ := db.FindAppraisal(processID, sibling)
				if a.Decision != decision {
					return
				}
			}
			db.UpsertAppraisal(processID, *parent, decision, "")
			matchParentForEqualSiblings(processID, m, *parent, decision)
		}
	}
}

func markAncestorsToBeArchived(processID string, m AppraisableRecordsMap, id string) {
	for parent := m[id].Parent; parent != nil; parent = m[*parent].Parent {
		a, _ := db.FindAppraisal(processID, *parent)
		if a.Decision != db.AppraisalDecisionA {
			db.UpsertAppraisal(processID, *parent, db.AppraisalDecisionA, "")
		}
	}
}

func FinalizeMessageAppraisal(message db.Message, completedBy string) db.Message {
	markUnappraisedRecordObjectsAsDiscardable(message)
	db.MustUpdateProcessStepCompletion(
		message.MessageHead.ProcessID,
		db.ProcessStepAppraisal,
		true,
		completedBy,
	)
	return message
}

func markUnappraisedRecordObjectsAsDiscardable(message db.Message) {
	rootRecords := db.FindAllRootRecords(context.Background(), message.MessageHead.ProcessID, message.MessageType)
	for id := range AppraisableRecords(&rootRecords) {
		a, _ := db.FindAppraisal(message.MessageHead.ProcessID, id)
		if a.Decision != "A" && a.Decision != "V" {
			db.UpsertAppraisalDecision(message.MessageHead.ProcessID, id, "V")
		}
	}
}

func updateAppraisalProcessStep(processID string) {
	process, found := db.FindProcess(context.Background(), processID)
	if !found {
		panic(fmt.Errorf("process not found: %s", processID))
	}
	rootRecords := db.FindAllRootRecords(context.Background(), processID, db.MessageType0501)
	var appraisableRootRecordIDs []string
	for _, r := range rootRecords.Files {
		appraisableRootRecordIDs = append(appraisableRootRecordIDs, r.RecordID)
	}
	for _, r := range rootRecords.Processes {
		appraisableRootRecordIDs = append(appraisableRootRecordIDs, r.RecordID)
	}
	numberAppraisalComplete := 0
	for _, r := range appraisableRootRecordIDs {
		a, _ := db.FindAppraisal(processID, r)
		if a.Decision == db.AppraisalDecisionA || a.Decision == db.AppraisalDecisionV {
			numberAppraisalComplete++
		}
	}
	db.MustUpdateProcessStepProgress(
		process.ProcessID,
		db.ProcessStepAppraisal,
		&db.ItemProgress{Done: numberAppraisalComplete, Total: len(appraisableRootRecordIDs)},
		primitive.NilObjectID,
		"",
	)
}
