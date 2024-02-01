package agency

import (
	"lath/xman/internal/db"
	"log"
)

func InitAgencies() {
	agencies := []db.Agency{
		{
			Name:         "Thüringer Ministerium für Inneres und Kommunales",
			Abbreviation: "TMIK",
			TransferDir:  "/xman/transfer_dir",
			UserIDs:      [][]byte{},
		},
	}
	db.InitAgencies(agencies)
}

func MonitorTransferDirs() {
	agencies, err := db.GetAgencies()
	if err != nil {
		log.Fatal(err)
	}
	go watchTransferDirectories(agencies)
}
