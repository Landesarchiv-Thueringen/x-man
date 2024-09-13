package report

import (
	"context"
	"lath/xman/internal/db"
)

type ContentStats struct {
	Files     int
	Processes int
	Documents int
}

func getMessageContentStats(ctx context.Context, message db.Message) ContentStats {
	rootRecords := db.FindAllRootRecords(ctx, message.MessageHead.ProcessID, message.MessageType)
	return ContentStats{
		Files:     len(rootRecords.Files),
		Processes: len(rootRecords.Processes),
		Documents: len(rootRecords.Documents),
	}

}
