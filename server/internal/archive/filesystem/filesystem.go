package filesystem

import (
	"encoding/json"
	"io"
	"lath/xman/internal/archive/shared"
	"lath/xman/internal/db"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

const temporaryArchivePath = "/xman/archive"

// StoreArchivePackage creates a folder on the file system for the archive
// package and copies all relevant files in this folder.
func StoreArchivePackage(
	process db.SubmissionProcess,
	message db.Message,
	archivePackage db.ArchivePackage,
) {
	id := uuid.New().String()
	archivePackagePath := filepath.Join(temporaryArchivePath, id)
	err := os.Mkdir(archivePackagePath, 0744)
	if err != nil {
		panic(err)
	}
	// Add primary documents.
	for _, primaryDocument := range archivePackage.PrimaryDocuments {
		err := copyFileIntoArchivePackage(message.StoreDir, archivePackagePath, primaryDocument.Filename)
		if err != nil {
			panic(err)
		}
	}
	// Add relevant part of xdomea message.
	prunedMessage := shared.PruneMessage(message, archivePackage)
	messageFileName := filepath.Base(message.MessagePath)
	err = writeFile(archivePackagePath, messageFileName, prunedMessage)
	if err != nil {
		panic(err)
	}
	// Add protocol.
	err = writeFile(archivePackagePath, shared.ProtocolFilename, shared.GenerateProtocol(process))
	if err != nil {
		panic(err)
	}
	// Add format verification results.
	if f, ok := shared.GenerateVerificationResults(process.ProcessID, archivePackage); ok {
		writeFile(archivePackagePath, "verification_results.json", f)
	}
	// Add internal archive-package object.
	writeObjectToTextfile(archivePackage, archivePackagePath, "aip.json")
	// Calculate and add checksums.
	checksums := shared.Sha512Sum(archivePackagePath, "", true)
	err = writeFile(archivePackagePath, "sha512sum.txt", checksums)
	if err != nil {
		panic(err)
	}
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
func writeObjectToTextfile(obj any, archivePackagePath string, filename string) {
	bytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(filepath.Join(archivePackagePath, filename), bytes, 0644)
	if err != nil {
		panic(err)
	}
}

// writeFile writes a textfile in the archive package.
func writeFile(aipPath string, filename string, content []byte) error {
	return os.WriteFile(filepath.Join(aipPath, filename), content, 0644)
}
