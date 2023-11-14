package format

import (
	"lath/xman/internal/db"
	"log"

	"github.com/google/uuid"
)

var BorgEndpoint = "http://localhost:3000/analyse"

func VerifyFileFormats(messageID uuid.UUID) {
	primaryDocuments, _ := db.GetAllPrimaryDocuments(messageID)
	log.Println(primaryDocuments)
}
