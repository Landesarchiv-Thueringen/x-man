package filesystem

import "lath/xman/internal/db"

type ArchivePackage struct {
	Title          string
	Lifetime       string
	Representation Representation
}

type Representation struct {
	Title            string
	PrimaryDocuments []db.PrimaryDocument
}
