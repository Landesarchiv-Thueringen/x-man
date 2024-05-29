package xdomea

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"lath/xman/internal/db"
	"log"
	"os"
	"path"

	"github.com/google/uuid"
)

const rootStoreDir = "message_store"

// extractMessage parses the given message file into a database entry and saves
// it to the database. It returns the saved entry.
//
// Returns the directories in message store for the process and the message.
func extractMessageToMessageStore(
	agency db.Agency,
	transferDirMessagePath string,
	localMessagePath string,
	processID uuid.UUID,
	messageType db.MessageType,
) (processStoreDir string, messageStoreDir string) {
	processStoreDir = path.Join(rootStoreDir, processID.String())
	// Create the message store directory if necessary.
	messageStoreDir = path.Join(processStoreDir, string(messageType))
	err := os.MkdirAll(messageStoreDir, 0700)
	if err != nil {
		panic(err)
	}
	// Open the message archive (zip).
	archive, err := zip.OpenReader(localMessagePath)
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
	return
}

// DeleteProcess deletes the given process from the database and removes all
// associated message files from the file system.
//
// Returns true, when an entry was found and deleted.
func DeleteProcess(processID uuid.UUID) bool {
	if processID == uuid.Nil {
		panic("called DeleteProcess with empty string")
	}
	process, found := db.FindProcess(context.Background(), processID)
	if !found {
		return false
	}
	storeDir := process.StoreDir
	transferFiles := getAllTransferFilesOfProcess(process)
	log.Println("Deleting process", processID)
	// Delete database entries
	db.DeleteProcess(processID)
	db.DeleteMessagesForProcess(processID)
	db.DeleteRecordsForProcess(processID)
	db.DeletePrimaryDocumentsDataForProcess(processID)
	db.DeleteAppraisalsForProcess(processID)
	db.DeleteArchivePackagesForProcess(processID)
	// Delete message storage
	if err := os.RemoveAll(storeDir); err != nil {
		panic(err)
	}
	// Delete transfer files
	for _, path := range transferFiles {
		RemoveFileFromTransferDir(process.Agency, path)
		db.DeleteProcessedTransferDirFile(process.Agency.ID, path)
	}
	db.DeleteProcessingErrorsForProcess(processID)
	return true
}

func DeleteMessage(processID uuid.UUID, messageType db.MessageType, keepTransferFile bool) {
	message, messageFound := db.FindMessage(context.Background(), processID, messageType)
	if !messageFound {
		panic(fmt.Sprintf("%s message not found for process %s", messageType, processID))
	}
	storeDir := message.StoreDir
	transferFile := message.TransferDirPath
	if keepTransferFile {
		log.Printf("Deleting %s message for process %s (keeping transfer file)", messageType, processID)
	} else {
		log.Printf("Deleting %s message for process %s", messageType, processID)
	}
	// Delete database entries
	db.DeleteMessage(message)
	db.DeleteRecordsForMessage(message.MessageHead.ProcessID, message.MessageType)
	if message.MessageType == db.MessageType0503 {
		db.DeletePrimaryDocumentsDataForProcess(message.MessageHead.ProcessID)
	}
	// Reset process step
	var processStepType db.ProcessStepType
	switch message.MessageType {
	case db.MessageType0501:
		processStepType = db.ProcessStepReceive0501
	case db.MessageType0503:
		processStepType = db.ProcessStepReceive0503
	case db.MessageType0505:
		processStepType = db.ProcessStepReceive0505
	}
	db.UpdateProcessStepCompletion(processID, processStepType, false, "")
	// Delete message storage
	if err := os.RemoveAll(storeDir); err != nil {
		panic(err)
	}
	process, processFound := db.FindProcess(context.Background(), message.MessageHead.ProcessID)
	if !processFound {
		panic("process not found " + message.MessageHead.ProcessID.String())
	}
	// Delete transfer file
	if !keepTransferFile {
		RemoveFileFromTransferDir(process.Agency, transferFile)
		cleanupEmptyProcess(message.MessageHead.ProcessID)
	}
	if processFound {
		// If the process cannot be found, the processed-transfer-dir entry was
		// already deleted with the process.
		db.DeleteProcessedTransferDirFile(process.Agency.ID, message.TransferDirPath)
	}
	db.DeleteProcessingErrorsForMessage(processID, message.MessageType)
}

// cleanupEmptyProcess deletes the given process if if does not have any
// messages.
func cleanupEmptyProcess(processID uuid.UUID) {
	if processID == uuid.Nil {
		panic("called cleanupEmptyProcess with empty string")
	}
	log.Println("cleanupEmptyProcess", processID)
	messages := db.FindMessagesForProcess(context.Background(), processID)
	if len(messages) == 0 {
		if found := DeleteProcess(processID); !found {
			panic(fmt.Sprintf("process not found: %v", processID))
		}
	}
}

// getAllTransferFilesOfProcess returns the transfer paths of all messages that
// belong to the given process.
func getAllTransferFilesOfProcess(p db.SubmissionProcess) []string {
	transferDirPaths := make([]string, 0)
	if p.Message0502Path != "" {
		transferDirPaths = append(transferDirPaths, p.Message0502Path)
	}
	if p.Message0504Path != "" {
		transferDirPaths = append(transferDirPaths, p.Message0504Path)
	}
	if p.Message0506Path != "" {
		transferDirPaths = append(transferDirPaths, p.Message0506Path)
	}
	messages := db.FindMessagesForProcess(context.Background(), p.ProcessID)
	for _, m := range messages {
		transferDirPaths = append(transferDirPaths, m.TransferDirPath)
	}
	return transferDirPaths
}
