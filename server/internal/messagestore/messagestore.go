package messagestore

import (
	"archive/zip"
	"fmt"
	"io"
	"lath/xman/internal/db"
	"lath/xman/internal/xdomea"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/google/uuid"
)

var storeDir = "message_store"

func StoreMessage(agency db.Agency, messagePath string) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error: StoreMessage panicked:", r)
			debug.PrintStack()
		}
	}()
	id := xdomea.GetMessageID(messagePath)
	transferDir := filepath.Dir(messagePath)
	messageName := filepath.Base(messagePath)
	// Create temporary directory. The name of the directory ist the message ID.
	tempDir, err := os.MkdirTemp("", id)
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)
	// Open the original message in the transfer directory.
	message, err := os.Open(messagePath)
	if err != nil {
		panic(err)
	}
	defer message.Close()
	// Create a file in the temporary directory.
	copyPath := path.Join(tempDir, messageName)
	copy, err := os.Create(copyPath)
	if err != nil {
		panic(err)
	}
	defer copy.Close()
	// Copy the message to the new file.
	_, err = io.Copy(copy, message)
	if err != nil {
		panic(err)
	}
	extractMessage(agency, transferDir, messagePath, copyPath, id)
}

func extractMessage(
	agency db.Agency,
	transferDir string,
	transferDirMessagePath string,
	messagePath string,
	id string,
) {
	messageType, err := xdomea.GetMessageTypeImpliedByPath(messagePath)
	// The error should never happen because the message filter should prevent the pross
	if err != nil {
		panic(fmt.Sprintf("failed to extract message: %v", err))
	}
	processStoreDir := path.Join(storeDir, id)
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
	process, message, err :=
		xdomea.AddMessage(
			agency,
			id,
			messageType,
			processStoreDir,
			messageStoreDir,
			transferDir,
			transferDirMessagePath,
		)
	// if no error occurred while processing the message
	if err == nil {
		// store the confirmation message that the 0501 message was received
		if messageType.Code == "0501" {
			messagePath := Store0504Message(message)
			process.Message0504Path = &messagePath
			db.UpdateProcess(process)
		}
	}
}

func Store0502Message(message db.Message) string {
	messageXml := xdomea.Generate0502Message(message)
	return storeMessage(
		message.MessageHead.ProcessID,
		messageXml,
		xdomea.Message0502MessageSuffix,
		message.TransferDir,
	)
}

func Store0504Message(message db.Message) string {
	messageXml := xdomea.Generate0504Message(message)
	return storeMessage(
		message.MessageHead.ProcessID,
		messageXml,
		xdomea.Message0504MessageSuffix,
		message.TransferDir,
	)
}

func storeMessage(
	messageID string,
	messageXml string,
	messageSuffix string,
	transferDir string,
) string {
	// Create temporary directory. The name of the directory ist the message ID.
	tempDir, err := os.MkdirTemp("", messageID)
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)
	xmlName := messageID + messageSuffix + ".xml"
	messageName := messageID + messageSuffix + ".zip"
	messagePath := path.Join(tempDir, messageName)
	messageArchive, err := os.Create(messagePath)
	if err != nil {
		panic(err)
	}
	defer messageArchive.Close()
	zipWriter := zip.NewWriter(messageArchive)
	defer zipWriter.Close()
	zipEntry, err := zipWriter.Create(xmlName)
	if err != nil {
		panic(err)
	}
	xmlStringReader := strings.NewReader(messageXml)
	_, err = io.Copy(zipEntry, xmlStringReader)
	if err != nil {
		panic(err)
	}
	// important close zip writer and message archive so it can be written on disk
	zipWriter.Close()
	messageArchive.Close()
	messageArchive, err = os.Open(messagePath)
	if err != nil {
		panic(err)
	}
	messageTransferDirPath := path.Join(transferDir, messageName)
	messageInTransferDir, err := os.Create(messageTransferDirPath)
	if err != nil {
		panic(err)
	}
	defer messageInTransferDir.Close()
	// Copy the message to the transfer directory.
	_, err = io.Copy(messageInTransferDir, messageArchive)
	if err != nil {
		panic(err)
	}
	return messageTransferDirPath
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
	for _, f := range transferFiles {
		if err := os.Remove(f); err != nil {
			panic(err)
		}
	}
	return true
}

func DeleteMessage(id uuid.UUID, keepTransferFile bool) {
	message, found := db.GetCompleteMessageByID(id)
	if !found {
		panic("message not found " + id.String())
	}
	storeDir := message.StoreDir
	transferFile := message.TransferDirMessagePath
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
