package report

import (
	"errors"
	"fmt"
	"lath/xman/internal/db"
	"os"
	"path"
)

type ReportData struct {
	Process          db.Process
	Message0501Stats *MessageStats
	Message0503Stats *MessageStats
	AppraisalStats   *AppraisalStats
	FileStats        FileStats
	Message0501      *db.Message
	Message0503      db.Message
}

type MessageStats struct {
	TotalFiles        uint
	TotalSubFiles     uint
	TotalProcesses    uint
	TotalSubProcesses uint
	TotalDocuments    uint
	TotalAttachments  uint
}

type AppraisalStats struct {
	Files        ObjectAppraisalStats
	SubFiles     ObjectAppraisalStats
	Processes    ObjectAppraisalStats
	SubProcesses ObjectAppraisalStats
	Documents    ObjectAppraisalStats
	Attachments  ObjectAppraisalStats
}

type ObjectAppraisalStats struct {
	Archived  uint
	Discarded uint
}

type FileStats struct {
	TotalFiles      uint
	TotalBytes      uint64
	FilesByFileType map[string]uint
}

func GetReportData(process db.Process) (reportData ReportData, err error) {
	// var reportData ReportData
	if process.Message0503ID == nil {
		return reportData, errors.New("tried to get report of process with Message0503ID == nil")
	}
	reportData.Process = process
	if process.Message0501ID != nil {
		message0501, found := db.GetCompleteMessageByID(*process.Message0501ID)
		if !found {
			panic(fmt.Sprintf("message not found: %v", *process.Message0501ID))
		}
		messageStats := getMessageStats(message0501)
		reportData.Message0501Stats = &messageStats
		appraisalStats := getAppraisalStats(message0501)
		reportData.AppraisalStats = &appraisalStats
		reportData.Message0501 = &message0501
	}
	message0503, found := db.GetCompleteMessageByID(*process.Message0503ID)
	if !found {
		panic(fmt.Sprintf("message not found: %v", *process.Message0503ID))
	}
	messageStats := getMessageStats(message0503)
	reportData.Message0503Stats = &messageStats
	reportData.Message0503 = message0503
	reportData.FileStats = getFileStats(process)
	// writeToFile(reportData, "/data/data.json")
	return
}

func getMessageStats(message db.Message) (messageStats MessageStats) {
	var processFiles func(files []db.FileRecordObject, isSubLevel bool)
	var processProcesses func(files []db.ProcessRecordObject, isSubLevel bool)
	var processDocument func(files []db.DocumentRecordObject, isSubLevel bool)
	processFiles = func(files []db.FileRecordObject, isSubLevel bool) {
		for _, file := range files {
			if isSubLevel {
				messageStats.TotalSubFiles += 1
			} else {
				messageStats.TotalFiles += 1
			}
			processFiles(file.SubFileRecordObjects, true)
			processProcesses(file.ProcessRecordObjects, false)
			processDocument(file.DocumentRecordObjects, false)
		}
	}
	processProcesses = func(processes []db.ProcessRecordObject, isSubLevel bool) {
		for _, process := range processes {
			if isSubLevel {
				messageStats.TotalSubProcesses += 1
			} else {
				messageStats.TotalProcesses += 1
			}
			processProcesses(process.SubProcessRecordObjects, false)
			processDocument(process.DocumentRecordObjects, false)
		}
	}
	processDocument = func(documents []db.DocumentRecordObject, isSubLevel bool) {
		for _, document := range documents {
			if isSubLevel {
				messageStats.TotalAttachments += 1
			} else {
				messageStats.TotalDocuments += 1
			}
			processDocument(document.Attachments, true)
		}
	}
	processFiles(message.FileRecordObjects, false)
	return
}

func getAppraisalStats(message db.Message) (appraisalStats AppraisalStats) {
	var processFiles func(files []db.FileRecordObject, isSubLevel bool)
	var processProcesses func(files []db.ProcessRecordObject, isSubLevel bool)
	var processDocument func(files []db.DocumentRecordObject, isSubLevel bool, archiveMetadata db.ArchiveMetadata)
	addObject := func(objectStats *ObjectAppraisalStats, archiveMetadata db.ArchiveMetadata) {
		switch *archiveMetadata.AppraisalCode {
		case "A":
			objectStats.Archived += 1
		case "V":
			objectStats.Discarded += 1
		default:
			panic("unexpected appraisal code: " + *archiveMetadata.AppraisalCode)
		}
	}
	processFiles = func(files []db.FileRecordObject, isSubLevel bool) {
		for _, file := range files {
			if isSubLevel {
				addObject(&appraisalStats.SubFiles, *file.ArchiveMetadata)
			} else {
				addObject(&appraisalStats.Files, *file.ArchiveMetadata)
			}
			processFiles(file.SubFileRecordObjects, true)
			processProcesses(file.ProcessRecordObjects, false)
			processDocument(file.DocumentRecordObjects, false, *file.ArchiveMetadata)
		}
	}
	processProcesses = func(processes []db.ProcessRecordObject, isSubLevel bool) {
		for _, process := range processes {
			if isSubLevel {
				addObject(&appraisalStats.SubProcesses, *process.ArchiveMetadata)
			} else {
				addObject(&appraisalStats.Processes, *process.ArchiveMetadata)
			}
			processProcesses(process.SubProcessRecordObjects, false)
			processDocument(process.DocumentRecordObjects, false, *process.ArchiveMetadata)
		}
	}
	processDocument = func(documents []db.DocumentRecordObject, isSubLevel bool, archiveMetadata db.ArchiveMetadata) {
		for _, document := range documents {
			if isSubLevel {
				addObject(&appraisalStats.Attachments, archiveMetadata)
			} else {
				addObject(&appraisalStats.Documents, archiveMetadata)
			}
			processDocument(document.Attachments, true, archiveMetadata)
		}
	}
	processFiles(message.FileRecordObjects, false)
	return
}

func getFileStats(process db.Process) (fileStats FileStats) {
	documents := db.GetAllPrimaryDocumentsWithFormatVerification(*process.Message0503ID)
	fileStats.FilesByFileType = make(map[string]uint)
	for _, document := range documents {
		mimeType := getMimeType(document)
		fileStats.FilesByFileType[mimeType] += 1
		fileStats.TotalFiles += 1
		fileSize := getFileSize(path.Join(process.Message0503.StoreDir, document.FileName))
		fileStats.TotalBytes += fileSize
	}
	return
}

func getFileSize(path string) uint64 {
	fi, err := os.Stat(path)
	if err != nil {
		panic(err)
	}
	return uint64(fi.Size())
}

func getMimeType(document db.PrimaryDocument) string {
	if document.FormatVerification == nil {
		return ""
	}
	for _, feature := range document.FormatVerification.Features {
		if feature.Key == "mimeType" {
			return feature.Values[0].Value
		}
	}
	return ""
}

// func writeToFile(reportData ReportData, fileName string) {
// 	f, err := os.Create(fileName)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer f.Close()
// 	j, _ := json.MarshalIndent(reportData, "", "\t")
// 	_, err = fmt.Fprintf(f, "%s \n", j)
// 	if err != nil {
// 		panic(err)
// 	}
// }
