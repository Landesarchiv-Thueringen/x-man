package filesystem

import (
	"context"
	"encoding/json"
	"io"
	"lath/xman/internal/archive"
	"lath/xman/internal/db"
	"lath/xman/internal/xdomea"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

const temporaryArchivePath = "/xman/archive"

// ArchiveMessage creates on distinct folder for every archive package on a local filesystem
func ArchiveMessage(process db.SubmissionProcess, message db.Message) error {
	rootRecords := db.FindRootRecords(context.Background(), process.ProcessID, message.MessageType)
	for _, f := range rootRecords.Files {
		aip := createAipFromFileRecordObject(process, f)
		err := StoreArchivePackage(process, message, aip)
		if err != nil {
			return err
		}
		db.InsertArchivePackage(aip)
	}
	for _, p := range rootRecords.Processes {
		archivePackage := createAipFromProcessRecordObject(process, p)
		err := StoreArchivePackage(process, message, archivePackage)
		if err != nil {
			return err
		}
		db.InsertArchivePackage(archivePackage)
	}
	// combine documents which don't belong to a file or process in one archive package
	if len(rootRecords.Documents) > 0 {
		archivePackage := createAipFromDocumentRecordObjects(process, rootRecords.Documents)
		err := StoreArchivePackage(process, message, archivePackage)
		if err != nil {
			return err
		}
		db.InsertArchivePackage(archivePackage)
	}
	return nil
}

// createAipFromFileRecordObject creates the archive package metadata from a file record object.
func createAipFromFileRecordObject(process db.SubmissionProcess, f db.FileRecord) db.ArchivePackage {
	archivePackageData := db.ArchivePackage{
		ProcessID:        process.ProcessID,
		IOTitle:          archive.GetFileRecordTitle(f),
		IOLifetime:       archive.GetCombinedLifetime(f.Lifetime),
		REPTitle:         "Original",
		RootRecordIDs:    []uuid.UUID{f.RecordID},
		PrimaryDocuments: xdomea.GetPrimaryDocumentsForFile(&f),
	}
	return archivePackageData
}

// createAipFromProcessRecordObject creates the archive package metadata from a process record object.
func createAipFromProcessRecordObject(process db.SubmissionProcess, p db.ProcessRecord) db.ArchivePackage {
	archivePackageData := db.ArchivePackage{
		ProcessID:        process.ProcessID,
		IOTitle:          archive.GetProcessRecordTitle(p),
		IOLifetime:       archive.GetCombinedLifetime(p.Lifetime),
		REPTitle:         "Original",
		RootRecordIDs:    []uuid.UUID{p.RecordID},
		PrimaryDocuments: xdomea.GetPrimaryDocumentsForProcess(&p),
	}
	return archivePackageData
}

// createAipFromDocumentRecordObjects creates the metadata for a shared archive package of multiple documents.
func createAipFromDocumentRecordObjects(
	process db.SubmissionProcess,
	documentRecords []db.DocumentRecord,
) db.ArchivePackage {
	var primaryDocuments []db.PrimaryDocument
	for _, d := range documentRecords {
		primaryDocuments = append(primaryDocuments, xdomea.GetPrimaryDocumentsForDocument(&d)...)
	}
	ioTitle := "Nicht zugeordnete Dokumente Behörde: " + process.Agency.Name +
		" Prozess-ID: " + process.ProcessID.String()
	repTitle := "Original"
	var rootRecordIDs []uuid.UUID
	for _, r := range documentRecords {
		rootRecordIDs = append(rootRecordIDs, r.RecordID)
	}
	aip := db.ArchivePackage{
		ProcessID:        process.ProcessID,
		IOTitle:          ioTitle,
		IOLifetime:       "-",
		REPTitle:         repTitle,
		RootRecordIDs:    rootRecordIDs,
		PrimaryDocuments: primaryDocuments,
	}
	return aip
}

// StoreArchivePackage creates a folder on the file system for the archive package and copies all primary files in this folder.
func StoreArchivePackage(
	process db.SubmissionProcess,
	message db.Message,
	archivePackage db.ArchivePackage,
) error {
	id := uuid.New().String()
	archivePackagePath := filepath.Join(temporaryArchivePath, id)
	err := os.Mkdir(archivePackagePath, 0744)
	if err != nil {
		return err
	}
	// copy all primary documents in archive package
	for _, primaryDocument := range archivePackage.PrimaryDocuments {
		err := copyFileIntoArchivePackage(message.StoreDir, archivePackagePath, primaryDocument.Filename)
		if err != nil {
			return err
		}
	}
	prunedMessage, err := xdomea.PruneMessage(message, archivePackage)
	if err != nil {
		return err
	}
	messageFileName := filepath.Base(message.MessagePath)
	err = writeTextFile(archivePackagePath, messageFileName, prunedMessage)
	if err != nil {
		return err
	}
	err = writeTextFile(archivePackagePath, archive.ProtocolFilename, archive.GenerateProtocol(process))
	if err != nil {
		return err
	}
	return writeObjectToTextfile(archivePackage, archivePackagePath, "aip.json")
}

// CopyFileIntoArchivePackage copies a file from the message store into an archive package.
func copyFileIntoArchivePackage(storePath string, archivePackagePath string, fileName string) error {
	srcPath := filepath.Join(storePath, fileName)
	dstPath := filepath.Join(archivePackagePath, fileName)
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}
	return nil
}

// writeObjectToTextfile writes an object to a textfile in the archive package.
func writeObjectToTextfile(obj any, archivePackagePath string, filename string) error {
	bytes, err := json.MarshalIndent(obj, "", " ")
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(archivePackagePath, filename), bytes, 0644)
	return err
}

// writeTextFile writes a textfile in the archive package.
func writeTextFile(aipPath string, filename string, content string) error {
	return os.WriteFile(filepath.Join(aipPath, filename), []byte(content), 0644)
}
