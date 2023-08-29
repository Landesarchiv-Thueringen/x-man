package temporarystorage

import (
	"io"
	"io/ioutil"
	"lath/xdomea/internal/xdomea"
	"log"
	"os"
	"path"
	filepath "path/filepath"
)

func StoreMessage(messagePath string) {
	message, err := os.Open(messagePath)
	if (err) != nil {
		log.Fatal(err)
	}
	defer message.Close()
	id := xdomea.GetMessageID(messagePath)
	messageName := filepath.Base(messagePath)
	tempDir, err := ioutil.TempDir("", id)
	if (err) != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	copyPath := path.Join(tempDir, messageName)
	copy, err := os.Create(copyPath)
	if (err) != nil {
		log.Fatal(err)
	}
	defer copy.Close()
	_, err = io.Copy(copy, message)
	if (err) != nil {
		log.Fatal(err)
	}
}
