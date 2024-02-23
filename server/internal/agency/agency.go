package agency

import (
	"lath/xman/internal/db"
)

func InitAgencies() {
	agencies := []db.Agency{
		{
			Name:         "Thüringer Ministerium für Inneres und Kommunales",
			Abbreviation: "TMIK",
			TransferDir:  "/xman/transfer_dir",
			Code:         "TMIK",
			UserIDs:      [][]byte{},
		},
	}
	db.InitAgencies(agencies)
}
