package db

import (
	"time"
)

// archive packages will be deleted when the process is deleted
type ArchivePackage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	ProcessID string    `json:"-"`
	Process   *Process  `gorm:"foreignKey:ProcessID;" json:"-"`
	// DIMAG collection (de: Bestand)
	CollectionID *uint       `json:"collectionID"`
	Collection   *Collection `gorm:"foreignKey:CollectionID;" json:"-"`
	// title of the information object in DIMAG
	IOTitle string `json:"ioTitle"`
	// combined lifetime begin and end of the information object in DIMAG
	IOLifetimeCombined string `json:"ioLifeTime"`
	// title of the representation in DIMAG
	REPTitle string `json:"repTitle"`
	// contained root record objects
	FileRecordObjects     []FileRecordObject     `gorm:"many2many:aip_file_record_objects;constraint:OnDelete:CASCADE"`
	ProcessRecordObjects  []ProcessRecordObject  `gorm:"many2many:aip_process_record_objects;constraint:OnDelete:CASCADE"`
	DocumentRecordObjects []DocumentRecordObject `gorm:"many2many:aip_document_record_objects;constraint:OnDelete:CASCADE"`
	// all primary documents that are contained in the archive package
	PrimaryDocuments []PrimaryDocument `gorm:"many2many:aip_primary_documents;constraint:OnDelete:CASCADE"`
	// PackageID is the ID assigned by DIMAG when importing the package.
	PackageID string
}

// GetRootIDs returns an array containing all IDs of root records in the archive package.
func (aip *ArchivePackage) GetRootIDs() []string {
	var rootIDs []string
	for _, fileRecord := range aip.FileRecordObjects {
		rootIDs = append(rootIDs, fileRecord.XdomeaID.String())
	}
	for _, processRecord := range aip.ProcessRecordObjects {
		rootIDs = append(rootIDs, processRecord.XdomeaID.String())
	}
	for _, documentRecord := range aip.DocumentRecordObjects {
		rootIDs = append(rootIDs, documentRecord.XdomeaID.String())
	}
	return rootIDs
}
