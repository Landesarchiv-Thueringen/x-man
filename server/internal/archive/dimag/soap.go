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
) (jobID int, err error) {
	requestData := importBagData{
		UserName:  DimagApiUser,
		Password:  DimagApiPassword,
		BagItPath: uploadDir,
		Async:     true,
	}
	response, err := soapRequest[importBagResponse](ctx, "importBag", requestData)
	if err != nil {
		return 0, err
	}
	if response.Status != 100 {
		return 0, fmt.Errorf(
			"DIMAG importBag: %d: %s", response.Status, response.Message,
		)
	}
	return response.JobID, nil
}

func getJobStatus(jobID int) (getJobStatusResponse, error) {
	requestData := getJobStatusData{
		UserName: DimagApiUser,
		Password: DimagApiPassword,
		JobID:    jobID,
	}
	return soapRequest[getJobStatusResponse](
		context.Background(), "getJobStatus", requestData,
	)
}

func packageID(r getJobStatusResponse) (string, error) {
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
func GetCollectionIDs() ([]string, error) {
	requestData := getAIDForKeyValueData{
		Username: DimagApiUser,
		Password: DimagApiPassword,
		Key:      "merkmal",
		Value:    "Bestand",
		Typ:      "Struct",
	}
	response, err := soapRequest[getAIDForKeyValueResponse](
		context.Background(), "getAIDforKeyValue", requestData,
	)
	if err != nil {
		return []string{}, err
	}
	if response.Status != 200 {
		return []string{}, fmt.Errorf(
			"DIMAG getAIDforKeyValue: %d: %s",
			response.Status, response.Message,
		)
	}
	return strings.Split(response.AIDList, ";"), nil
}

func soapRequest[R any](ctx context.Context, action string, requestData interface{}) (R, error) {
	// Create request
	envelope := makeEnvelope(requestData)
	xmlBytes, err := xml.MarshalIndent(envelope, " ", " ")
	if err != nil {
		var null R
		return null, err
	}
	requestString := string(xmlBytes)
	req, err := http.NewRequestWithContext(
		ctx, "POST", DimagApiEndpoint,
		strings.NewReader(requestString),
	)
	if err != nil {
		var null R
		return null, err
	}
	req.Header.Set("Content-Type", "text/xml;charset=UTF-8")
	req.Header.Set("SOAPAction", action)
	// Do request
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		var null R
		return null, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		var null R
		return null, fmt.Errorf("DIMAG %s: %s", action, response.Status)
	}
	// Parse response
	body, err := io.ReadAll(response.Body)
	if err != nil {
		var null R
		return null, err
	}
	var responseEnvelope responseEnvelope[R]
	err = xml.Unmarshal(body, &responseEnvelope)
	if err != nil {
		var null R
		return null, err
	}
	return responseEnvelope.Body.Data, nil
}
