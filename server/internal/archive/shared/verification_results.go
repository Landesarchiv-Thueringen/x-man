package shared

import (
	"context"
	"encoding/json"
	"lath/xman/internal/db"
)

func GenerateVerificationResults(
	processID string,
	archivePackage db.ArchivePackage,
) ([]byte, bool) {
	results := make(map[string]db.FormatVerification)
	for _, d := range archivePackage.PrimaryDocuments {
		data, ok := db.FindPrimaryDocumentData(context.Background(), processID, d.Filename)
		if ok && data.FormatVerification != nil {
			results[d.Filename] = *data.FormatVerification
		}
	}
	bytes, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		panic(err)
	}
	return bytes, len(results) > 0
}
