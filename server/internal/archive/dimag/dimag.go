package dimag

import (
	"context"
	"fmt"
	"lath/xman/internal/db"
)

// ImportArchivePackage archives a file record object in DIMAG.
func ImportArchivePackage(
	ctx context.Context,
	process db.SubmissionProcess,
	message db.Message,
	aip *db.ArchivePackage,
	c Connection,
) error {
	bagit := createArchiveBagit(process, message, *aip)
	uploadDir, err := uploadBagit(ctx, c, bagit)
	if err != nil {
		return err
	}
	importBagResponse, err := importBag(ctx, uploadDir)
	if err != nil {
		return err
	}
	// We remove the BagIt when it was processed without errors. Otherwise, we
	// leave it for debugging purposes.
	bagit.Remove()
	packageID, err := packageID(importBagResponse)
	if err != nil {
		fmt.Printf("%#v\n", importBagResponse)
		return err
	}
	aip.PackageID = packageID
	ok := db.ReplaceArchivePackage(aip)
	if !ok {
		return fmt.Errorf("failed to set PackageID for archive package %v", aip.ID.Hex())
	}
	return nil
}
