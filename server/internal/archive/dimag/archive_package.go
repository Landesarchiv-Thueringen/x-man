package dimag

import (
	"lath/xman/internal/archive/shared"
	"lath/xman/internal/db"
	"path/filepath"
)

// This file handles creation of an archive package (AIP) as a BagIt structure.

func createArchiveBagit(
	process db.SubmissionProcess,
	message db.Message,
	archivePackage db.ArchivePackage,
) bagitHandle {
	bagit := makeBagit()
	bagit.CreateFile(
		filepath.Join("data", filepath.Base(message.MessagePath)),
		shared.PruneMessage(message, archivePackage),
	)
	for _, d := range archivePackage.PrimaryDocuments {
		bagit.CopyFile(
			filepath.Join("data", d.Filename),
			filepath.Join(message.StoreDir, d.Filename),
		)
	}
	bagit.CreateFile(
		filepath.Join("dimag", "control.xml"),
		generateControlFile(message, archivePackage, filepath.Join(getUploadDir(bagit), "data")),
	)
	// TODO: generate protocol in XML format.
	//
	// bagit.CreateFile(
	// 	filepath.Join("dimag", "protocol.xml"),
	// 	shared.GenerateProtocol(process),
	// )
	bagit.Finalize()
	return bagit
}
