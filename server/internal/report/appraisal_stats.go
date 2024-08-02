package report

import (
	"context"
	"lath/xman/internal/db"

	"github.com/google/uuid"
)

type ObjectAppraisalStats struct {
	Total     uint
	Archived  uint
	Discarded uint
}

func (objectStats *ObjectAppraisalStats) addObject(a db.Appraisal) {
	switch a.Decision {
	case "A":
		objectStats.Archived += 1
	case "V":
		objectStats.Discarded += 1
	default:
		panic("unexpected appraisal code: " + a.Decision)
	}
	objectStats.Total += 1
}

type AppraisalStats struct {
	// HasDeviatingAppraisals indicates whether there are any appraisals within
	// the stats that differ from the parent element's own appraisal code.
	//
	// If the parent is a message, appraisal code "A" is used for comparison.
	HasDeviatingAppraisals bool
	Files                  ObjectAppraisalStats
	SubFiles               ObjectAppraisalStats
	Processes              ObjectAppraisalStats
	SubProcesses           ObjectAppraisalStats
	Documents              ObjectAppraisalStats
	Attachments            ObjectAppraisalStats
}

type appraisalMap = map[uuid.UUID]db.Appraisal

func (a *AppraisalStats) processFiles(files []db.FileRecord, isSubLevel bool, m appraisalMap) {
	for _, file := range files {
		if isSubLevel {
			a.SubFiles.addObject(m[file.RecordID])
		} else {
			a.Files.addObject(m[file.RecordID])
		}
		a.processFiles(file.Subfiles, true, m)
		a.processProcesses(file.Processes, false, m)
		a.processDocuments(file.Documents, false, m[file.RecordID])
	}
}

func (a *AppraisalStats) processProcesses(processes []db.ProcessRecord, isSubLevel bool, m appraisalMap) {
	for _, process := range processes {
		if isSubLevel {
			a.SubProcesses.addObject(m[process.RecordID])
		} else {
			a.Processes.addObject(m[process.RecordID])
		}
		a.processProcesses(process.Subprocesses, false, m)
		a.processDocuments(process.Documents, false, m[process.RecordID])
	}
}

func (a *AppraisalStats) processDocuments(
	documents []db.DocumentRecord,
	isSubLevel bool,
	appraisal db.Appraisal,
) {
	for _, document := range documents {
		if isSubLevel {
			a.Attachments.addObject(appraisal)
		} else {
			a.Documents.addObject(appraisal)
		}
		a.processDocuments(document.Attachments, true, appraisal)
	}
}

// checkForDeviatingAppraisals checks if the stats object has processed  any
// elements with an appraisal code different from the one given and sets
// HasDeviatingAppraisals accordingly.
//
// Should be called after all objects have been processed.
func (a *AppraisalStats) checkForDeviatingAppraisals(appraisalCode string) {
	switch appraisalCode {
	case "A":
		a.HasDeviatingAppraisals = a.Files.Discarded+a.SubFiles.Discarded+
			a.Processes.Discarded+a.SubProcesses.Discarded > 0
	case "V":
		a.HasDeviatingAppraisals = a.Files.Archived+a.SubFiles.Archived+
			a.Processes.Archived+a.SubProcesses.Archived > 0
	default:
		panic("unexpected appraisal code: " + appraisalCode)
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

func getAppraisalStats(ctx context.Context, message db.Message) (a AppraisalStats) {
	m := getAppraisalsMap(message.MessageHead.ProcessID)
	rootRecords := db.FindAllRootRecords(ctx, message.MessageHead.ProcessID, message.MessageType)
	a.processFiles(rootRecords.Files, false, m)
	a.processProcesses(rootRecords.Processes, false, m)
	// Treat all root-level documents as appraised to "A" since documents always
	// inherit their appraisal from their parent element (which in this case is
	// the message itself).
	a.processDocuments(rootRecords.Documents, false, db.Appraisal{Decision: "A"})
	a.checkForDeviatingAppraisals("A")
	return
}
