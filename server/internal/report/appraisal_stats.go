package report

import (
	"lath/xman/internal/db"
)

type ObjectAppraisalStats struct {
	Total     uint
	Archived  uint
	Discarded uint
}

func (objectStats *ObjectAppraisalStats) addObject(archiveMetadata db.ArchiveMetadata) {
	switch *archiveMetadata.AppraisalCode {
	case "A":
		objectStats.Archived += 1
	case "V":
		objectStats.Discarded += 1
	default:
		panic("unexpected appraisal code: " + *archiveMetadata.AppraisalCode)
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

func (a *AppraisalStats) processFiles(files []db.FileRecordObject, isSubLevel bool) {
	for _, file := range files {
		if isSubLevel {
			a.SubFiles.addObject(*file.ArchiveMetadata)
		} else {
			a.Files.addObject(*file.ArchiveMetadata)
		}
		a.processFiles(file.SubFileRecordObjects, true)
		a.processProcesses(file.ProcessRecordObjects, false)
		a.processDocuments(file.DocumentRecordObjects, false, *file.ArchiveMetadata)
	}
}

func (a *AppraisalStats) processProcesses(processes []db.ProcessRecordObject, isSubLevel bool) {
	for _, process := range processes {
		if isSubLevel {
			a.SubProcesses.addObject(*process.ArchiveMetadata)
		} else {
			a.Processes.addObject(*process.ArchiveMetadata)
		}
		a.processProcesses(process.SubProcessRecordObjects, false)
		a.processDocuments(process.DocumentRecordObjects, false, *process.ArchiveMetadata)
	}
}

func (a *AppraisalStats) processDocuments(
	documents []db.DocumentRecordObject,
	isSubLevel bool,
	archiveMetadata db.ArchiveMetadata,
) {
	for _, document := range documents {
		if isSubLevel {
			a.Attachments.addObject(archiveMetadata)
		} else {
			a.Documents.addObject(archiveMetadata)
		}
		a.processDocuments(document.Attachments, true, archiveMetadata)
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

func getAppraisalStats(message db.Message) (a AppraisalStats) {
	a.processFiles(message.FileRecordObjects, false)
	a.processProcesses(message.ProcessRecordObjects, false)
	// Treat all root-level documents as appraised to "A" since documents always
	// inherit their appraisal from their parent element (which in this case is
	// the message itself).
	code := "A"
	a.processDocuments(message.DocumentRecordObjects, false, db.ArchiveMetadata{AppraisalCode: &code})
	a.checkForDeviatingAppraisals(code)
	return
}

func getFileAppraisalStats(file db.FileRecordObject) (a AppraisalStats) {
	a.processFiles(file.SubFileRecordObjects, true)
	a.processProcesses(file.ProcessRecordObjects, false)
	a.processDocuments(file.DocumentRecordObjects, false, *file.ArchiveMetadata)
	a.checkForDeviatingAppraisals(*file.ArchiveMetadata.AppraisalCode)
	return
}

func getProcessAppraisalStats(process db.ProcessRecordObject) (a AppraisalStats) {
	a.processProcesses(process.SubProcessRecordObjects, false)
	a.processDocuments(process.DocumentRecordObjects, false, *process.ArchiveMetadata)
	a.checkForDeviatingAppraisals(*process.ArchiveMetadata.AppraisalCode)
	return
}
