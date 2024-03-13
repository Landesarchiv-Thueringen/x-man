package report

import (
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

func (a *AppraisalStats) processFiles(files []db.FileRecordObject, isSubLevel bool, m appraisalMap) {
	for _, file := range files {
		if isSubLevel {
			a.SubFiles.addObject(m[file.XdomeaID])
		} else {
			a.Files.addObject(m[file.XdomeaID])
		}
		a.processFiles(file.SubFileRecordObjects, true, m)
		a.processProcesses(file.ProcessRecordObjects, false, m)
		a.processDocuments(file.DocumentRecordObjects, false, m[file.XdomeaID])
	}
}

func (a *AppraisalStats) processProcesses(processes []db.ProcessRecordObject, isSubLevel bool, m appraisalMap) {
	for _, process := range processes {
		if isSubLevel {
			a.SubProcesses.addObject(m[process.XdomeaID])
		} else {
			a.Processes.addObject(m[process.XdomeaID])
		}
		a.processProcesses(process.SubProcessRecordObjects, false, m)
		a.processDocuments(process.DocumentRecordObjects, false, m[process.XdomeaID])
	}
}

func (a *AppraisalStats) processDocuments(
	documents []db.DocumentRecordObject,
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

func getAppraisalsMap(processID string) appraisalMap {
	m := make(appraisalMap)
	appraisals := db.GetAppraisalsForProcess(processID)
	for _, a := range appraisals {
		m[a.RecordObjectID] = a
	}
	return m
}

func getAppraisalStats(message db.Message) (a AppraisalStats) {
	m := getAppraisalsMap(message.MessageHead.ProcessID)
	a.processFiles(message.FileRecordObjects, false, m)
	a.processProcesses(message.ProcessRecordObjects, false, m)
	// Treat all root-level documents as appraised to "A" since documents always
	// inherit their appraisal from their parent element (which in this case is
	// the message itself).
	a.processDocuments(message.DocumentRecordObjects, false, db.Appraisal{Decision: "A"})
	a.checkForDeviatingAppraisals("A")
	return
}
