package report

import (
	"context"
	"lath/xman/internal/db"

	"github.com/google/uuid"
)

type ObjectAppraisalStats struct {
	Total             int
	Archived          int
	PartiallyArchived int
	Discarded         int
}

type AppraisalStats struct {
	Files     ObjectAppraisalStats
	Processes ObjectAppraisalStats
	Documents ObjectAppraisalStats
}

type appraisalMap = map[uuid.UUID]db.Appraisal

func fileHasDiscardedChildren(file db.FileRecord, m appraisalMap) bool {
	for _, s := range file.Subfiles {
		if m[s.RecordID].Decision != db.AppraisalDecisionA || fileHasDiscardedChildren(s, m) {
			return true
		}
	}
	for _, p := range file.Processes {
		if m[p.RecordID].Decision != db.AppraisalDecisionA || processHasDiscardedChildren(p, m) {
			return true
		}
	}
	return false
}

func processHasDiscardedChildren(process db.ProcessRecord, m appraisalMap) bool {
	for _, s := range process.Subprocesses {
		if m[s.RecordID].Decision != db.AppraisalDecisionA || processHasDiscardedChildren(s, m) {
			return true
		}
	}
	return false
}

// processFiles adds the given files to the appraisal stats.
//
// Files have to be on the root level of submission message.
func (a *AppraisalStats) processFiles(files []db.FileRecord, m appraisalMap) {
	for _, f := range files {
		switch d := m[f.RecordID].Decision; d {
		case db.AppraisalDecisionA:
			if fileHasDiscardedChildren(f, m) {
				a.Files.PartiallyArchived++
			} else {
				a.Files.Archived++
			}
		case db.AppraisalDecisionV:
			a.Files.Discarded++
		default:
			panic("unexpected appraisal decision: " + d)
		}
		a.Files.Total++
	}
}

// processProcesses adds the given processes to the appraisal stats.
//
// Processes have to be on the root level of submission message.
func (a *AppraisalStats) processProcesses(processes []db.ProcessRecord, m appraisalMap) {
	for _, p := range processes {
		switch d := m[p.RecordID].Decision; d {
		case db.AppraisalDecisionA:
			if processHasDiscardedChildren(p, m) {
				a.Processes.PartiallyArchived++
			} else {
				a.Processes.Archived++
			}
		case db.AppraisalDecisionV:
			a.Processes.Discarded++
		default:
			panic("unexpected appraisal decision: " + d)
		}
		a.Processes.Total++
	}
}

// processDocuments adds the given documents to the appraisal stats.
//
// Documents have to be on the root level of submission message.
func (a *AppraisalStats) processDocuments(
	documents []db.DocumentRecord,
) {
	// Documents on the root level cannot be appraised and are therefore
	// automatically archived.
	a.Documents.Archived += len(documents)
	a.Documents.Total += len(documents)
}

func getAppraisalsMap(processID uuid.UUID) appraisalMap {
	m := make(appraisalMap)
	appraisals := db.FindAppraisalsForProcess(context.Background(), processID)
	for _, a := range appraisals {
		m[a.RecordID] = a
	}
	return m
}

func getAppraisalStats(ctx context.Context, message db.Message) (a AppraisalStats) {
	m := getAppraisalsMap(message.MessageHead.ProcessID)
	rootRecords := db.FindAllRootRecords(ctx, message.MessageHead.ProcessID, message.MessageType)
	a.processFiles(rootRecords.Files, m)
	a.processProcesses(rootRecords.Processes, m)
	a.processDocuments(rootRecords.Documents)
	return
}
