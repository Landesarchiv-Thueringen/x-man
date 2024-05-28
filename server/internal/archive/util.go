package archive

import "lath/xman/internal/db"

// GetCombinedLifetime returns a string representation of lifetime start and end.
func GetCombinedLifetime(l *db.Lifetime) string {
	if l != nil {
		if l.Start != "" && l.End != "" {
			return l.Start + " - " + l.End
		} else if l.Start != "" {
			return l.Start + " - "
		} else if l.End != "" {
			return " - " + l.End
		}
	}
	return ""
}

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
