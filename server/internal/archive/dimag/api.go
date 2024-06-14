package dimag

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"lath/xman/internal/db"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var DimagApiEndpoint = os.Getenv("DIMAG_CORE_SOAP_ENDPOINT")
var DimagApiUser = os.Getenv("DIMAG_CORE_USER")
var DimagApiPassword = os.Getenv("DIMAG_CORE_PASSWORD")

// ImportArchivePackage archives a file record object in DIMAG.
func ImportArchivePackage(
	ctx context.Context,
	process db.SubmissionProcess,
	message db.Message,
	aip *db.ArchivePackage,
	c Connection,
) error {
	importDir, err := uploadArchivePackage(ctx, c, process, message, *aip)
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
		panic(err)
	}
	requestString := string(xmlBytes)
	req, err := http.NewRequestWithContext(ctx, "POST", DimagApiEndpoint, strings.NewReader(requestString))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/xml;charset=UTF-8")
	req.Header.Set("SOAPAction", "importDoc")
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	processImportResponse(response, aip)
	return nil
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
