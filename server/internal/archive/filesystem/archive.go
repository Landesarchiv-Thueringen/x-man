package filesystem

import (
	"encoding/json"
	"io"
	"lath/xman/internal/db"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

const temporaryArchivePath = "/xman/archive"

// ArchiveMessage creates on distinct folder for every archive package on a local filesystem
func ArchiveMessage(process db.Process, message db.Message) error {
	for _, fileRecordObject := range message.FileRecordObjects {
		archivePackage := createAipFromFileRecordObject(process, fileRecordObject)
		err := StoreArchivePackage(process, message, archivePackage)
		if err != nil {
			return err
		}
	}
	for _, processRecordObject := range message.ProcessRecordObjects {
		archivePackage := createAipFromProcessRecordObject(process, processRecordObject)
		err := StoreArchivePackage(process, message, archivePackage)
		if err != nil {
			return err
		}
	}
	// combine documents which don't belong to a file or process in one archive package
	if len(message.DocumentRecordObjects) > 0 {
		archivePackage := createAipFromDocumentRecordObjects(process, message.DocumentRecordObjects)
		err := StoreArchivePackage(process, message, archivePackage)
		if err != nil {
			return err
		}
	}
	return nil
}

// createAipFromFileRecordObject creates the archive package metadata from a file record object.
func createAipFromFileRecordObject(process db.Process, fileRecordObject db.FileRecordObject) db.ArchivePackage {
	archivePackageData := db.ArchivePackage{
		ProcessID:          process.ID,
		IOTitle:            fileRecordObject.GetTitle(),
		IOLifetimeCombined: fileRecordObject.GetCombinedLifetime(),
		REPTitle:           "Original",
		PrimaryDocuments:   fileRecordObject.GetPrimaryDocuments(),
		FileRecordObjects:  []db.FileRecordObject{fileRecordObject},
	}
	return archivePackageData
}

// createAipFromProcessRecordObject creates the archive package metadata from a process record object.
func createAipFromProcessRecordObject(process db.Process, processRecordObject db.ProcessRecordObject) db.ArchivePackage {
	archivePackageData := db.ArchivePackage{
		ProcessID:            process.ID,
		IOTitle:              processRecordObject.GetTitle(),
		IOLifetimeCombined:   processRecordObject.GetCombinedLifetime(),
		REPTitle:             "Original",
		PrimaryDocuments:     processRecordObject.GetPrimaryDocuments(),
		ProcessRecordObjects: []db.ProcessRecordObject{processRecordObject},
	}
	return archivePackageData
}

// createAipFromDocumentRecordObjects creates the metadata for a shared archive package of multiple documents.
func createAipFromDocumentRecordObjects(
	process db.Process,
	documentRecordObjects []db.DocumentRecordObject,
) db.ArchivePackage {
	var primaryDocuments []db.PrimaryDocument
	for _, documentRecordObject := range documentRecordObjects {
		primaryDocuments = append(primaryDocuments, documentRecordObject.GetPrimaryDocuments()...)
	}
	ioTitle := "Nicht zugeordnete Dokumente Behörde: " + process.Agency.Name +
		" Prozess-ID: " + process.ID
	repTitle := "Original"
	archivePackageData := db.ArchivePackage{
		ProcessID:             process.ID,
		IOTitle:               ioTitle,
		IOLifetimeCombined:    "-",
		REPTitle:              repTitle,
		PrimaryDocuments:      primaryDocuments,
		DocumentRecordObjects: documentRecordObjects,
	}
	return archivePackageData
}

// StoreArchivePackage creates a folder on the file system for the archive package and copies all primary files in this folder.
func StoreArchivePackage(process db.Process, message db.Message, archivePackage db.ArchivePackage) error {
	id := uuid.New().String()
	archivePackagePath := filepath.Join(temporaryArchivePath, id)
	err := os.Mkdir(archivePackagePath, 0744)
	if err != nil {
		return err
	}
	// copy all primary documents in archive package
	for _, primaryDocument := range archivePackage.PrimaryDocuments {
		err := copyFileIntoArchivePackage(message.StoreDir, archivePackagePath, primaryDocument.FileName)
		if err != nil {
			return err
		}
	}
	messageFileName := filepath.Base(message.MessagePath)
	err = copyFileIntoArchivePackage(message.StoreDir, archivePackagePath, messageFileName)
	if err != nil {
		return err
	}
	err = writeObjectToTextfile(process.ProcessState, archivePackagePath, "protocol.json")
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

// writeObjectToTextfile
func writeObjectToTextfile(obj any, archivePackagePath string, filename string) error {
	bytes, err := json.MarshalIndent(obj, "", " ")
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(archivePackagePath, filename), bytes, 0644)
	return err
}
