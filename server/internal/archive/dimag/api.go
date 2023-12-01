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

	"github.com/google/uuid"
)

var DimagApiEndpoint = os.Getenv("DIMAG_CORE_SOAP_ENDPOINT")
var DimagApiUser = os.Getenv("DIMAG_CORE_USER")
var DimagApiPassword = os.Getenv("DIMAG_CORE_PASSWORD")

func ImportMessage(messageID uuid.UUID) error {
	message, err := db.GetMessageByID(messageID)
	if err != nil {
		log.Println(err)
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
	return nil
}

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
		process, err := db.GetProcessByXdomeaID(message.MessageHead.ProcessID)
		if err != nil {
			log.Println(err)
			return err
		}
		processStep := process.ProcessState.Archiving
		processStep.Complete = true
		processStep.CompletionTime = time.Now()
		err = db.UpdateProcessStep(processStep)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}
