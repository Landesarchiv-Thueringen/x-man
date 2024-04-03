package xdomea

import (
	"archive/zip"
	"fmt"
	"io"
	"lath/xman/internal/db"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/google/uuid"
)

const storeDir = "message_store"

// StoreMessage extracts the message from the transfer directory into the message store.
func StoreMessage(agency db.Agency, messagePath string) (db.Message, error) {
	processID := GetMessageID(messagePath)
	messageName := filepath.Base(messagePath)
	// Create temporary directory. The name of the directory contains the message ID.
	tempDir, err := os.MkdirTemp("", processID)
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)
	// Open the original messageFile in the transfer directory.
	messageFile, err := os.Open(messagePath)
	if err != nil {
		panic(err)
	}
	defer messageFile.Close()
	// Create a file in the temporary directory.
	copyPath := path.Join(tempDir, messageName)
	copy, err := os.Create(copyPath)
	if err != nil {
		panic(err)
	}
	defer copy.Close()
	// Copy the message to the new file.
	_, err = io.Copy(copy, messageFile)
	if err != nil {
		panic(err)
	}
	transferDirMessagePath := filepath.Base(messagePath)
	return extractMessage(agency, transferDirMessagePath, copyPath, processID)
}

// extractMessage parses the given message file into a database entry and saves
// it to the database. It returns the saved entry.
func extractMessage(
	agency db.Agency,
	transferDirMessagePath string,
	messagePath string,
	processID string,
) (db.Message, error) {
	messageType, err := GetMessageTypeImpliedByPath(messagePath)
	// The error should never happen because the message filter should prevent the processing of unknown message types.
	if err != nil {
		panic(fmt.Sprintf("failed to extract message: %v", err))
	}
	processStoreDir := path.Join(storeDir, processID)
	// Create the message store directory if necessary.
	messageStoreDir := path.Join(processStoreDir, messageType.Code)
	err = os.MkdirAll(messageStoreDir, 0700)
	if err != nil {
		panic(err)
	}
	// Open the message archive (zip).
	archive, err := zip.OpenReader(messagePath)
	if err != nil {
		panic(err)
	}
	defer archive.Close()
	for _, f := range archive.File {
		fileInArchive, err := f.Open()
		if err != nil {
			panic(err)
		}
		defer fileInArchive.Close()
		fileStorePath := path.Join(messageStoreDir, f.Name)
		fileInStore, err := os.Create(fileStorePath)
		if err != nil {
			panic(err)
		}
		defer fileInStore.Close()
		_, err = io.Copy(fileInStore, fileInArchive)
		if err != nil {
			panic(err)
		}
	}
	_, message, err :=
		AddMessage(
			agency,
			processID,
			messageType,
			processStoreDir,
			messageStoreDir,
			filepath.Base(transferDirMessagePath),
		)
	return message, err
}

// DeleteProcess deletes the given process from the database and removes all
// associated message files from the file system.
//
// Returns true, when an entry was found and deleted.
func DeleteProcess(processID string) bool {
	if processID == "" {
		panic("called DeleteProcess with empty string")
	}
	process, found := db.GetProcess(processID)
	if !found {
		return false
	}
	storeDir := process.StoreDir
	transferFiles := db.GetAllTransferFilesOfProcess(process)
	log.Println("Deleting process", processID)
	// Delete database entries
	db.DeleteProcess(process.ID)
	// Delete message storage
	if err := os.RemoveAll(storeDir); err != nil {
		panic(err)
	}
	// Delete transfer files
	for _, path := range transferFiles {
		RemoveFileFromTransferDir(process.Agency, path)
	}
	return true
}

func DeleteMessage(id uuid.UUID, keepTransferFile bool) {
	message, found := db.GetCompleteMessageByID(id)
	if !found {
		panic("message not found " + id.String())
	}
	storeDir := message.StoreDir
	transferFile := message.TransferDirPath
	if keepTransferFile {
		log.Println("Deleting message", message.ID, "(keeping transfer file)")
	} else {
		log.Println("Deleting message", message.ID)
	}
	db.DeleteMessage(message)
	// Delete message storage
	if err := os.RemoveAll(storeDir); err != nil {
		panic(err)
	}
	// Delete transfer file
	if !keepTransferFile {
		if err := os.Remove(transferFile); err != nil {
			panic(err)
		}
		cleanupEmptyProcess(message.MessageHead.ProcessID)
	}
}

// cleanupEmptyProcess deletes the given process if if does not have any
// messages.
func cleanupEmptyProcess(processID string) {
	if processID == "" {
		panic("called cleanupEmptyProcess with empty string")
	}
	process, found := db.GetProcess(processID)
	if !found {
		panic(fmt.Sprintf("process not found: %v", processID))
	}
	log.Println("cleanupEmptyProcess", processID)
	if process.Message0501ID == nil && process.Message0503ID == nil && process.Message0505ID == nil {
		if found = DeleteProcess(processID); !found {
			panic(fmt.Sprintf("process not found: %v", processID))
		}
	}
}
