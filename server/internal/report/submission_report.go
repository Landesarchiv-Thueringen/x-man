package report

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"lath/xman/internal/db"
	"net/http"
	"os"
)

type SubmissionReportData struct {
	Process          db.SubmissionProcess
	ArchivePackages  []ArchivePackageStructure
	Message0503Stats *ContentStats
	AppraisalStats   *appraisalStats
	FileStats        FileStats
	Discrepancies    Discrepancies
}

// GetSubmissionReport sends process data to the report service and returns the generated PDF.
func GetSubmissionReport(
	ctx context.Context,
	process db.SubmissionProcess,
) (contentLength int64, contentType string, body io.Reader) {
	values, err := getSubmissionReportData(ctx, process)
	if err != nil {
		panic(err)
	}
	jsonValue, _ := json.Marshal(values)
	resp, err := http.Post(
		os.Getenv("REPORT_URL")+"/render/submission", "application/json",
		bytes.NewBuffer(jsonValue),
	)
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

// getSubmissionReportData accumulates process data for use by the report service.
func getSubmissionReportData(
	ctx context.Context,
	process db.SubmissionProcess,
) (reportData SubmissionReportData, err error) {
	messages := make(map[db.MessageType]db.Message)
	for _, m := range db.FindMessagesForProcess(ctx, process.ProcessID) {
		messages[m.MessageType] = m
	}
	message0503, found := messages[db.MessageType0503]
	if !found {
		return reportData, errors.New("tried to get report of process without 0503 message")
	}
	reportData.Process = process

	if message0501, ok := messages[db.MessageType0501]; ok {
		appraisalStats := getAppraisalStats(ctx, message0501, &message0503)
		reportData.AppraisalStats = &appraisalStats
		reportData.Discrepancies = findDiscrepancies(&message0501, message0503)
	} else {
		reportData.Discrepancies = findDiscrepancies(nil, message0503)
	}
	reportData.ArchivePackages = archivePackagesInfo(ctx, process)
	messageStats := getMessageContentStats(ctx, message0503)
	reportData.Message0503Stats = &messageStats
	reportData.FileStats = getFileStats(ctx, process)
	if os.Getenv("DEBUG_MODE") == "true" {
		writeToFile(reportData, "/debug-data/submission-data.json")
	}
	return
}

// writeToFile writes accumulated report data to a json file for use for
// development of the report service.
func writeToFile(reportData any, fileName string) {
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
