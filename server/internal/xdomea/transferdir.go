package xdomea

import (
	"io"
	"lath/xman/internal/db"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/studio-b12/gowebdav"
)

type URLScheme string

const (
	Local     URLScheme = "file"
	WebDAV    URLScheme = "dav"
	WebDAVSec URLScheme = "davs"
)

var ticker time.Ticker
var stop chan bool

func TestTransferDir(dir string) bool {
	_, err := os.ReadDir(dir)
	return err == nil
}

func MonitorTransferDirs() {
	ticker = *time.NewTicker(time.Second * 5)
	stop = make(chan bool)
	go watchLoop(ticker, stop)
}

func watchLoop(timer time.Ticker, stop chan bool) {
	for {
		select {
		case <-stop:
			timer.Stop()
			return
		case <-timer.C:
			readMessages()
		}
	}
}

func readMessages() {
	agencies := db.GetAgencies()
	for _, agency := range agencies {
		transferDirURL, err := url.Parse(agency.TransferDirURL)
		if err != nil {
			panic(err)
		}
		switch transferDirURL.Scheme {
		case string(Local):
			readMessagesFromLocalFilesystem(agency, transferDirURL)
		case string(WebDAV):
			readMessagesFromWebDAV(agency, transferDirURL)
		case string(WebDAVSec):
			readMessagesFromWebDAV(agency, transferDirURL)
		default:
			panic("unknown transfer directory scheme")
		}
	}
}

func readMessagesFromLocalFilesystem(agency db.Agency, transferDirURL *url.URL) {
	rootDir := filepath.Join(transferDirURL.Path)
	files, err := os.ReadDir(rootDir)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if !db.IsMessageAlreadyProcessed(file.Name()) && !file.IsDir() && IsMessage(file.Name()) {
			log.Println("Processing new message " + file.Name())
			go ProcessNewMessage(agency, file.Name())
		}
	}
}

func readMessagesFromWebDAV(agency db.Agency, transferDirURL *url.URL) {
	client, err := getWebDAVClient(transferDirURL)
	if err != nil {
		panic(err)
	}
	path := transferDirURL.Path
	files, err := client.ReadDir(path)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if !db.IsMessageAlreadyProcessed(file.Name()) && !file.IsDir() && IsMessage(file.Name()) {
			log.Println("Processing new message " + file.Name())
			go ProcessNewMessage(agency, file.Name())
		}
	}
}

func getWebDAVClient(transferDirURL *url.URL) (*gowebdav.Client, error) {
	var root string
	switch transferDirURL.Scheme {
	case string(WebDAV):
		root = "http://" + transferDirURL.Host + "/" + transferDirURL.Path
	case string(WebDAVSec):
		root = "https://" + transferDirURL.Host + "/" + transferDirURL.Path
	default:
		panic("unknown transfer directory scheme")
	}
	user := transferDirURL.User.Username()
	password, set := transferDirURL.User.Password()
	if !set {
		password = ""
	}
	client := gowebdav.NewClient(root, user, password)
	err := client.Connect()
	return client, err
}

func CopyMessageToTransferDirectory(agency db.Agency, messagePath string) string {
	transferDirURL, err := url.Parse(agency.TransferDirURL)
	if err != nil {
		panic(err)
	}
	switch transferDirURL.Scheme {
	case string(Local):
		return copyMessageToLocalFilesystem(transferDirURL, messagePath)
	case string(WebDAV):
		return copyMessageToWebDAV(transferDirURL, messagePath)
	case string(WebDAVSec):
		return copyMessageToWebDAV(transferDirURL, messagePath)
	default:
		panic("unknown transfer directory scheme")
	}
}

func copyMessageToLocalFilesystem(transferDirURL *url.URL, messagePath string) string {
	messageFilename := path.Base(messagePath)
	messageFile, err := os.Open(messagePath)
	if err != nil {
		panic(err)
	}
	defer messageFile.Close()
	messageTransferDirPath := path.Join(transferDirURL.Path, messageFilename)
	messageInTransferDir, err := os.Create(messageTransferDirPath)
	if err != nil {
		panic(err)
	}
	defer messageInTransferDir.Close()
	_, err = io.Copy(messageInTransferDir, messageFile)
	if err != nil {
		panic(err)
	}
	return messageFilename
}

func copyMessageToWebDAV(transferDirURL *url.URL, messagePath string) string {
	client, err := getWebDAVClient(transferDirURL)
	if err != nil {
		panic(err)
	}
	webdavFilePath := path.Base(messagePath)
	messageFile, err := os.Open(messagePath)
	if err != nil {
		panic(err)
	}
	defer messageFile.Close()
	err = client.WriteStream(webdavFilePath, messageFile, 0644)
	if err != nil {
		panic(err)
	}
	return webdavFilePath
}

func CopyMessageFromTransferDirectory(agency db.Agency, messagePath string) string {
	transferDirURL, err := url.Parse(agency.TransferDirURL)
	if err != nil {
		panic(err)
	}
	switch transferDirURL.Scheme {
	case string(Local):
		return copyFileFromLocalFilesystem(transferDirURL, messagePath)
	case string(WebDAV):
		return copMessageFromWebDAV(transferDirURL, messagePath)
	case string(WebDAVSec):
		return copMessageFromWebDAV(transferDirURL, messagePath)
	default:
		panic("unknown transfer directory scheme")
	}
}

// copMessageFromWebDAV copies the file specified by webDAVFilePath.
// The copied file is localy stored in a temporary directory.
// The caller of this function should remove the temporary directory.
//
// Returns the local path of the copied file.
func copMessageFromWebDAV(transferDirURL *url.URL, webDAVFilePath string) string {
	client, err := getWebDAVClient(transferDirURL)
	if err != nil {
		panic(err)
	}
	reader, err := client.ReadStream(webDAVFilePath)
	if err != nil {
		panic(err)
	}
	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		panic(err)
	}
	fileName := filepath.Base(webDAVFilePath)
	filePath := filepath.Join(tempDir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	_, err = io.Copy(file, reader)
	if err != nil {
		panic(err)
	}
	return filePath
}

// copyFileFromLocalFilesystem copies the file specified by messagePath.
// The copied file is localy stored in a temporary directory.
// The caller of this function should remove the temporary directory.
//
// Returns the local path of the copied file.
func copyFileFromLocalFilesystem(transferDirURL *url.URL, messagePath string) string {
	processID := GetMessageID(messagePath)
	messageName := filepath.Base(messagePath)
	// Create temporary directory. The name of the directory contains the message ID.
	tempDir, err := os.MkdirTemp("", processID)
	if err != nil {
		panic(err)
	}
	transferDirPath := filepath.Join(transferDirURL.Path, messagePath)
	// Open the original messageFile in the transfer directory.
	messageFile, err := os.Open(transferDirPath)
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
	return copyPath
}

func RemoveFileFromTransferDir(agency db.Agency, path string) {
	transferDirURL, err := url.Parse(agency.TransferDirURL)
	if err != nil {
		panic(err)
	}
	switch transferDirURL.Scheme {
	case string(Local):
		RemoveFileFromLocalFilesystem(transferDirURL, path)
	case string(WebDAV):
		RemoveFileFromWebDAV(transferDirURL, path)
	case string(WebDAVSec):
		RemoveFileFromWebDAV(transferDirURL, path)
	default:
		panic("unknown transfer directory scheme")
	}
}

func RemoveFileFromLocalFilesystem(transferDirURL *url.URL, path string) {
	fullPath := filepath.Join(transferDirURL.Path, path)
	err := os.Remove(fullPath)
	if err != nil {
		panic(err)
	}
}

func RemoveFileFromWebDAV(transferDirURL *url.URL, path string) {
	client, err := getWebDAVClient(transferDirURL)
	if err != nil {
		panic(err)
	}
	err = client.Remove(path)
	if err != nil {
		panic(err)
	}
}
