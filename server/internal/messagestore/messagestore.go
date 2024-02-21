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
	"strings"

	"github.com/google/uuid"
)

var storeDir = "message_store"

func StoreMessage(agency db.Agency, messagePath string) {
	id := xdomea.GetMessageID(messagePath)
	transferDir := filepath.Dir(messagePath)
	messageName := filepath.Base(messagePath)
	// Create temporary directory. The name of the directory ist the message ID.
	tempDir, err := os.MkdirTemp("", id)
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	// Open the original message in the transfer directory.
	message, err := os.Open(messagePath)
	if err != nil {
		log.Fatal(err)
	}
	defer message.Close()
	// Create a file in the temporary directory.
	copyPath := path.Join(tempDir, messageName)
	copy, err := os.Create(copyPath)
	if err != nil {
		log.Fatal(err)
	}
	defer copy.Close()
	// Copy the message to the new file.
	_, err = io.Copy(copy, message)
	if err != nil {
		log.Fatal(err)
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
		log.Fatal("failed extracting message: unknown message type")
	}
	processStoreDir := path.Join(storeDir, id)
	// Create the message store directory if necessary.
	messageStoreDir := path.Join(processStoreDir, messageType.Code)
	err = os.MkdirAll(messageStoreDir, 0700)
	if err != nil {
		log.Fatal(err)
	}
	// Open the message archive (zip).
	archive, err := zip.OpenReader(messagePath)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()
	for _, f := range archive.File {
		fileInArchive, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		defer fileInArchive.Close()
		fileStorePath := path.Join(messageStoreDir, f.Name)
		fileInStore, err := os.Create(fileStorePath)
		if err != nil {
			log.Fatal(err)
		}
		defer fileInStore.Close()
		_, err = io.Copy(fileInStore, fileInArchive)
		if err != nil {
			log.Fatal(err)
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
			err = db.UpdateProcess(process)
			if err != nil {
				log.Fatal(err)
			}
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
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	xmlName := messageID + messageSuffix + ".xml"
	messageName := messageID + messageSuffix + ".zip"
	messagePath := path.Join(tempDir, messageName)
	messageArchive, err := os.Create(messagePath)
	if err != nil {
		log.Fatal(err)
	}
	defer messageArchive.Close()
	zipWriter := zip.NewWriter(messageArchive)
	defer zipWriter.Close()
	zipEntry, err := zipWriter.Create(xmlName)
	if err != nil {
		log.Fatal(err)
	}
	xmlStringReader := strings.NewReader(messageXml)
	_, err = io.Copy(zipEntry, xmlStringReader)
	if err != nil {
		log.Fatal(err)
	}
	// important close zip writer and message archive so it can be written on disk
	zipWriter.Close()
	messageArchive.Close()
	messageArchive, err = os.Open(messagePath)
	if err != nil {
		log.Fatal(err)
	}
	messageTransferDirPath := path.Join(transferDir, messageName)
	messageInTransferDir, err := os.Create(messageTransferDirPath)
	if err != nil {
		log.Fatal(err)
	}
	defer messageInTransferDir.Close()
	// Copy the message to the transfer directory.
	_, err = io.Copy(messageInTransferDir, messageArchive)
	if err != nil {
		log.Fatal(err)
	}
	return messageTransferDirPath
}

// DeleteProcess deletes the given process from the database and removes all
// associated message files from the file system.
//
// Returns true, when an entry was found and deleted.
func DeleteProcess(processID string) (bool, error) {
	process, err := db.GetProcessByXdomeaID(processID)
	if err != nil {
		return false, err
	}
	storeDir := process.StoreDir
	transferFiles, err := db.GetAllTransferFilesOfProcess(process)
	if err != nil {
		return false, err
	}
	// Delete database entries
	deleted, err := db.DeleteProcess(process.ID)
	if !deleted || err != nil {
		return deleted, err
	}
	// Delete message storage
	if err = os.RemoveAll(storeDir); err != nil {
		return false, err
	}
	// Delete transfer files
	for _, f := range transferFiles {
		if err = os.Remove(f); err != nil {
			return false, err
		}
	}
	return true, nil
}

func DeleteMessage(id uuid.UUID, keepTransferFile bool) (bool, error) {
	message, err := db.GetCompleteMessageByID(id)
	if err != nil {
		return false, err
	}
	storeDir := message.StoreDir
	transferFile := message.TransferDirMessagePath
	deleted, err := db.DeleteMessage(message)
	if !deleted || err != nil {
		return deleted, err
	}
	// Delete message storage
	if err = os.RemoveAll(storeDir); err != nil {
		return false, err
	}
	// Delete transfer file
	if !keepTransferFile {
		if err = os.Remove(transferFile); err != nil {
			return false, err
		}
		if err = cleanupEmptyProcess(message.MessageHead.ProcessID); err != nil {
			return true, err
		}
	}
	return true, nil
}

// cleanupEmptyProcess deletes the given process if if does not have any
// messages.
func cleanupEmptyProcess(processID string) error {
	process, err := db.GetProcessByXdomeaID(processID)
	if err != nil {
		return err
	}
	fmt.Println("cleanupEmptyProcess", processID)
	if process.Message0501ID == nil && process.Message0503ID == nil && process.Message0505ID == nil {
		_, err = DeleteProcess(processID)
	}
	return err
}
