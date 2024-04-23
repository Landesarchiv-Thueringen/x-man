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
	"regexp"
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
		archivePackageData := db.ArchivePackage{
			ProcessID:          process.ID,
			IOTitle:            fileRecordObject.GetTitle(),
			IOLifetimeCombined: fileRecordObject.GetCombinedLifetime(),
			REPTitle:           "Original",
			PrimaryDocuments:   fileRecordObject.GetPrimaryDocuments(),
			Collection:         &collection,
			FileRecordObjects:  []db.FileRecordObject{fileRecordObject},
		}
		err = importArchivePackage(message, &archivePackageData)
		if err != nil {
			return err
		}
		db.AddArchivePackage(archivePackageData)
	}
	for _, processRecordObject := range message.ProcessRecordObjects {
		archivePackageData := db.ArchivePackage{
			ProcessID:            process.ID,
			IOTitle:              processRecordObject.GetTitle(),
			IOLifetimeCombined:   processRecordObject.GetCombinedLifetime(),
			REPTitle:             "Original",
			PrimaryDocuments:     processRecordObject.GetPrimaryDocuments(),
			Collection:           &collection,
			ProcessRecordObjects: []db.ProcessRecordObject{processRecordObject},
		}
		err = importArchivePackage(message, &archivePackageData)
		if err != nil {
			return err
		}
		db.AddArchivePackage(archivePackageData)
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
		archivePackageData := db.ArchivePackage{
			ProcessID:             process.ID,
			IOTitle:               ioTitle,
			IOLifetimeCombined:    "-",
			REPTitle:              repTitle,
			PrimaryDocuments:      primaryDocuments,
			Collection:            &collection,
			DocumentRecordObjects: message.DocumentRecordObjects,
		}
		err = importArchivePackage(message, &archivePackageData)
		if err != nil {
			return err
		}
		db.AddArchivePackage(archivePackageData)
	}
	return nil
}

// importArchivePackage archives a file record object in DIMAG.
func importArchivePackage(message db.Message, archivePackageData *db.ArchivePackage) error {
	importDir, err := uploadFileRecordObjectFiles(sftpClient, message, *archivePackageData)
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
	return processImportResponse(response, archivePackageData)
}

func processImportResponse(response *http.Response, archivePackageData *db.ArchivePackage) error {
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
	// Extract package ID from response message.
	re := regexp.MustCompile(`ok: Informationsobjekt (\S+) \[\] : .+ inserted<br\/>`)
	match := re.FindStringSubmatch(parsedResponse.Body.ImportDocResponse.Message)
	if len(match) != 2 {
		return fmt.Errorf("unexpected DIMAG response message: %s", parsedResponse.Body.ImportDocResponse.Message)
	}
	archivePackageData.PackageID = match[1]
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
