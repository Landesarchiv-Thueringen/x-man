package report

import (
	"errors"
	"fmt"
	"lath/xman/internal/db"
	"time"
)

type ReportData struct {
	Institution  string
	CreationTime string
	FileStats    FileStats
	Message      db.Message
}

type FileStats struct {
	Total      uint
	ByFileType map[string]uint
}

func GetReportData(processID string) (ReportData, error) {
	if processID == "" {
		panic("called GetReportData with empty string")
	}
	var reportData ReportData
	process, found := db.GetProcessByXdomeaID(processID)
	if !found {
		panic(fmt.Sprintf("process not found: %v", processID))
	}
	reportData.Institution = process.Agency.Name
	if process.Message0503ID == nil {
		return reportData, errors.New("tried to get report of process with Message0503ID == nil")
	}
	documents, err := db.GetAllPrimaryDocumentsWithFormatVerification(*process.Message0503ID)
	if err != nil {
		return reportData, err
	}
	reportData.FileStats.ByFileType = make(map[string]uint)
	for _, document := range documents {
		mimeType := getMimeType(document)
		reportData.FileStats.ByFileType[mimeType] += 1
		reportData.FileStats.Total += 1
	}
	message, found := db.GetCompleteMessageByID(*process.Message0501ID)
	if !found {
		return reportData, fmt.Errorf("message not found: %v", *process.Message0501ID)
	}
	reportData.Message = message
	reportData.CreationTime = formatTime(message.MessageHead.CreationTime)
	return reportData, nil
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

func formatTime(input string) string {
	layout := "2006-01-02T15:04:05"
	t, _ := time.Parse(layout, input)
	return t.Local().Format("02.01.2006 15:04 Uhr")
}
