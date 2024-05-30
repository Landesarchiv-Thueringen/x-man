package dimag

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"lath/xman/internal/archive"
	"lath/xman/internal/db"
	"lath/xman/internal/xdomea"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var DimagApiEndpoint = os.Getenv("DIMAG_CORE_SOAP_ENDPOINT")
var DimagApiUser = os.Getenv("DIMAG_CORE_USER")
var DimagApiPassword = os.Getenv("DIMAG_CORE_PASSWORD")

// ImportMessageSync archives all metadata and files of a 0503 message in DIMAG.
//
// The record objects in the message should be complete loaded.
//
// ImportMessageSync returns after the archiving process completed.
func ImportMessageSync(process db.SubmissionProcess, message db.Message, collection db.ArchiveCollection) {
	InitConnection()
	defer CloseConnection()
	rootRecords := db.FindRootRecords(context.Background(), process.ProcessID, message.MessageType)
	for _, f := range rootRecords.Files {
		aip := db.ArchivePackage{
			ProcessID:        process.ProcessID,
			IOTitle:          archive.GetFileRecordTitle(f),
			IOLifetime:       f.Lifetime,
			REPTitle:         "Original",
			PrimaryDocuments: xdomea.GetPrimaryDocumentsForFile(&f),
			CollectionID:     collection.ID,
			RootRecordIDs:    []uuid.UUID{f.RecordID},
		}
		importArchivePackage(process, message, &aip)
		db.InsertArchivePackage(aip)
	}
	for _, p := range rootRecords.Processes {
		aip := db.ArchivePackage{
			ProcessID:        process.ProcessID,
			IOTitle:          archive.GetProcessRecordTitle(p),
			IOLifetime:       p.Lifetime,
			REPTitle:         "Original",
			PrimaryDocuments: xdomea.GetPrimaryDocumentsForProcess(&p),
			CollectionID:     collection.ID,
			RootRecordIDs:    []uuid.UUID{p.RecordID},
		}
		importArchivePackage(process, message, &aip)
		db.InsertArchivePackage(aip)
	}
	// Combine documents which don't belong to a file or process in one archive package.
	if len(rootRecords.Documents) > 0 {
		var primaryDocuments []db.PrimaryDocument
		for _, d := range rootRecords.Documents {
			primaryDocuments = append(primaryDocuments, xdomea.GetPrimaryDocumentsForDocument(&d)...)
		}
		ioTitle := "Nicht zugeordnete Dokumente Behörde: " + process.Agency.Name +
			" Prozess-ID: " + process.ProcessID.String()
		repTitle := "Original"
		var rootRecordIDs []uuid.UUID
		for _, r := range rootRecords.Documents {
			rootRecordIDs = append(rootRecordIDs, r.RecordID)
		}
		aip := db.ArchivePackage{
			ProcessID:        process.ProcessID,
			IOTitle:          ioTitle,
			IOLifetime:       nil,
			REPTitle:         repTitle,
			PrimaryDocuments: primaryDocuments,
			CollectionID:     collection.ID,
			RootRecordIDs:    rootRecordIDs,
		}
		importArchivePackage(process, message, &aip)
		db.InsertArchivePackage(aip)
	}
}

// importArchivePackage archives a file record object in DIMAG.
func importArchivePackage(
	process db.SubmissionProcess,
	message db.Message,
	aip *db.ArchivePackage,
) {
	importDir := uploadArchivePackage(sftpClient, process, message, *aip)
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
		panic(err)
	}
	requestString := string(xmlBytes)
	req, err := http.NewRequest("POST", DimagApiEndpoint, strings.NewReader(requestString))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "text/xml;charset=UTF-8")
	req.Header.Set("SOAPAction", "importDoc")
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	processImportResponse(response, aip)
}

func processImportResponse(response *http.Response, archivePackageData *db.ArchivePackage) {
	if response.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("DIMAG ingest error: status code %d", response.StatusCode))
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	var parsedResponse EnvelopeImportDocResponse
	err = xml.Unmarshal(body, &parsedResponse)
	if err != nil {
		panic(err)
	}
	if parsedResponse.Body.ImportDocResponse.Status != 200 {
		panic(fmt.Sprintf("DIMAG ingest error: %s", parsedResponse.Body.ImportDocResponse.Message))
	}
	// Extract package ID from response message.
	re := regexp.MustCompile(`ok: Informationsobjekt (\S+) \[\] : .+ inserted<br\/>`)
	match := re.FindStringSubmatch(parsedResponse.Body.ImportDocResponse.Message)
	if len(match) != 2 {
		panic(fmt.Sprintf("unexpected DIMAG response message: %s", parsedResponse.Body.ImportDocResponse.Message))
	}
	archivePackageData.PackageID = match[1]
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
