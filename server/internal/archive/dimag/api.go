package dimag

import (
	"encoding/xml"
	"fmt"
	"io"
	"lath/xman/internal/db"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var DimagApiEndpoint = os.Getenv("DIMAG_CORE_SOAP_ENDPOINT")
var DimagApiUser = os.Getenv("DIMAG_CORE_USER")
var DimagApiPassword = os.Getenv("DIMAG_CORE_PASSWORD")

// ImportMessageSync archives all metadata and files of a 0503 message in DIMAG.
//
// The record objects in the message should be complete loaded.
//
// ImportMessageSync returns after the archiving process completed.
func ImportMessageSync(process db.Process, message db.Message, collection db.Collection) error {
	err := InitConnection()
	if err != nil {
		log.Println("couldn't init connection to DIMAG sftp server")
		return err
	}
	defer CloseConnection()
	for _, fileRecordObject := range message.FileRecordObjects {
		archivePackageData := ArchivePackageData{
			IOTitle:          fileRecordObject.GetTitle(),
			IOLifetime:       fileRecordObject.GetCombinedLifetime(),
			REPTitle:         "Original",
			PrimaryDocuments: fileRecordObject.GetPrimaryDocuments(),
			CollectionID:     collection.DimagID,
		}
		err = importArchivePackage(message, archivePackageData)
		if err != nil {
			return err
		}
	}
	for _, processRecordObject := range message.ProcessRecordObjects {
		archivePackageData := ArchivePackageData{
			IOTitle:          processRecordObject.GetTitle(),
			IOLifetime:       processRecordObject.GetCombinedLifetime(),
			REPTitle:         "Original",
			PrimaryDocuments: processRecordObject.GetPrimaryDocuments(),
			CollectionID:     collection.DimagID,
		}
		err = importArchivePackage(message, archivePackageData)
		if err != nil {
			return err
		}
	}
	// combine documents which don't belong to a file or process in one archive package
	if len(message.DocumentRecordObjects) > 0 {
		var primaryDocuments []db.PrimaryDocument
		for _, documentRecordObject := range message.DocumentRecordObjects {
			primaryDocuments = append(primaryDocuments, documentRecordObject.GetPrimaryDocuments()...)
		}
		ioTitle := "Nicht zugeordnete Dokumente Behörde: " + process.Agency.Name +
			" Prozess-ID: " + process.ID
		repTitle := "Original"
		archivePackageData := ArchivePackageData{
			IOTitle:          ioTitle,
			IOLifetime:       "-",
			REPTitle:         repTitle,
			PrimaryDocuments: primaryDocuments,
			CollectionID:     collection.DimagID,
		}
		err = importArchivePackage(message, archivePackageData)
		if err != nil {
			return err
		}
	}
	return nil
}

// importArchivePackage archives a file record object in DIMAG.
func importArchivePackage(message db.Message, archivePackageData ArchivePackageData) error {
	importDir, err := uploadFileRecordObjectFiles(sftpClient, message, archivePackageData)
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
	return processImportResponse(response)
}

func processImportResponse(response *http.Response) error {
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("DIMAG ingest error: status code %d", response.StatusCode)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	var parsedResponse EnvelopeImportDocResponse
	err = xml.Unmarshal(body, &parsedResponse)
	if err != nil {
		return err
	}
	if parsedResponse.Body.ImportDocResponse.Status != 200 {
		log.Println(parsedResponse.Body.ImportDocResponse.Message)
		return fmt.Errorf("DIMAG ingest error: %s", parsedResponse.Body.ImportDocResponse.Message)
	}
	return nil
}

// getCollectionIDs gets a list of all collection IDs via a SOAP request from DIMAG.
func GetCollectionIDs() []string {
	requestMetadata := GetAIDforKeyValue{
		Username: DimagApiUser,
		Password: DimagApiPassword,
		Key:      "merkmal",
		Value:    "Bestand",
		Typ:      "Struct",
	}
	soapRequest := EnvelopeGetAIDforKeyValue{
		SoapNs:  SoapNs,
		DimagNs: DimagNs,
		Header:  EnvelopeHeader{},
		Body: EnvelopeBodyGetAIDforKeyValue{
			GetAIDforKeyValue: requestMetadata,
		},
	}
	xmlBytes, err := xml.Marshal(soapRequest)
	if err != nil {
		panic(err)
	}
	requestString := string(xmlBytes)
	req, err := http.NewRequest("POST", DimagApiEndpoint, strings.NewReader(requestString))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "text/xml;charset=UTF-8")
	req.Header.Set("SOAPAction", "getAIDforKeyValue")
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("status code: %d", response.StatusCode))
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	var parsedResponse EnvelopeGetAIDforKeyValueResponse
	err = xml.Unmarshal(body, &parsedResponse)
	if err != nil {
		panic(err)
	}
	if parsedResponse.Body.GetAIDforKeyValueResponse.Status != 200 {
		panic(parsedResponse.Body.GetAIDforKeyValueResponse.Message)
	}
	return strings.Split(parsedResponse.Body.GetAIDforKeyValueResponse.AIDList, ";")
}
