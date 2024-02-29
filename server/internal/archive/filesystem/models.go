package filesystem

import "lath/xman/internal/db"

type ArchivePackage struct {
	Title          string
	Representation Representation
}

type Representation struct {
	Title            string
	PrimaryDocuments []db.PrimaryDocument
}
