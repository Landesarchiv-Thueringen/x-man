package report

import (
	"errors"
	"fmt"
	"lath/xman/internal/db"
	"time"

	"github.com/google/uuid"
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

func GetReportData(process db.Process) (ReportData, error) {
	var reportData ReportData
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
	var messageID uuid.UUID
	if process.Message0501ID != nil {
		messageID = *process.Message0501ID
	} else {
		messageID = *process.Message0503ID
	}
	message, found := db.GetCompleteMessageByID(messageID)
	if !found {
		panic(fmt.Sprintf("message not found: %v", messageID))
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
