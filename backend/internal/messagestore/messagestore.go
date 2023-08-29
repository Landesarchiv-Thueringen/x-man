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

var storagePath = "message_storage"

func StoreMessage(messagePath string) {
	id := xdomea.GetMessageID(messagePath)
	messageName := filepath.Base(messagePath)
	// Create temporary directory. The name of the directory ist the message ID.
	tempDir, err := ioutil.TempDir("", id)
	if (err) != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	// Open the original message in the transfer directory.
	message, err := os.Open(messagePath)
	if (err) != nil {
		log.Fatal(err)
	}
	defer message.Close()
	// Create a file in the temporary directory.
	copyPath := path.Join(tempDir, messageName)
	copy, err := os.Create(copyPath)
	if (err) != nil {
		log.Fatal(err)
	}
	defer copy.Close()
	// Copy the message to the new file.
	_, err = io.Copy(copy, message)
	if (err) != nil {
		log.Fatal(err)
	}
}

func extractMessage(path string, id string) {
	archive, err := zip.OpenReader(path)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()
}
