package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/fs"
	"lath/xman/internal/db"
	"lath/xman/internal/errors"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/studio-b12/gowebdav"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// control of the watch loop for the transfer directories
var ticker time.Ticker

type unknownFilesError []string

func (err unknownFilesError) Error() string {
	return "unknown files"
}

// TestTransferDir checks if an transfer directory configuration is works.
func TestTransferDir(transferDir db.TransferDir) TestResult {
	switch transferDir.Protocol {
	case db.ProtocolFile:
		return testLocalFilesystem(transferDir)
	case db.ProtocolWebDAV, db.ProtocolWebDAVSecure:
		return testWebDAV(transferDir)
	default:
		panic("unknown transfer directory scheme")
	}
}

type TestResult struct {
	Success           bool `json:"success"`
	ConnectionSuccess bool `json:"connectionSuccess"`
	Path0502Exists    bool `json:"path0502Exists"`
	Path0504Exists    bool `json:"path0504Exists"`
	Path0506Exists    bool `json:"path0506Exists"`
	Path0507Exists    bool `json:"path0507Exists"`
}

// testLocalFilesystem checks if an transfer directory configuration for a local filesystem works.
func testLocalFilesystem(transferDir db.TransferDir) TestResult {
	result := TestResult{}
	root := filepath.Join("/", transferDir.Path)
	_, err := os.ReadDir(root)
	result.ConnectionSuccess = err == nil
	_, err = os.Stat(filepath.Join(root, transferDir.Path0502))
	result.Path0502Exists = err == nil
	_, err = os.Stat(filepath.Join(root, transferDir.Path0504))
	result.Path0504Exists = err == nil
	_, err = os.Stat(filepath.Join(root, transferDir.Path0506))
	result.Path0506Exists = err == nil
	_, err = os.Stat(filepath.Join(root, transferDir.Path0507))
	result.Path0507Exists = err == nil
	result.Success = result.ConnectionSuccess &&
		result.Path0502Exists &&
		result.Path0504Exists &&
		result.Path0506Exists &&
		result.Path0507Exists
	return result
}

// testWebDAV checks if an transfer directory configuration for a webDAV works.
func testWebDAV(transferDir db.TransferDir) TestResult {
	result := TestResult{
		Path0502Exists: true,
		Path0504Exists: true,
		Path0506Exists: true,
		Path0507Exists: true,
	}
	client, err := connectWebDAV(transferDir)
	result.ConnectionSuccess = err == nil
	if !result.ConnectionSuccess {
		return result
	}
	if transferDir.Path0502 != "" {
		result.Path0502Exists = existsDir(client, transferDir.Path0502)
	}
	if transferDir.Path0504 != "" {
		result.Path0504Exists = existsDir(client, transferDir.Path0504)
	}
	if transferDir.Path0506 != "" {
		result.Path0506Exists = existsDir(client, transferDir.Path0506)
	}
	if transferDir.Path0507 != "" {
		result.Path0507Exists = existsDir(client, transferDir.Path0507)
	}
	result.Success = result.ConnectionSuccess &&
		result.Path0502Exists &&
		result.Path0504Exists &&
		result.Path0506Exists &&
		result.Path0507Exists
	return result
}

// existsDir checks if a directory exists on given webDAV client
func existsDir(client *gowebdav.Client, dir string) bool {
	_, err := client.Stat(getWebDAVPath(dir))
	return err == nil
}

// getWebDAVPath adds a trailing / to a directory path.
// The webDAV standard expects a trailing / for directories.
func getWebDAVPath(messageDir string) string {
	if !strings.HasSuffix(messageDir, "/") {
		return messageDir + "/"
	}
	return messageDir
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
			var err error
			errorData.Agency = &agency
			switch agency.TransferDir.Protocol {
			case db.ProtocolFile:
				err = readMessagesFromFilesystem(agency)
			case db.ProtocolWebDAV, db.ProtocolWebDAVSecure:
				err = readMessagesFromWebDAV(agency)
			default:
				panic("unknown transfer directory scheme")
			}
			hasUnknownFiles := updateUnknownFilesError(agency, unknownFilesErrors, err)
			// Handle errors other than unknown files.
			hasError := err != nil && !hasUnknownFiles
			if knownError, hasKnownError := accessErrors[agency.ID]; hasError && !hasKnownError {
				errors.AddProcessingErrorWithData(err, errorData)
			} else if hasKnownError && !hasError {
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
func readMessagesFromFilesystem(agency db.Agency) error {
	rootDir := filepath.Join("/", agency.TransferDir.Path)
	files, err := os.ReadDir(rootDir)
	if err != nil {
		return err
	}
	processedPaths := getProcessedTransferFiles(agency.ID)
	var unknownFiles []string
	for _, file := range files {
		fileInfo, err := file.Info()
		if err != nil {
			return err
		}
		// ignored files
		if processedPaths[file.Name()] ||
			file.Name() == ".gitkeep" ||
			isMessagePath(agency.TransferDir, fileInfo) {
			continue
		}
		// unknown files
		if file.IsDir() || !isMessage(file.Name()) {
			unknownFiles = append(unknownFiles, file.Name())
			continue
		}
		processID := getProcessID(file.Name())
		db.InsertTransferFile(agency.ID, &processID, file.Name())
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

// isMessagePath returns true if the directory is a configured path for messages
func isMessagePath(transferDir db.TransferDir, dir fs.FileInfo) bool {
	return dir.IsDir() && (dir.Name() == transferDir.Path0502 ||
		dir.Name() == transferDir.Path0504 ||
		dir.Name() == transferDir.Path0506 ||
		dir.Name() == transferDir.Path0507)
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
func readMessagesFromWebDAV(agency db.Agency) error {
	client, err := connectWebDAV(agency.TransferDir)
	if err != nil {
		return err
	}
	files, err := client.ReadDir("/")
	if err != nil {
		return err
	}
	processedPaths := getProcessedTransferFiles(agency.ID)
	var unknownFiles []string
	for _, file := range files {
		if processedPaths[file.Name()] || isMessagePath(agency.TransferDir, file) {
			continue
		}
		if file.IsDir() || !isMessage(file.Name()) {
			unknownFiles = append(unknownFiles, file.Name())
			continue
		}
		processID := getProcessID(file.Name())
		db.InsertTransferFile(agency.ID, &processID, file.Name())
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
func CopyMessageToTransferDirectory(
	agency db.Agency,
	processID *string,
	tempMessagePath string,
	messageType db.MessageType,
) error {
	switch agency.TransferDir.Protocol {
	case db.ProtocolFile:
		return copyMessageToLocalFilesystem(agency, processID, tempMessagePath, messageType)
	case db.ProtocolWebDAV, db.ProtocolWebDAVSecure:
		return copyMessageToWebDAV(agency, processID, tempMessagePath, messageType)
	default:
		panic("unknown transfer directory scheme")
	}
}

// copyMessageToLocalFilesystem copies a file from the local filesystem to another path in the local filesystem.
func copyMessageToLocalFilesystem(
	agency db.Agency,
	processID *string,
	tempMessagePath string,
	messageType db.MessageType,
) error {
	rootDir := filepath.Join("/", agency.TransferDir.Path)
	messageDir := getRemoteMessageDir(agency, messageType)
	messageFilename := path.Base(tempMessagePath)
	relMessagePath := filepath.Join(messageDir, messageFilename)
	absMessagePath := path.Join(rootDir, relMessagePath)
	// mark message as known, so it will not be added to unknown files on the transfer directory
	ok := db.InsertTransferFile(agency.ID, processID, relMessagePath)
	if !ok {
		return errTransferFileExists
	}
	messageFile, err := os.Open(tempMessagePath)
	if err != nil {
		// the copy process for the message failed
		// unmark the message as known, so it can be added in future
		db.DeleteTransferFile(agency.ID, relMessagePath)
		panic(err)
	}
	defer messageFile.Close()
	messageInTransferDir, err := os.Create(absMessagePath)
	if err != nil {
		db.DeleteTransferFile(agency.ID, relMessagePath)
		panic(err)
	}
	defer messageInTransferDir.Close()
	_, err = io.Copy(messageInTransferDir, messageFile)
	if err != nil {
		db.DeleteTransferFile(agency.ID, relMessagePath)
		panic(err)
	}
	return nil
}

// copyMessageToWebDAV copies a file from the local filesystem to a webDAV.
func copyMessageToWebDAV(
	agency db.Agency,
	processID *string,
	tempMessagePath string,
	messageType db.MessageType,
) error {
	client, err := connectWebDAV(agency.TransferDir)
	if err != nil {
		panic(err)
	}
	messageDir := getRemoteMessageDir(agency, messageType)
	webDAVFilePath := path.Join(messageDir, path.Base(tempMessagePath))
	// mark message as known, so it will not be added to unknown files on the transfer directory
	ok := db.InsertTransferFile(agency.ID, processID, webDAVFilePath)
	if !ok {
		return errTransferFileExists
	}
	messageFile, err := os.Open(tempMessagePath)
	if err != nil {
		// the copy process for the message failed
		// unmark the message as known, so it can be added in future
		db.DeleteTransferFile(agency.ID, webDAVFilePath)
		panic(err)
	}
	defer messageFile.Close()
	err = client.WriteStream(webDAVFilePath, messageFile, 0644)
	if err != nil {
		db.DeleteTransferFile(agency.ID, webDAVFilePath)
		panic(err)
	}
	return nil
}

// getRemoteMessageDir returns the configured message path for message type
func getRemoteMessageDir(agency db.Agency, messageType db.MessageType) string {
	var messageDir string
	switch messageType {
	case db.MessageType0502:
		return agency.TransferDir.Path0502
	case db.MessageType0504:
		return agency.TransferDir.Path0504
	case db.MessageType0506:
		messageDir = agency.TransferDir.Path0506
	case db.MessageType0507:
		messageDir = agency.TransferDir.Path0507
	}
	return path.Clean(messageDir)
}

// CopyMessageFromTransferDirectory copies a file from a transfer directory to a temporary directory.
func CopyMessageFromTransferDirectory(agency db.Agency, messagePath string) string {
	switch agency.TransferDir.Protocol {
	case db.ProtocolFile:
		return copyFileFromLocalFilesystem(agency.TransferDir, messagePath)
	case db.ProtocolWebDAV, db.ProtocolWebDAVSecure:
		return copMessageFromWebDAV(agency.TransferDir, messagePath)
	default:
		panic("unknown transfer directory scheme")
	}
}

// copMessageFromWebDAV copies the file specified by webDAVFilePath from a webDAV to a temporary directory.
// The copied file is locally stored in a temporary directory.
// The caller of this function should remove the temporary directory.
//
// Returns the local path of the copied file.
func copMessageFromWebDAV(transferDir db.TransferDir, webDAVFilePath string) string {
	client, err := connectWebDAV(transferDir)
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
func copyFileFromLocalFilesystem(transferDir db.TransferDir, messagePath string) string {
	processID := getProcessID(messagePath)
	messageName := filepath.Base(messagePath)
	// Create temporary directory. The name of the directory contains the message ID.
	tempDir, err := os.MkdirTemp("", processID)
	if err != nil {
		panic(err)
	}
	transferDirPath := filepath.Join("/", transferDir.Path, messagePath)
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
	log.Printf("Removing file from transfer dir for %s: %s\n", agency.Name, path)
	switch agency.TransferDir.Protocol {
	case db.ProtocolFile:
		RemoveFileFromLocalFilesystem(agency.TransferDir, path)
	case db.ProtocolWebDAV, db.ProtocolWebDAVSecure:
		RemoveFileFromWebDAV(agency.TransferDir, path)
	default:
		panic("unknown transfer directory scheme")
	}
	db.DeleteTransferFile(agency.ID, path)
}

// RemoveFileFromLocalFilesystem deletes a file on a local filesystem.
func RemoveFileFromLocalFilesystem(transferDir db.TransferDir, path string) {
	fullPath := filepath.Join("/", transferDir.Path, path)
	err := os.RemoveAll(fullPath)
	if err != nil {
		panic(err)
	}
}

// RemoveFileFromWebDAV deletes a file on a webDAV.
func RemoveFileFromWebDAV(transferDir db.TransferDir, path string) {
	client, err := connectWebDAV(transferDir)
	if err != nil {
		panic(err)
	}
	err = client.RemoveAll(path)
	if err != nil {
		panic(err)
	}
}

// connectWebDAV creates a client from an parsed transfer directory URL.
// Checks if a connection with the transfer directory with the given configuration is possible.
func connectWebDAV(transferDir db.TransferDir) (*gowebdav.Client, error) {
	var url string
	var protocol string
	switch transferDir.Protocol {
	case db.ProtocolWebDAV:
		protocol = "http://"
	case db.ProtocolWebDAVSecure:
		protocol = "https://"
	default:
		return nil, fmt.Errorf("unknown transfer directory protocol %s", transferDir.Protocol)
	}
	url = protocol + path.Join(transferDir.Host, transferDir.Path)
	var client *gowebdav.Client
	tlsConfig := &tls.Config{
		InsecureSkipVerify: transferDir.AllowInsecureTLS,
	}
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	client = gowebdav.NewClient(url, transferDir.User, transferDir.Password)
	client.SetTransport(transport)
	err := client.Connect()
	return client, err
}
