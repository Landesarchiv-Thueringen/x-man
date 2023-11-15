package format

import (
	"bytes"
	"encoding/json"
	"io"
	"lath/xman/internal/db"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
)

var BorgEndpoint = "http://localhost:3330/analyse-file"

func VerifyFileFormats(messageID uuid.UUID) {
	message, err := db.GetMessageByID(messageID)
	if err != nil {
		log.Fatal(err)
	}
	primaryDocuments, err := db.GetAllPrimaryDocuments(messageID)
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{
		Timeout: time.Second * 60,
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
	}
	message.FormatVerificationComplete = true
	err = db.UpdateMessage(message)
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
