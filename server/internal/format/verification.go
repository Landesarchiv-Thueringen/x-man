package format

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"lath/xman/internal/db"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"time"
)

var BorgEndpoint = "https://borg.tsa.thlv.de/analyse-file"
var tr http.Transport = http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}
var client http.Client = http.Client{
	Timeout:   time.Second * 60,
	Transport: &tr,
}

func VerifyFileFormats(process db.Process, message db.Message) {
	primaryDocuments, err := db.GetAllPrimaryDocuments(message.ID)
	if err != nil {
		log.Fatal(err)
	}
	processStep := process.ProcessState.FormatVerification
	processStep.ItemCount = uint(len(primaryDocuments))
	err = db.UpdateProcessStep(processStep)
	if err != nil {
		log.Fatal(err)
	}
	for _, primaryDocument := range primaryDocuments {
		filePath := path.Join(message.StoreDir, primaryDocument.FileName)
		_, err := os.Stat(filePath)
		if err != nil {
			log.Fatal(err)
		}
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		fw, err := writer.CreateFormFile("file", primaryDocument.FileName)
		if err != nil {
			log.Fatal(err)
		}
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatal(err)
		}
		_, err = io.Copy(fw, file)
		if err != nil {
			log.Fatal(err)
		}
		writer.Close()
		request, err := http.NewRequest("POST", BorgEndpoint, bytes.NewReader(body.Bytes()))
		if err != nil {
			log.Fatal(err)
		}
		request.Header.Set("Content-Type", writer.FormDataContentType())
		response, err := client.Do(request)
		if err != nil {
			log.Fatal(err)
		}
		if response.StatusCode != http.StatusOK {
			log.Println(response.StatusCode)
		}
		var parsedResponse db.FormatVerification
		err = json.NewDecoder(response.Body).Decode(&parsedResponse)
		if err != nil {
			log.Fatal(err)
		}
		prepareForDatabase(&parsedResponse)
		primaryDocument.FormatVerification = &parsedResponse
		err = db.UpdatePrimaryDocument(primaryDocument)
		if err != nil {
			log.Fatal(err)
		}
		processStep.ItemCompletetCount = processStep.ItemCompletetCount + 1
		err = db.UpdateProcessStep(processStep)
		if err != nil {
			log.Fatal(err)
		}
		message.VerificationCompleteCount = message.VerificationCompleteCount + 1
		err = db.UpdateMessage(message)
		if err != nil {
			log.Fatal(err)
		}
	}
	message.FormatVerificationComplete = true
	err = db.UpdateMessage(message)
	if err != nil {
		log.Fatal(err)
	}
	processStep.Complete = true
	processStep.CompletionTime = time.Now()
	err = db.UpdateProcessStep(processStep)
	if err != nil {
		log.Fatal(err)
	}
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
