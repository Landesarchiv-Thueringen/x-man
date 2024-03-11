package agency

import (
	"lath/xman/internal/db"
)

func InitAgencies() {
	agencies := []db.Agency{
		{
			Name:         "Thüringer Ministerium für Inneres und Kommunales",
			Abbreviation: "TMIK",
			TransferDir:  "/xman/transfer_dir/tmik",
			Code:         "TMIK",
		},
		{
			Name:         "Thüringer Staatskanzlei",
			Abbreviation: "TSK",
			TransferDir:  "/xman/transfer_dir/tsk",
			Code:         "TMIK",
		},
	}
	db.InitAgencies(agencies)
}
