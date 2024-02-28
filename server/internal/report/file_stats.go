package report

import (
	"lath/xman/internal/db"
	"os"
	"path"
)

type FileStats struct {
	TotalFiles      uint
	TotalBytes      uint64
	FilesByFileType map[string]uint
}

func getFileStats(process db.Process) (fileStats FileStats) {
	documents := db.GetAllPrimaryDocumentsWithFormatVerification(*process.Message0503ID)
	fileStats.FilesByFileType = make(map[string]uint)
	for _, document := range documents {
		mimeType := getMimeType(document)
		fileStats.FilesByFileType[mimeType] += 1
		fileStats.TotalFiles += 1
		fileSize := getFileSize(path.Join(process.Message0503.StoreDir, document.FileName))
		fileStats.TotalBytes += fileSize
	}
	return
}

func getFileSize(path string) uint64 {
	fi, err := os.Stat(path)
	if err != nil {
		panic(err)
	}
	return uint64(fi.Size())
}

func getMimeType(document db.PrimaryDocument) string {
	if document.FormatVerification == nil {
		return ""
	}
	for _, feature := range document.FormatVerification.Features {
		if feature.Key == "mimeType" {
			return feature.Values[0].Value
		}
	}
	return ""
}
