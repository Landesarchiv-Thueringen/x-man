package xdomea

import (
	"context"
	"lath/xman/internal/db"

	"github.com/google/uuid"
)

// GetPrimaryDocuments traverses the given root records and returns all included
// primary documents.
func GetPrimaryDocuments(r *db.RootRecords) []db.PrimaryDocumentContext {
	if r == nil {
		return []db.PrimaryDocumentContext{}
	}
	var d []db.PrimaryDocumentContext
	for _, c := range r.Files {
		d = append(d, GetPrimaryDocumentsForFile(&c)...)
	}
	for _, c := range r.Processes {
		d = append(d, GetPrimaryDocumentsForProcess(&c)...)
	}
	for _, c := range r.Documents {
		d = append(d, GetPrimaryDocumentsForDocument(&c)...)
	}
	return d
}

func GetPrimaryDocumentsForFile(r *db.FileRecord) []db.PrimaryDocumentContext {
	var d []db.PrimaryDocumentContext
	for _, c := range r.Subfiles {
		d = append(d, GetPrimaryDocumentsForFile(&c)...)
	}
	for _, c := range r.Processes {
		d = append(d, GetPrimaryDocumentsForProcess(&c)...)
	}
	for _, c := range r.Documents {
		d = append(d, GetPrimaryDocumentsForDocument(&c)...)
	}
	return d
}

func GetPrimaryDocumentsForProcess(r *db.ProcessRecord) []db.PrimaryDocumentContext {
	var d []db.PrimaryDocumentContext
	for _, c := range r.Subprocesses {
		d = append(d, GetPrimaryDocumentsForProcess(&c)...)
	}
	for _, c := range r.Documents {
		d = append(d, GetPrimaryDocumentsForDocument(&c)...)
	}
	return d
}

func GetPrimaryDocumentsForDocument(r *db.DocumentRecord) []db.PrimaryDocumentContext {
	var d []db.PrimaryDocumentContext
	for _, version := range r.Versions {
		for _, format := range version.Formats {
			d = append(d, db.PrimaryDocumentContext{
				PrimaryDocument: format.PrimaryDocument,
				RecordID:        r.RecordID,
			})
		}
	}
	for _, c := range r.Attachments {
		d = append(d, GetPrimaryDocumentsForDocument(&c)...)
	}
	return d
}

// FilterMissingPrimaryDocuments segregates primary documents based on whether
// their files are found in the filesystem.
func FilterMissingPrimaryDocuments(
	processID uuid.UUID,
	primaryDocument []db.PrimaryDocumentContext,
) (found, missing []db.PrimaryDocumentContext) {
	primaryDocumentsData := db.FindPrimaryDocumentsDataForProcess(
		context.Background(), processID,
	)
	dataMap := make(map[string]bool)
	for _, d := range primaryDocumentsData {
		dataMap[d.Filename] = true
	}
	for _, d := range primaryDocument {
		if dataMap[d.Filename] {
			found = append(found, d)
		} else {
			missing = append(missing, d)
		}
	}
	return
}
