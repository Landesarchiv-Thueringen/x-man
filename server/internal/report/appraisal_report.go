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

type appraisalReportData struct {
	Process        db.SubmissionProcess
	AppraisalStats appraisalStats
	AppraisalInfo  []AppraisalStructure
}

// GetAppraisalReport sends process data to the report service and returns the generated PDF.
func GetAppraisalReport(
	ctx context.Context,
	process db.SubmissionProcess,
) (contentLength int64, contentType string, body io.Reader) {
	values, err := getAppraisalReportData(ctx, process)
	if err != nil {
		panic(err)
	}
	jsonValue, _ := json.Marshal(values)
	resp, err := http.Post(
		os.Getenv("REPORT_URL")+"/render/appraisal", "application/json",
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

// getAppraisalReportData accumulates process data for use by the report service.
func getAppraisalReportData(
	ctx context.Context,
	process db.SubmissionProcess,
) (reportData appraisalReportData, err error) {
	message, ok := db.FindMessage(ctx, process.ProcessID, db.MessageType0501)
	if !ok {
		return reportData, errors.New("tried to get appraisal report of process without 0501 message")
	}
	reportData = appraisalReportData{
		Process:        process,
		AppraisalStats: getAppraisalStats(ctx, message, nil),
		AppraisalInfo:  appraisalInfo(ctx, process),
	}
	if os.Getenv("DEBUG_MODE") == "true" {
		writeToFile(reportData, "/debug-data/appraisal-data.json")
	}
	return
}
