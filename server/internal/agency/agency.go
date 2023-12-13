package agency

import (
	"errors"
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
	go watchTransferDirs(agencies)
}

func GetAgencyFromMessagePath(messagePath string) (db.Agency, error) {
	var agency db.Agency
	agencies, err := db.GetAgencies()
	if err != nil {
		log.Println(err)
		return agency, err
	}
	for _, agency := range agencies {
		if agency.IsFromTransferDir(messagePath) {
			return agency, nil
		}
	}
	return agency, errors.New("no agency existing with transfer dir for " + messagePath)
}
