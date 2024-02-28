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
	Files        ObjectAppraisalStats
	SubFiles     ObjectAppraisalStats
	Processes    ObjectAppraisalStats
	SubProcesses ObjectAppraisalStats
	Documents    ObjectAppraisalStats
	Attachments  ObjectAppraisalStats
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

func getAppraisalStats(message db.Message) (a AppraisalStats) {
	a.processFiles(message.FileRecordObjects, false)
	a.processProcesses(message.ProcessRecordObjects, false)
	// Treat all root-level documents as appraised to "A" since documents always
	// inherit their appraisal from their parent element (which in this case is
	// the message itself).
	code := "A"
	a.processDocuments(message.DocumentRecordObjects, false, db.ArchiveMetadata{AppraisalCode: &code})
	return
}

func getFileAppraisalStats(file db.FileRecordObject) (a AppraisalStats) {
	a.processFiles(file.SubFileRecordObjects, true)
	a.processProcesses(file.ProcessRecordObjects, false)
	a.processDocuments(file.DocumentRecordObjects, false, *file.ArchiveMetadata)
	return
}

func getProcessAppraisalStats(process db.ProcessRecordObject) (a AppraisalStats) {
	a.processProcesses(process.SubProcessRecordObjects, false)
	a.processDocuments(process.DocumentRecordObjects, false, *process.ArchiveMetadata)
	return
}
