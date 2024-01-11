package dimag

import (
	"encoding/xml"
	"errors"
	"io/ioutil"
	"lath/xman/internal/db"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var DimagApiEndpoint = os.Getenv("DIMAG_CORE_SOAP_ENDPOINT")
var DimagApiUser = os.Getenv("DIMAG_CORE_USER")
var DimagApiPassword = os.Getenv("DIMAG_CORE_PASSWORD")

// ImportMessage archives 0503 message in DIMAG.
func ImportMessage(process db.Process, message db.Message) error {
	processStep := process.ProcessState.Archiving
	startTime := time.Now()
	processStep.StartTime = &startTime
	err := db.UpdateProcessStep(processStep)
	if err != nil {
		return err
	}
	err = InitConnection()
	if err != nil {
		log.Println("couldn't init connection to DIMAG sftp server")
		return err
	}
	defer CloseConnection()
	fileRecordObjects, err := db.GetAllFileRecordObjects(message.ID)
	if err != nil {
		log.Println(err)
		return err
	}
	for _, fileRecordObject := range fileRecordObjects {
		err = importFileRecordObject(message, fileRecordObject)
		if err != nil {
			return err
		}
	}
	processStep.Complete = true
	processStep.CompletionTime = time.Now()
	return db.UpdateProcessStep(processStep)
}

// importFileRecordObject archives a file record object in DIMAG.
func importFileRecordObject(message db.Message, fileRecordObject db.FileRecordObject) error {
	importDir, err := uploadFileRecordObjectFiles(sftpClient, message, fileRecordObject)
	if err != nil {
		return err
	}
	requestMetadata := ImportDoc{
		UserName:        DimagApiUser,
		Password:        DimagApiPassword,
		ControlFilePath: filepath.Join(importDir, ControlFileName),
	}
	soapRequest := EnvelopeImportDoc{
		SoapNs:  SoapNs,
		DimagNs: DimagNs,
		Header:  EnvelopeHeader{},
		Body: EnvelopeBodyImportDoc{
			ImportDoc: requestMetadata,
		},
	}
	xmlBytes, err := xml.MarshalIndent(soapRequest, " ", " ")
	if err != nil {
		log.Println(err)
		return err
	}
	requestString := string(xmlBytes)
	req, err := http.NewRequest("POST", DimagApiEndpoint, strings.NewReader(requestString))
	if err != nil {
		log.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "text/xml;charset=UTF-8")
	req.Header.Set("SOAPAction", "importDoc")
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	defer response.Body.Close()
	return processImportResponse(message, response)
}

func processImportResponse(message db.Message, response *http.Response) error {
	if response.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Println(err)
			return err
		}
		var parsedResponse EnvelopeImportDocResponse
		err = xml.Unmarshal(body, &parsedResponse)
		if err != nil {
			log.Println(err)
			return err
		}
		if parsedResponse.Body.ImportDocResponse.Status != 200 {
			log.Println(parsedResponse.Body.ImportDocResponse.Message)
			return errors.New("DIMAG ingest error")
		}
	}
	return nil
}
