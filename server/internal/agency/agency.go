package agency

import (
	"lath/xman/internal/db"
)

func InitAgencies() {
	agencies := []db.Agency{
		{
			Name:           "Thüringer Ministerium für Inneres und Kommunales",
			Abbreviation:   "TMIK",
			TransferDirURL: "file:///xman/transfer_dir",
			Code:           "TMIK",
		},
		{
			Name:           "Thüringer Staatskanzlei",
			Abbreviation:   "TSK",
			TransferDirURL: "webdav:///xman/transfer_dir",
			Code:           "TMIK",
		},
	}
	db.InitAgencies(agencies)
}
