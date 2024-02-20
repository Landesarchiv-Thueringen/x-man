package filesystem

import (
	"fmt"
	"lath/xman/internal/db"
	"os"

	"github.com/google/uuid"
)

// ArchiveMessage creates on distinct folder for every archive package on a local filesystem
func ArchiveMessage(process db.Process, message db.Message, path string) error {
	for _, fileRecordObject := range message.FileRecordObjects {
		representation := Representation{
			Title:            fileRecordObject.GetTitle(),
			PrimaryDocuments: fileRecordObject.GetPrimaryDocuments(),
		}
		archivePackage := ArchivePackage{
			Title:          "originale Repräsentation einer Akte extrahiert aus einer E-Akten-Ablieferung",
			Lifetime:       fileRecordObject.GetCombinedLifetime(),
			Representation: representation,
		}
	}
	return nil
}

// StoreArchivePackage creates a folder on the file system for the archive package and copies all primary files in this folder.
func StoreArchivePackage(message db.Message, path string, archivePackage ArchivePackage) error {
	id := uuid.New().String()
	err := os.Mkdir(id, 0700)
	if err != nil {
		return fmt.Errorf()
	}
	return nil
}
