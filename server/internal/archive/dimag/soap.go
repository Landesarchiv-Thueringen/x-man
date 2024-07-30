package dimag

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

var DimagApiEndpoint = os.Getenv("DIMAG_CORE_SOAP_ENDPOINT")
var DimagApiUser = os.Getenv("DIMAG_CORE_USER")
var DimagApiPassword = os.Getenv("DIMAG_CORE_PASSWORD")

func importBag(
	ctx context.Context,
	uploadDir string,
) (importBagResponse, error) {
	// Create request
	requestData := importBagData{
		UserName:  DimagApiUser,
		Password:  DimagApiPassword,
		BagItPath: uploadDir,
	}
	envelope := makeImportBagEnvelope(requestData)
	xmlBytes, err := xml.MarshalIndent(envelope, " ", " ")
	if err != nil {
		panic(err)
	}
	requestString := string(xmlBytes)
	req, err := http.NewRequestWithContext(
		ctx, "POST", DimagApiEndpoint,
		strings.NewReader(requestString),
	)
	if err != nil {
		return importBagResponse{}, err
	}
	req.Header.Set("Content-Type", "text/xml;charset=UTF-8")
	req.Header.Set("SOAPAction", "importBag")
	// Do request
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return importBagResponse{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return importBagResponse{}, fmt.Errorf("DIMAG importBag: %s", response.Status)
	}
	// Parse response
	var parsedResponse importBagResponseEnvelope
	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	err = xml.Unmarshal(body, &parsedResponse)
	if err != nil {
		panic(err)
	}
	responseData := parsedResponse.Body.ImportBagResponse
	if responseData.Status != 200 {
		return importBagResponse{}, fmt.Errorf(
			"DIMAG importBag: %d: %s", responseData.Status, responseData.Message,
		)
	}
	return responseData, nil
}

func packageID(r importBagResponse) (string, error) {
	types := strings.Split(r.Types, ";")
	aids := strings.Split(r.AIDs, ";")
	for i, t := range types {
		if itemType(t) == itemTypeInformationObject {
			return aids[i], nil
		}
	}
	return "", errors.New("failed to find AID for information object in import-bag response")
}

// GetCollectionIDs gets a list of all collection IDs via a SOAP request from DIMAG.
func GetCollectionIDs() []string {
	requestData := getAIDForKeyValueData{
		Username: DimagApiUser,
		Password: DimagApiPassword,
		Key:      "merkmal",
		Value:    "Bestand",
		Typ:      "Struct",
	}
	soapRequest := makeGetAIDForKeyValueEnvelope(requestData)
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
	var parsedResponse getAIDForKeyValueResponseEnvelope
	err = xml.Unmarshal(body, &parsedResponse)
	if err != nil {
		panic(err)
	}
	if parsedResponse.Body.GetAIDforKeyValueResponse.Status != 200 {
		panic(parsedResponse.Body.GetAIDforKeyValueResponse.Message)
	}
	return strings.Split(parsedResponse.Body.GetAIDforKeyValueResponse.AIDList, ";")
}
