package messagestore

import (
	"archive/zip"
	"io"
	"lath/xman/internal/db"
	"lath/xman/internal/xdomea"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var storeDir = "message_store"

func StoreMessage(messagePath string) {
	id := xdomea.GetMessageID(messagePath)
	messageName := filepath.Base(messagePath)
	transferDir := filepath.Dir(messagePath)
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
	extractMessage(transferDir, copyPath, id)
}

func extractMessage(transferDir string, messagePath string, id string) {
	messageType, err := xdomea.GetMessageTypeImpliedByPath(messagePath)
	// The error should never happen because the message filter should prevent the pross
	if err != nil {
		log.Fatal("failed extracting message: unknown message type")
	}
	processStoreDir := path.Join(storeDir, id)
	// Create the message store directory if necessarry.
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
	process, message :=
		xdomea.AddMessage(id, messageType, processStoreDir, messageStoreDir, transferDir)
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
