package xdomea

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"lath/xman/internal/db"
	"lath/xman/internal/tasks"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

// MAX_CONCURRENT_CALLS is the maximum number of simultaneous connections to the
// format verification endpoint. Connections are shared between all users and
// messages.
const MAX_CONCURRENT_CALLS = 10

var borgEndpoint = os.Getenv("BORG_ENDPOINT")
var tr http.Transport = http.Transport{}
var client http.Client = http.Client{
	Timeout:   time.Second * 60,
	Transport: &tr,
}
var guard = make(chan struct{}, MAX_CONCURRENT_CALLS)

func VerifyFileFormats(process db.SubmissionProcess, message db.Message) error {
	log.Printf("Starting VerifyFileFormats for process %v...\n", process.ProcessID)
	rootRecords := db.FindRootRecords(context.Background(), process.ProcessID, db.MessageType0503)
	primaryDocuments := GetPrimaryDocuments(&rootRecords)
	task := tasks.Start(
		db.ProcessStepFormatVerification,
		process.ProcessID,
		fmt.Sprintf("0 / %d", len(primaryDocuments)),
	)
	var wg sync.WaitGroup
	errorMessages := make([]string, 0)
	itemsComplete := 0
	for _, primaryDocument := range primaryDocuments {
		// Suppress warning about loop-variable scope. Actual problem is fixed
		// since go 1.22. Can be removed when tooling is updated to not show the
		// warning anymore.
		primaryDocument := primaryDocument
		wg.Add(1)
		guard <- struct{}{} // would block if guard channel is already filled
		go func() {
			defer func() {
				wg.Done()
				<-guard
			}()
			err := verifyDocument(message, primaryDocument)
			if err != nil {
				errorMessage := fmt.Sprintf("%s\n%s", primaryDocument.Filename, err.Error())
				errorMessages = append(errorMessages, errorMessage)
			}
			itemsComplete++
			tasks.Progress(task, fmt.Sprintf("%d / %d", itemsComplete, len(primaryDocuments)))
		}()
	}
	wg.Wait()
	if len(errorMessages) == 0 {
		tasks.MarkDone(task, "")
	} else {
		return tasks.MarkFailed(&task, strings.Join(errorMessages, "\n\n"))
	}
	log.Printf("VerifyFileFormats for process %v done\n", process.ProcessID)
	return nil
}

// verifyDocument runs format verification on the given document using the
// remote BORG service and saves the result to the document object.
func verifyDocument(message db.Message, primaryDocument db.PrimaryDocument) error {
	filePath := path.Join(message.StoreDir, primaryDocument.Filename)
	_, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fw, err := writer.CreateFormFile("file", primaryDocument.Filename)
	if err != nil {
		return err
	}
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	_, err = io.Copy(fw, file)
	if err != nil {
		return err
	}
	writer.Close()
	request, err := http.NewRequest("POST", borgEndpoint, bytes.NewReader(body.Bytes()))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("POST \"%s\": %d", borgEndpoint, response.StatusCode)
	}
	var parsedResponse db.FormatVerification
	err = json.NewDecoder(response.Body).Decode(&parsedResponse)
	if err != nil {
		return err
	}
	db.UpdatePrimaryDocumentFormatVerification(message.MessageHead.ProcessID, primaryDocument.Filename, &parsedResponse)
	return nil
}
