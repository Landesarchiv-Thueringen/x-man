package xdomea

import (
	"context"
	"fmt"
	"lath/xman/internal/db"
	"lath/xman/internal/tasks"
	"log"
	"os"

	"github.com/google/uuid"
)

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
	// Cancel running tasks
	tasks.CancelAndDeleteTasksForProcess(processID, nil)
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
	}
	db.DeleteProcessingErrorsForProcess(processID)
	return true
}

func DeleteMessage(processID uuid.UUID, messageType db.MessageType, keepTransferFile bool) error {
	message, ok := db.FindMessage(context.Background(), processID, messageType)
	if !ok {
		return fmt.Errorf("%s message not found for process %s", messageType, processID)
	}
	storeDir := message.StoreDir
	transferFile := message.TransferDirPath
	if keepTransferFile {
		log.Printf("Deleting %s message for process %s (keeping transfer file)", messageType, processID)
	} else {
		log.Printf("Deleting %s message for process %s", messageType, processID)
	}
	// Cancel running tasks
	taskTypes := make(map[db.ProcessStepType]bool)
	switch messageType {
	case db.MessageType0503:
		taskTypes[db.ProcessStepFormatVerification] = true
		taskTypes[db.ProcessStepArchiving] = true
	}
	tasks.CancelAndDeleteTasksForProcess(processID, taskTypes)
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
	db.MustUpdateProcessStepCompletion(processID, processStepType, false, "")
	// Delete message storage
	if err := os.RemoveAll(storeDir); err != nil {
		panic(err)
	}
	process, processFound := db.FindProcess(context.Background(), message.MessageHead.ProcessID)
	if !processFound {
		panic("process not found " + message.MessageHead.ProcessID.String())
	}
	// Delete transfer file
	if keepTransferFile {
		db.DeleteTransferFile(process.Agency.ID, transferFile)
	} else {
		RemoveFileFromTransferDir(process.Agency, transferFile)
		cleanupEmptyProcess(message.MessageHead.ProcessID)
	}
	db.DeleteProcessingErrorsForMessage(processID, message.MessageType)
	return nil
}

// cleanupEmptyProcess deletes the given process if if does not have any
// messages.
func cleanupEmptyProcess(processID uuid.UUID) {
	if processID == uuid.Nil {
		panic("called cleanupEmptyProcess with empty string")
	}
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
	files := db.FindTransferDirFilesForProcess(p.ProcessID)
	filenames := make([]string, len(files))
	for i, f := range files {
		filenames[i] = f.Path
	}
	return filenames
}
