package xdomea

import (
	"lath/xman/internal/db"
)

func GetPrimaryDocuments(r *db.RootRecords) []db.PrimaryDocument {
	var d []db.PrimaryDocument
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

func GetPrimaryDocumentsForFile(r *db.FileRecord) []db.PrimaryDocument {
	var d []db.PrimaryDocument
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

func GetPrimaryDocumentsForProcess(r *db.ProcessRecord) []db.PrimaryDocument {
	var d []db.PrimaryDocument
	for _, c := range r.Subprocesses {
		d = append(d, GetPrimaryDocumentsForProcess(&c)...)
	}
	for _, c := range r.Documents {
		d = append(d, GetPrimaryDocumentsForDocument(&c)...)
	}
	return d
}

func GetPrimaryDocumentsForDocument(r *db.DocumentRecord) []db.PrimaryDocument {
	var d []db.PrimaryDocument
	for _, version := range r.Versions {
		for _, format := range version.Formats {
			d = append(d, format.PrimaryDocument)
		}
	}
	for _, c := range r.Attachments {
		d = append(d, GetPrimaryDocumentsForDocument(&c)...)
	}
	return d
}
