package dimag

import (
	"encoding/xml"
)

const soapNs string = "http://schemas.xmlsoap.org/soap/envelope/"
const dimagNs string = "http://dimag.la-bw.de/WebService.wsdl"

// ImportBag

type importBagData struct {
	XMLName   xml.Name `xml:"web:importBag"`
	UserName  string   `xml:"username"`
	Password  string   `xml:"password"`
	BagItPath string   `xml:"bagItPath"`
}

type importBagEnvelope struct {
	XMLName xml.Name `xml:"soapenv:Envelope"`
	SoapNs  string   `xml:"xmlns:soapenv,attr"`
	DimagNs string   `xml:"xmlns:web,attr"`
	Header  struct {
		XMLName xml.Name `xml:"soapenv:Header"`
	}
	Body struct {
		XMLName   xml.Name `xml:"soapenv:Body"`
		ImportBag importBagData
	}
}

func makeImportBagEnvelope(data importBagData) importBagEnvelope {
	envelope := importBagEnvelope{
		SoapNs:  soapNs,
		DimagNs: dimagNs,
	}
	envelope.Body.ImportBag = data
	return envelope
}

type importBagResponse struct {
	XMLName xml.Name `xml:"importBagResponse"`
	Status  int      `xml:"status"`
	Message string   `xml:"msg"`
	JobID   int      `xml:"jid"`
	AIDs    string   `xml:"aids"`
	Types   string   `xml:"types"`
}

type importBagResponseEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	SoapNs  string   `xml:"xmlns:SOAP-ENV,attr"`
	DimagNs string   `xml:"xmlns:ns1,attr"`
	Body    struct {
		XMLName           xml.Name `xml:"Body"`
		ImportBagResponse importBagResponse
	}
}

// GetAIDforKeyValue

type getAIDForKeyValueEnvelope struct {
	XMLName xml.Name `xml:"soapenv:Envelope"`
	SoapNs  string   `xml:"xmlns:soapenv,attr"`
	DimagNs string   `xml:"xmlns:web,attr"`
	Header  struct {
		XMLName xml.Name `xml:"soapenv:Header"`
	}
	Body struct {
		XMLName           xml.Name `xml:"soapenv:Body"`
		GetAIDforKeyValue getAIDForKeyValueData
	}
}

func makeGetAIDForKeyValueEnvelope(data getAIDForKeyValueData) getAIDForKeyValueEnvelope {
	envelope := getAIDForKeyValueEnvelope{
		SoapNs:  soapNs,
		DimagNs: dimagNs,
	}
	envelope.Body.GetAIDforKeyValue = data
	return envelope
}

type getAIDForKeyValueData struct {
	XMLName  xml.Name `xml:"web:getAIDforKeyValue"`
	Username string   `xml:"username"`
	Password string   `xml:"password"`
	Key      string   `xml:"key"`
	Value    string   `xml:"value"`
	Typ      string   `xml:"typ"`
}

type getAIDForKeyValueResponseEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	SoapNs  string   `xml:"xmlns:SOAP-ENV,attr"`
	DimagNs string   `xml:"xmlns:ns1,attr"`
	Body    struct {
		XMLName                   xml.Name `xml:"Body"`
		GetAIDforKeyValueResponse getAIDForKeyValueResponse
	}
}

type getAIDForKeyValueResponse struct {
	XMLName xml.Name `xml:"getAIDforKeyValueResponse"`
	Status  uint     `xml:"status"`
	Message string   `xml:"msg"`
	Hits    uint     `xml:"hits"`
	AIDList string   `xml:"aIDList"`
}
