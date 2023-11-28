package dimag

import (
	"encoding/xml"
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

func ImportMessage(messageID uuid.UUID) {
	endpoint := os.Getenv("DIMAG_CORE_SOAP_ENDPOINT")
	user := os.Getenv("DIMAG_CORE_USER")
	password := os.Getenv("DIMAG_CORE_PASSWORD")
	message, err := db.GetMessageByID(messageID)
	if err != nil {
		log.Fatal(err)
	}
	err = InitConnection()
	if err != nil {
		log.Fatal("couldn't init connection to DIMAG sftp server")
	}
	defer CloseConnection()
	fileRecordObjects, err := db.GetAllFileRecordObjects(message.ID)
	if err != nil {
		log.Fatal(err)
	}
	for _, fileRecordObject := range fileRecordObjects {
		importDir, err := uploadFileRecordObjectFiles(sftpClient, message, fileRecordObject)
		if err != nil {
			log.Fatal(err)
		}
		requestMetadata := SoapImportDoc{
			UserName:        user,
			Password:        password,
			ControlFilePath: filepath.Join(importDir, ControlFileName),
		}
		soapRequest := SoapEnvelopeImportDoc{
			SoapNs:  SoapNs,
			DimagNs: DimagNs,
			Header:  SoapEnvelopeHeader{},
			Body: SoapEnvelopeBodyImportDoc{
				ImportDoc: requestMetadata,
			},
		}
		xmlBytes, err := xml.MarshalIndent(soapRequest, " ", " ")
		if err != nil {
			log.Fatal(err)
		}
		requestString := string(xmlBytes)
		req, err := http.NewRequest("POST", endpoint, strings.NewReader(requestString))
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("Content-Type", "text/xml;charset=UTF-8")
		req.Header.Set("SOAPAction", "importDoc")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		// Read the response body
		if resp.StatusCode == http.StatusOK {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			log.Println(string(body))
			process, err := db.GetProcessByXdomeaID(message.MessageHead.ProcessID)
			if err != nil {
				log.Fatal(err)
			}
			processStep := process.ProcessState.Archiving
			processStep.Complete = true
			processStep.CompletionTime = time.Now()
			err = db.UpdateProcessStep(processStep)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
