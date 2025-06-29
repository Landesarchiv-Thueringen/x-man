package core

import (
	"context"
	"lath/xman/internal/db"
)

type Discrepancies struct {
	MissingRecords []string
	SurplusRecords []string
}

type recordNode struct {
	Title  string
	Type   db.RecordType
	Parent *string // nil for root-level records
}

type recordMap map[string]recordNode

// appraisableID returns the record ID of the record of whose appraisal decision
// is relevant for the given record. In case of file and process records, this
// is the record itself. In case of documents, it is the nearest appraisable
// parent.
func (m recordMap) appraisableID(recordID string) string {
	for {
		r := m[recordID]
		if r.Type != db.RecordTypeDocument {
			return recordID
		}
		recordID = *r.Parent
	}
}

func getRecordMap(r *db.RootRecords) recordMap {
	m := make(recordMap)
	var appendFiles func(parent *string, files []db.FileRecord, isSubFile bool)
	var appendProcesses func(parent *string, processes []db.ProcessRecord, subProcesses bool)
	var appendDocuments func(parent *string, documents []db.DocumentRecord, attachments bool)
	appendFiles = func(parent *string, files []db.FileRecord, subFiles bool) {
		for _, f := range files {
			m[f.RecordID] = recordNode{
				Parent: parent,
				Type:   db.RecordTypeFile,
				Title:  FileRecordTitle(f, subFiles),
			}
			appendFiles(&f.RecordID, f.Subfiles, true)
			appendProcesses(&f.RecordID, f.Processes, false)
			appendDocuments(&f.RecordID, f.Documents, false)
		}
	}
	appendProcesses = func(parent *string, processes []db.ProcessRecord, subProcesses bool) {
		for _, p := range processes {
			m[p.RecordID] = recordNode{
				Parent: parent,
				Type:   db.RecordTypeProcess,
				Title:  ProcessRecordTitle(p, subProcesses),
			}
			appendProcesses(&p.RecordID, p.Subprocesses, true)
			appendDocuments(&p.RecordID, p.Documents, false)
		}
	}
	appendDocuments = func(parent *string, documents []db.DocumentRecord, attachments bool) {
		for _, d := range documents {
			m[d.RecordID] = recordNode{
				Parent: parent,
				Type:   db.RecordTypeDocument,
				Title:  DocumentRecordTitle(d, attachments),
			}
			appendDocuments(&d.RecordID, d.Attachments, true)
		}
	}
	appendFiles(nil, r.Files, false)
	appendProcesses(nil, r.Processes, false)
	appendDocuments(nil, r.Documents, false)
	return m
}

// FindDiscrepancies compares a 0503 message with the appraisal of a 0501
// message and returns a list of any records missing in the 0503 message that
// were marked as to be archived in the appraisal and any surplus records
// included in the 0503 message.
//
// If a record is found be be missing or surplus, its child records will not be
// listed.
func FindDiscrepancies(
	message0501,
	message0503 db.Message,
) Discrepancies {
	// Gather data
	var result Discrepancies
	processID := message0501.MessageHead.ProcessID
	appraisals := make(map[string]db.Appraisal)
	// represents the root record and is implicitly marked 'A'.
	appraisals["root"] = db.Appraisal{Decision: db.AppraisalDecisionA}
	for _, a := range db.FindAppraisalsForProcess(context.Background(), processID) {
		appraisals[a.RecordID] = a
	}
	appraisedRootRecords := db.FindAllRootRecords(
		context.Background(), processID, db.MessageType0501,
	)
	appraisedRecords := getRecordMap(&appraisedRootRecords)
	submittedRootRecords := db.FindAllRootRecords(
		context.Background(), processID, db.MessageType0503,
	)
	submittedRecords := getRecordMap(&submittedRootRecords)
	// Check for objects missing from the 0503 message
L1:
	for id, r := range appraisedRecords {
		if _, ok := submittedRecords[id]; ok {
			// Not missing.
			continue
		}
		appraisal := appraisals[appraisedRecords.appraisableID(id)]
		if appraisal.Decision != db.AppraisalDecisionA {
			// Not missing.
			continue
		}
		for ancestorID := r.Parent; ancestorID != nil; ancestorID = appraisedRecords[*ancestorID].Parent {
			// Missing...
			if _, ok := submittedRecords[*ancestorID]; !ok {
				// ...but an ancestor is also missing and will be printed.
				continue L1
			}
		}
		// Missing.
		result.MissingRecords = append(result.MissingRecords, r.Title)
	}
	// Check for surplus objects in the 0503 message
L2:
	for id, r := range submittedRecords {
		appraisal := appraisals[submittedRecords.appraisableID(id)]
		if appraisal.Decision == db.AppraisalDecisionA {
			// Not surplus.
			continue
		}
		// Surplus...
		for ancestorID := r.Parent; ancestorID != nil; ancestorID = submittedRecords[*ancestorID].Parent {
			if appraisals[*ancestorID].Decision == db.AppraisalDecisionV {
				// ...but an ancestor was appraised "V" and will be printed.
				continue L2
			}
		}
		// Surplus.
		result.SurplusRecords = append(result.SurplusRecords, r.Title)
	}
	return result
}
