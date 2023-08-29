package messagestore

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"lath/xdomea/internal/xdomea"
	"log"
	"os"
	"path"
	"path/filepath"
)

var storeDir = "message_store"

func StoreMessage(messagePath string) {
	id := xdomea.GetMessageID(messagePath)
	messageName := filepath.Base(messagePath)
	// Create temporary directory. The name of the directory ist the message ID.
	tempDir, err := ioutil.TempDir("", id)
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
	extractMessage(copyPath, id)
}

func extractMessage(messagePath string, id string) {
	// Create the message store directory if necessarry.
	messageStorePath := path.Join(storeDir, id)
	err := os.MkdirAll(messageStorePath, 0700)
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
		fileStorePath := path.Join(messageStorePath, f.Name)
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

}
