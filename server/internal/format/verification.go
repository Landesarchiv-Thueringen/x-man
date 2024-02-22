package format

import (
	"bytes"
	"crypto/tls"
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
var tr http.Transport = http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}
var client http.Client = http.Client{
	Timeout:   time.Second * 60,
	Transport: &tr,
}
var guard = make(chan struct{}, MAX_CONCURRENT_CALLS)

func VerifyFileFormats(process db.Process, message db.Message) {
	primaryDocuments, err := db.GetAllPrimaryDocuments(message.ID)
	if err != nil {
		panic(err)
	}
	task, err := tasks.Start(db.TaskTypeFormatVerification, process, uint(len(primaryDocuments)))
	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	errorMessages := make([]string, 0)
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
			err = verifyDocument(message.StoreDir, primaryDocument)
			if err != nil {
				errorMessage := fmt.Sprintf("%s: %s", primaryDocument.FileName, err.Error())
				errorMessages = append(errorMessages, errorMessage)
			}
			err = tasks.MarkItemComplete(&task)
			if err != nil {
				panic(err)
			}
		}()
	}
	wg.Wait()
	if len(errorMessages) == 0 {
		err = tasks.MarkDone(&task)
	} else {
		err = tasks.MarkFailed(&task, strings.Join(errorMessages, "\n"), true)
	}
	if err != nil {
		panic(err)
	}
}

// verifyDocument runs format verification on the given document using the
// remote BORG service and saves the result to the document object.
func verifyDocument(storeDir string, primaryDocument db.PrimaryDocument) error {
	filePath := path.Join(storeDir, primaryDocument.FileName)
	_, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fw, err := writer.CreateFormFile("file", primaryDocument.FileName)
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
		log.Println(response.StatusCode)
	}
	var parsedResponse db.FormatVerification
	err = json.NewDecoder(response.Body).Decode(&parsedResponse)
	if err != nil {
		return err
	}
	prepareForDatabase(&parsedResponse)
	primaryDocument.FormatVerification = &parsedResponse
	err = db.UpdatePrimaryDocument(primaryDocument)
	if err != nil {
		return err
	}
	return nil
}

func prepareForDatabase(formatVerification *db.FormatVerification) {
	var features []db.Feature
	for _, feature := range formatVerification.Summary {
		features = append(features, feature)
	}
	formatVerification.Features = features
	if formatVerification.FileIdentificationResults != nil {
		for toolIndex := range formatVerification.FileIdentificationResults {
			prepareToolResponseForDatabase(&formatVerification.FileIdentificationResults[toolIndex])
		}
	}
	if formatVerification.FileValidationResults != nil {
		for toolIndex := range formatVerification.FileValidationResults {
			prepareToolResponseForDatabase(&formatVerification.FileValidationResults[toolIndex])
		}
	}
}

func prepareToolResponseForDatabase(toolResponse *db.ToolResponse) {
	if toolResponse.ExtractedFeatures != nil {
		var extractedFeatures []db.ExtractedFeature
		for key, value := range *toolResponse.ExtractedFeatures {
			extractedFeature := db.ExtractedFeature{
				Key:   key,
				Value: value,
			}
			extractedFeatures = append(extractedFeatures, extractedFeature)
		}
		toolResponse.Features = extractedFeatures
	}
}
