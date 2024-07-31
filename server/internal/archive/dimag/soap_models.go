package dimag

import (
	"encoding/xml"
)

// General

type envelope[T any] struct {
	XMLName xml.Name `xml:"soapenv:Envelope"`
	SoapNs  string   `xml:"xmlns:soapenv,attr"`
	DimagNs string   `xml:"xmlns:web,attr"`
	Header  struct {
		XMLName xml.Name `xml:"soapenv:Header"`
	}
	Body struct {
		XMLName xml.Name `xml:"soapenv:Body"`
		Data    T
	}
}

func makeEnvelope[T any](data T) envelope[T] {
	envelope := envelope[T]{
		SoapNs:  "http://schemas.xmlsoap.org/soap/envelope/",
		DimagNs: "http://dimag.la-bw.de/WebService.wsdl",
	}
	envelope.Body.Data = data
	return envelope
}

type responseEnvelope[T any] struct {
	XMLName xml.Name `xml:"Envelope"`
	SoapNs  string   `xml:"xmlns:SOAP-ENV,attr"`
	DimagNs string   `xml:"xmlns:ns1,attr"`
	Body    struct {
		XMLName xml.Name `xml:"Body"`
		Data    T
	}
}

// ImportBag

type importBagData struct {
	XMLName   xml.Name `xml:"web:importBag"`
	UserName  string   `xml:"username"`
	Password  string   `xml:"password"`
	BagItPath string   `xml:"bagItPath"`
	Async     bool     `xml:"async"`
}

type importBagResponse struct {
	XMLName xml.Name `xml:"importBagResponse"`
	Status  int      `xml:"status"`
	Message string   `xml:"msg"`
	JobID   int      `xml:"jid"`
	AIDs    string   `xml:"aids"`
	Types   string   `xml:"types"`
}

// GetJobStatus

type getJobStatusData struct {
	XMLName  xml.Name `xml:"web:getJobStatus"`
	UserName string   `xml:"username"`
	Password string   `xml:"password"`
	JobID    int      `xml:"jid"`
}

type getJobStatusResponse struct {
	XMLName xml.Name `xml:"getJobStatusResponse"`
	Status  int      `xml:"status"`
	JobID   int      `xml:"jid"`
	Message string   `xml:"msg"`
	AIDs    string   `xml:"aids"`
	Types   string   `xml:"types"`
}

// GetAIDforKeyValue

type getAIDForKeyValueData struct {
	XMLName  xml.Name `xml:"web:getAIDforKeyValue"`
	Username string   `xml:"username"`
	Password string   `xml:"password"`
	Key      string   `xml:"key"`
	Value    string   `xml:"value"`
	Typ      string   `xml:"typ"`
}

type getAIDForKeyValueResponse struct {
	XMLName xml.Name `xml:"getAIDforKeyValueResponse"`
	Status  uint     `xml:"status"`
	Message string   `xml:"msg"`
	Hits    uint     `xml:"hits"`
	AIDList string   `xml:"aIDList"`
}
