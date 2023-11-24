package dimag

import (
	"lath/xman/internal/db"
	"log"
	"path/filepath"

	"github.com/google/uuid"
)

func ImportMessage(messageID uuid.UUID) {
	message, err := db.GetMessageByID(messageID)
	if err != nil {
		log.Fatal(err)
	}
	importDirs := TransferToArchive(message)
	for _, importDir := range importDirs {
		log.Println(importDir)
		requestMetadata := SoapImportDoc{
			UserName:        "grochow",
			Password:        "blablabla",
			ControlFilePath: filepath.Join(importDir, ControlFileName),
		}
		log.Println(requestMetadata)
	}
}
