package report

import (
	"context"
	"lath/xman/internal/db"
)

type ContentStats struct {
	Files     uint
	Processes uint
	Documents uint
}

func getMessageContentStats(ctx context.Context, message db.Message) ContentStats {
	rootRecords := db.FindRootRecords(ctx, message.MessageHead.ProcessID, message.MessageType)
	return ContentStats{
		Files:     uint(len(rootRecords.Files)),
		Processes: uint(len(rootRecords.Processes)),
		Documents: uint(len(rootRecords.Documents)),
	}

}
