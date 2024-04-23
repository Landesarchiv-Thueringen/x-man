package report

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"lath/xman/internal/db"
	"net/http"
	"os"
)

type ReportData struct {
	Process          db.Process
	Content          []ContentObject
	Message0501Stats *ContentStats
	Message0503Stats *ContentStats
	AppraisalStats   *AppraisalStats
	FileStats        FileStats
	Message0501      *db.Message
	Message0503      db.Message
}

// GetReport sends process data to the report service and returns the generated PDF.
func GetReport(process db.Process) (contentLength int64, contentType string, body io.Reader) {
	values, err := getReportData(process)
	if err != nil {
		panic(err)
	}
	jsonValue, _ := json.Marshal(values)
	resp, err := http.Post("http://report/render", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		panic(err)
	} else if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		println(string(body))
		panic(fmt.Sprintf("status code: %d", resp.StatusCode))
	}
	contentLength = resp.ContentLength
	contentType = resp.Header.Get("Content-Type")
	body = resp.Body
	return
}

// getReportData accumulates process data for use by the report service.
func getReportData(process db.Process) (reportData ReportData, err error) {
	if process.Message0503ID == nil {
		return reportData, errors.New("tried to get report of process with Message0503ID == nil")
	}
	reportData.Process = process
	var message0501 db.Message
	if process.Message0501ID != nil {
		var found bool
		message0501, found = db.GetCompleteMessageByID(*process.Message0501ID)
		if !found {
			panic(fmt.Sprintf("message not found: %v", *process.Message0501ID))
		}
		messageStats := getMessageContentStats(message0501)
		reportData.Message0501Stats = &messageStats
		appraisalStats := getAppraisalStats(message0501)
		reportData.AppraisalStats = &appraisalStats
		reportData.Message0501 = &message0501
	}
	message0503, found := db.GetCompleteMessageByID(*process.Message0503ID)
	if !found {
		panic(fmt.Sprintf("message not found: %v", *process.Message0503ID))
	}
	reportData.Content = getContentObjects(message0503)
	messageStats := getMessageContentStats(message0503)
	reportData.Message0503Stats = &messageStats
	reportData.Message0503 = message0503
	reportData.FileStats = getFileStats(process)
	if os.Getenv("DEBUG_MODE") == "true" {
		writeToFile(reportData, "/debug-data/data.json")
	}
	return
}

// writeToFile writes accumulated report data to a json file for use for
// development of the report service.
func writeToFile(reportData ReportData, fileName string) {
	f, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	j, _ := json.MarshalIndent(reportData, "", "\t")
	_, err = fmt.Fprintf(f, "%s \n", j)
	if err != nil {
		panic(err)
	}
}
