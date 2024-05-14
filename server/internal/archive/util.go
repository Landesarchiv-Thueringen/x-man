package archive

import "lath/xman/internal/db"

// GetCombinedLifetime returns a string representation of lifetime start and end.
func GetCombinedLifetime(l *db.Lifetime) string {
	if l != nil {
		if l.Start != nil && l.End != nil {
			return *l.Start + " - " + *l.End
		} else if l.Start != nil {
			return *l.Start + " - "
		} else if l.End != nil {
			return " - " + *l.End
		}
	}
	return ""
}

func GetFileRecordTitle(f db.FileRecord) string {
	title := "Akte"
	if f.GeneralMetadata != nil {
		if f.GeneralMetadata.RecordNumber != nil {
			title += " " + *f.GeneralMetadata.RecordNumber
		}
		if f.GeneralMetadata.Subject != nil {
			title += ": " + *f.GeneralMetadata.Subject
		}
	}
	return title
}

func GetProcessRecordTitle(p db.ProcessRecord) string {
	title := "Vorgang"
	if p.GeneralMetadata != nil {
		if p.GeneralMetadata.RecordNumber != nil {
			title += " " + *p.GeneralMetadata.RecordNumber
		}
		if p.GeneralMetadata.Subject != nil {
			title += ": " + *p.GeneralMetadata.Subject
		}
	}
	return title
}
