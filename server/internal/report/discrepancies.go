package report

import (
	"context"
	"lath/xman/internal/db"
	"lath/xman/internal/xdomea"
)

type Discrepancies struct {
	xdomea.Discrepancies
	MissingPrimaryDocuments []string
}

func findDiscrepancies(
	message0501 *db.Message,
	message0503 db.Message,
) Discrepancies {
	var result Discrepancies
	if message0501 != nil {
		result.Discrepancies = xdomea.FindDiscrepancies(*message0501, message0503)
	}
	result.MissingPrimaryDocuments = findMissingPrimaryDocuments(message0503)
	return result
}

// findMissingPrimaryDocuments returns a list of primary documents that are
// referenced in the given xdomea message but are not present in the filesystem.
func findMissingPrimaryDocuments(message0503 db.Message) []string {
	var result []string
	rootRecords := db.FindAllRootRecords(
		context.Background(), message0503.MessageHead.ProcessID, db.MessageType0503,
	)
	primaryDocuments := xdomea.GetPrimaryDocuments(&rootRecords)
	_, missing := xdomea.FilterMissingPrimaryDocuments(
		message0503.MessageHead.ProcessID, primaryDocuments,
	)
	for _, r := range missing {
		result = append(result, r.Filename)
	}
	return result
}
