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
			TransferDir:  "/xman/transfer/tmik",
		},
	}
	db.InitAgencies(agencies)
}

func MonitorTransferDirs() {
	agencies, err := db.GetAgencies()
	if err != nil {
		log.Println(err)
		log.Println("couldn't initialize the monitoring of the transfer directories")
		return
	}
	go watchTransferDirectories(agencies)
}
