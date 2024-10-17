package report

import (
	"context"
	"lath/xman/internal/db"

	"github.com/google/uuid"
)

type ObjectAppraisalStats struct {
	Total             int
	Offered           int
	Archived          int
	PartiallyArchived int
	Discarded         int
	Missing           int
	Surplus           int
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
func (a *AppraisalStats) processFiles(offeredFiles, submittedFiles []db.FileRecord, m appraisalMap) {
	offeredFilesMap := make(map[uuid.UUID]bool)
	for _, f := range offeredFiles {
		offeredFilesMap[f.RecordID] = true
	}
	submittedFilesMap := make(map[uuid.UUID]bool)
	for _, f := range submittedFiles {
		submittedFilesMap[f.RecordID] = true
	}
	for _, f := range offeredFiles {
		switch d := m[f.RecordID].Decision; d {
		case db.AppraisalDecisionA:
			if !submittedFilesMap[f.RecordID] {
				a.Files.Missing++
			} else if fileHasDiscardedChildren(f, m) {
				a.Files.PartiallyArchived++
			} else {
				a.Files.Archived++
			}
		case db.AppraisalDecisionV:
			if submittedFilesMap[f.RecordID] {
				a.Files.Surplus++
			} else {
				a.Files.Discarded++
			}
		default:
			panic("unexpected appraisal decision: " + d)
		}
		a.Files.Offered++
		a.Files.Total++
	}
	for _, f := range submittedFiles {
		if !offeredFilesMap[f.RecordID] {
			a.Files.Surplus++
			a.Files.Total++
		}
	}
}

// processProcesses adds the given processes to the appraisal stats.
//
// Processes have to be on the root level of submission message.
func (a *AppraisalStats) processProcesses(offeredProcesses, submittedProcesses []db.ProcessRecord, m appraisalMap) {
	offeredProcessesMap := make(map[uuid.UUID]bool)
	for _, p := range offeredProcesses {
		offeredProcessesMap[p.RecordID] = true
	}
	submittedProcessesMap := make(map[uuid.UUID]bool)
	for _, p := range submittedProcesses {
		submittedProcessesMap[p.RecordID] = true
	}
	for _, p := range offeredProcesses {
		switch d := m[p.RecordID].Decision; d {
		case db.AppraisalDecisionA:
			if !submittedProcessesMap[p.RecordID] {
				a.Processes.Missing++
			} else if processHasDiscardedChildren(p, m) {
				a.Processes.PartiallyArchived++
			} else {
				a.Processes.Archived++
			}
		case db.AppraisalDecisionV:
			if submittedProcessesMap[p.RecordID] {
				a.Processes.Surplus++
			} else {
				a.Processes.Discarded++
			}
		default:
			panic("unexpected appraisal decision: " + d)
		}
		a.Processes.Offered++
		a.Processes.Total++
	}
	for _, f := range submittedProcesses {
		if !offeredProcessesMap[f.RecordID] {
			a.Processes.Surplus++
			a.Processes.Total++
		}
	}
}

// processDocuments adds the given documents to the appraisal stats.
//
// Documents have to be on the root level of submission message.
func (a *AppraisalStats) processDocuments(
	offeredDocuments, submittedDocuments []db.DocumentRecord,
) {
	offeredDocumentsMap := make(map[uuid.UUID]bool)
	for _, d := range offeredDocuments {
		offeredDocumentsMap[d.RecordID] = true
	}
	submittedDocumentsMap := make(map[uuid.UUID]bool)
	for _, d := range submittedDocuments {
		submittedDocumentsMap[d.RecordID] = true
	}
	// Documents on the root level cannot be appraised and are therefore
	// automatically archived.
	for _, d := range offeredDocuments {
		if submittedDocumentsMap[d.RecordID] {
			a.Documents.Archived++
		} else {
			a.Documents.Missing++
		}
		a.Documents.Offered++
		a.Documents.Total++
	}
	for _, f := range submittedDocuments {
		if !offeredDocumentsMap[f.RecordID] {
			a.Documents.Surplus++
			a.Documents.Total++
		}
	}
}

func getAppraisalsMap(processID uuid.UUID) appraisalMap {
	m := make(appraisalMap)
	appraisals := db.FindAppraisalsForProcess(context.Background(), processID)
	for _, a := range appraisals {
		m[a.RecordID] = a
	}
	return m
}

func getAppraisalStats(ctx context.Context, message0501, message0503 db.Message) (a AppraisalStats) {
	m := getAppraisalsMap(message0501.MessageHead.ProcessID)
	offeredRootRecords := db.FindAllRootRecords(ctx, message0501.MessageHead.ProcessID, message0501.MessageType)
	submittedRootRecords := db.FindAllRootRecords(ctx, message0503.MessageHead.ProcessID, message0503.MessageType)
	a.processFiles(offeredRootRecords.Files, submittedRootRecords.Files, m)
	a.processProcesses(offeredRootRecords.Processes, submittedRootRecords.Processes, m)
	a.processDocuments(offeredRootRecords.Documents, submittedRootRecords.Documents)
	return
}
