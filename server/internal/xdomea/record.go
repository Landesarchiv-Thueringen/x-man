package xdomea

import "lath/xman/internal/db"

func FileRecordTitle(f db.FileRecord, isSubFile bool) string {
	title := "Akte"
	if isSubFile {
		title = "Teilakte"
	}
	if f.GeneralMetadata != nil {
		if f.GeneralMetadata.RecordNumber != "" {
			title += " " + f.GeneralMetadata.RecordNumber
		}
		if f.GeneralMetadata.Subject != "" {
			title += ": " + f.GeneralMetadata.Subject
		}
	}
	return title
}

func ProcessRecordTitle(p db.ProcessRecord) string {
	title := "Vorgang"
	if p.GeneralMetadata != nil {
		if p.GeneralMetadata.RecordNumber != "" {
			title += " " + p.GeneralMetadata.RecordNumber
		}
		if p.GeneralMetadata.Subject != "" {
			title += ": " + p.GeneralMetadata.Subject
		}
	}
	return title
}
