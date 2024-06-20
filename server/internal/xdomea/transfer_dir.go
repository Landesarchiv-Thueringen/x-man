package xdomea

import (
	"context"
	"io"
	"io/fs"
	"lath/xman/internal/db"
	"lath/xman/internal/errors"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/studio-b12/gowebdav"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// all possible URL protocol schemes for transfer directories
type URLScheme string

const (
	Local     URLScheme = "file"
	WebDAV    URLScheme = "dav"
	WebDAVSec URLScheme = "davs"
)

// control of the watch loop for the transfer directories
var ticker time.Ticker
var stop chan bool

type unknownFilesError []string

func (err unknownFilesError) Error() string {
	return "unknown files"
}

// TestTransferDir checks if an transfer directory configuration is works.
func TestTransferDir(testURL string) bool {
	transferDirURL, err := url.Parse(testURL)
	if err != nil {
		return false
	}
	switch transferDirURL.Scheme {
	case string(Local):
		return testLocalFilesystem(transferDirURL)
	case string(WebDAV):
		fallthrough
	case string(WebDAVSec):
		return testWebDAV(transferDirURL)
	default:
		panic("unknown transfer directory scheme")
	}
}

// testLocalFilesystem checks if an transfer directory configuration for a local filesystem works.
func testLocalFilesystem(transferDirURL *url.URL) bool {
	_, err := os.ReadDir(transferDirURL.Path)
	return err == nil
}

// testWebDAV checks if an transfer directory configuration for a webDAV works.
func testWebDAV(transferDirURL *url.URL) bool {
	_, err := connectWebDAV(transferDirURL)
	return err == nil
}

// MonitorTransferDirs starts the watch loop to process the contents of the transfer directories.
func MonitorTransferDirs() {
	defer errors.HandlePanic("MonitorTransferDirs", nil)
	interval := time.Minute
	intervalString := os.Getenv("TRANSFER_DIR_SCAN_INTERVAL_SECONDS")
	if intervalString != "" {
		intervalSeconds, err := strconv.Atoi(intervalString)
		if err != nil {
			panic(err)
		}
		interval = time.Second * time.Duration(intervalSeconds)
	}
	ticker = *time.NewTicker(interval)

	errorData := db.ProcessingError{
		Title:     "Fehler beim Lesen des Transferverzeichnisses",
		ErrorType: "access-transfer-dir",
	}
	// Regularly check all transfer dirs.
	for {
		<-ticker.C
		// accessErrors maps agency IDs to known errors, so we won't add
		// existing errors again and we can mark errors as resolved when they stop
		// occurring.
		unknownFilesErrors := getUnknownFilesErrors()
		accessErrors := getAccessErrors()
		agencies := db.FindAgencies(context.Background())
		for _, agency := range agencies {
			errorData.Agency = &agency
			transferDirURL, err := url.Parse(agency.TransferDirURL)
			if err != nil {
				panic(err)
			}
			switch transferDirURL.Scheme {
			case string(Local):
				err = readMessagesFromFilesystem(agency, transferDirURL)
			case string(WebDAV), string(WebDAVSec):
				err = readMessagesFromWebDAV(agency, transferDirURL)
			default:
				panic("unknown transfer directory scheme")
			}
			hasUnknownFiles := updateUnknownFilesError(agency, unknownFilesErrors, err)
			if knownError, hasKnownError := accessErrors[agency.ID]; !hasUnknownFiles && !hasKnownError && err != nil {
				errors.AddProcessingErrorWithData(err, errorData)
			} else if hasKnownError && err == nil {
				db.UpdateProcessingErrorResolve(knownError, db.ErrorResolutionObsolete)
			}
		}
	}
}

// updateUnknownFilesError takes the returned error from a readMessage function
// and updates or creates a processing error according to what the returned
// error indicates.
func updateUnknownFilesError(
	agency db.Agency,
	unknownFilesErrors map[primitive.ObjectID]db.ProcessingError,
	err error,
) bool {
	unknownFiles, hasUnknownFiles := err.(unknownFilesError)
	e, hasProcessingError := unknownFilesErrors[agency.ID]
	if hasProcessingError && hasUnknownFiles {
		// Update existing processing error if unknown files changed.
		if !reflect.DeepEqual(db.UnmarshalData[unknownFilesError](e.Data), unknownFiles) {
			e.Data = unknownFiles
			db.MustReplaceProcessingError(e)
		}
	} else if hasProcessingError && err == nil {
		// Unknown files have disappeared. Mark the processing error as solved.
		db.UpdateProcessingErrorResolve(e, db.ErrorResolutionObsolete)
	} else if !hasProcessingError && hasUnknownFiles {
		// Unknown files appeared. Created a processing error.
		errors.AddProcessingError(db.ProcessingError{
			Title:     "Unbekannte Dateien oder Ordner in Transferverzeichnis",
			ErrorType: "unknown-files-in-transfer-dir",
			Agency:    &agency,
			Data:      unknownFiles,
		})
	}
	return hasUnknownFiles
}

func getAccessErrors() map[primitive.ObjectID]db.ProcessingError {
	errors := db.FindUnresolvedProcessingErrorsByType(context.Background(), "access-transfer-dir")
	m := make(map[primitive.ObjectID]db.ProcessingError)
	for _, e := range errors {
		m[e.Agency.ID] = e
	}
	return m
}

func getUnknownFilesErrors() map[primitive.ObjectID]db.ProcessingError {
	errors := db.FindUnresolvedProcessingErrorsByType(context.Background(), "unknown-files-in-transfer-dir")
	m := make(map[primitive.ObjectID]db.ProcessingError)
	for _, e := range errors {
		m[e.Agency.ID] = e
	}
	return m
}

func getProcessedTransferFiles(agencyID primitive.ObjectID) map[string]bool {
	files := db.FindTransferDirFilesForAgency(agencyID)
	m := make(map[string]bool)
	for _, file := range files {
		m[file.Path] = true
	}
	return m
}

// readMessagesFromFilesystem checks if new messages exist for a local filesystem.
func readMessagesFromFilesystem(agency db.Agency, transferDirURL *url.URL) error {
	rootDir := filepath.Join(transferDirURL.Path)
	files, err := os.ReadDir(rootDir)
	if err != nil {
		return err
	}
	processedPaths := getProcessedTransferFiles(agency.ID)
	var unknownFiles []string
	for _, file := range files {
		if processedPaths[file.Name()] || file.Name() == ".gitkeep" {
			continue
		}
		if file.IsDir() || !isMessage(file.Name()) {
			unknownFiles = append(unknownFiles, file.Name())
			continue
		}
		db.InsertTransferFile(agency.ID, getProcessID(file.Name()), file.Name())
		go func() {
			defer errors.HandlePanic("readMessagesFromFilesystem", &db.ProcessingError{
				Agency:       &agency,
				TransferPath: file.Name(),
			})
			waitUntilStable(file)
			ProcessNewMessage(agency, file.Name())
		}()
	}
	if len(unknownFiles) > 0 {
		return unknownFilesError(unknownFiles)
	}
	return nil
}

// waitUntilStable regularly inspects the given file's stats for changes and
// returns as soon as the file stops changing on disk.
func waitUntilStable(file fs.DirEntry) {
	var modTime time.Time
	for {
		info, err := file.Info()
		if err != nil {
			panic(err)
		}
		if modTime == info.ModTime() {
			return
		}
		modTime = info.ModTime()
		time.Sleep(1 * time.Second)
	}
}

// readMessagesFromWebDAV checks if new messages exist for a webDAV.
func readMessagesFromWebDAV(agency db.Agency, transferDirURL *url.URL) error {
	client, err := connectWebDAV(transferDirURL)
	if err != nil {
		return err
	}
	path := transferDirURL.Path
	files, err := client.ReadDir(path)
	if err != nil {
		return err
	}
	processedPaths := getProcessedTransferFiles(agency.ID)
	var unknownFiles []string
	for _, file := range files {
		if processedPaths[file.Name()] {
			continue
		}
		if file.IsDir() || !isMessage(file.Name()) {
			unknownFiles = append(unknownFiles, file.Name())
			continue
		}
		db.InsertTransferFile(agency.ID, getProcessID(file.Name()), file.Name())
		go func() {
			defer errors.HandlePanic("readMessagesFromWebDAV", &db.ProcessingError{
				Agency:       &agency,
				TransferPath: file.Name(),
			})
			waitUntilStableWebDav(client, file)
			ProcessNewMessage(agency, file.Name())
		}()
	}
	if len(unknownFiles) > 0 {
		return unknownFilesError(unknownFiles)
	}
	return nil
}

// waitUntilStableWebDav regularly inspects the given file's stats for changes
// and returns as soon as the file has a non-null size, which indicates that its
// upload is complete.
func waitUntilStableWebDav(client *gowebdav.Client, file fs.FileInfo) {
	for {
		info, err := client.Stat(file.Name())
		if err != nil {
			panic(err)
		}
		if info.Size() > 0 {
			return
		}
		time.Sleep(1 * time.Second)
	}
}

// CopyMessageToTransferDirectory copies a file from the local filesystem to a transfer directory.
func CopyMessageToTransferDirectory(agency db.Agency, processID uuid.UUID, messagePath string) string {
	transferDirURL, err := url.Parse(agency.TransferDirURL)
	if err != nil {
		panic(err)
	}
	db.InsertTransferFile(agency.ID, processID, filepath.Base(messagePath))
	switch transferDirURL.Scheme {
	case string(Local):
		return copyMessageToLocalFilesystem(transferDirURL, messagePath)
	case string(WebDAV):
		fallthrough
	case string(WebDAVSec):
		return copyMessageToWebDAV(transferDirURL, messagePath)
	default:
		panic("unknown transfer directory scheme")
	}
}

// copyMessageToLocalFilesystem copies a file from the local filesystem to another path in the local filesystem.
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

// copyMessageToWebDAV copies a file from the local filesystem to a webDAV.
func copyMessageToWebDAV(transferDirURL *url.URL, messagePath string) string {
	client, err := connectWebDAV(transferDirURL)
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

// CopyMessageFromTransferDirectory copies a file from a transfer directory to a temporary directory.
func CopyMessageFromTransferDirectory(agency db.Agency, messagePath string) string {
	transferDirURL, err := url.Parse(agency.TransferDirURL)
	if err != nil {
		panic(err)
	}
	switch transferDirURL.Scheme {
	case string(Local):
		return copyFileFromLocalFilesystem(transferDirURL, messagePath)
	case string(WebDAV):
		fallthrough
	case string(WebDAVSec):
		return copMessageFromWebDAV(transferDirURL, messagePath)
	default:
		panic("unknown transfer directory scheme")
	}
}

// copMessageFromWebDAV copies the file specified by webDAVFilePath from a webDAV to a temporary directory.
// The copied file is localy stored in a temporary directory.
// The caller of this function should remove the temporary directory.
//
// Returns the local path of the copied file.
func copMessageFromWebDAV(transferDirURL *url.URL, webDAVFilePath string) string {
	client, err := connectWebDAV(transferDirURL)
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
// The copied file is locally stored in a temporary directory.
// The caller of this function should remove the temporary directory.
//
// Returns the local path of the copied file.
func copyFileFromLocalFilesystem(transferDirURL *url.URL, messagePath string) string {
	processID := getProcessID(messagePath)
	messageName := filepath.Base(messagePath)
	// Create temporary directory. The name of the directory contains the message ID.
	tempDir, err := os.MkdirTemp("", processID.String())
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

// RemoveFileFromTransferDir deletes a file on a transfer directory.
func RemoveFileFromTransferDir(agency db.Agency, path string) {
	transferDirURL, err := url.Parse(agency.TransferDirURL)
	if err != nil {
		panic(err)
	}
	switch transferDirURL.Scheme {
	case string(Local):
		RemoveFileFromLocalFilesystem(transferDirURL, path)
	case string(WebDAV), string(WebDAVSec):
		RemoveFileFromWebDAV(transferDirURL, path)
	default:
		panic("unknown transfer directory scheme")
	}
	db.DeleteTransferFile(agency.ID, path)
}

// RemoveFileFromLocalFilesystem deletes a file on a local filesystem.
func RemoveFileFromLocalFilesystem(transferDirURL *url.URL, path string) {
	fullPath := filepath.Join(transferDirURL.Path, path)
	err := os.Remove(fullPath)
	if err != nil {
		panic(err)
	}
}

// RemoveFileFromWebDAV deletes a file on a webDAV.
func RemoveFileFromWebDAV(transferDirURL *url.URL, path string) {
	client, err := connectWebDAV(transferDirURL)
	if err != nil {
		panic(err)
	}
	err = client.Remove(path)
	if err != nil {
		panic(err)
	}
}

// connectWebDAV creates a client from an parsed transfer directory URL.
// Checks if a connection with the transfer directory with the given configuration is possible.
func connectWebDAV(transferDirURL *url.URL) (*gowebdav.Client, error) {
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
