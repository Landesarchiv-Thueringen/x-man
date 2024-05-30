package archive

import "lath/xman/internal/db"

func GetFileRecordTitle(f db.FileRecord) string {
	title := "Akte"
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

func GetProcessRecordTitle(p db.ProcessRecord) string {
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
