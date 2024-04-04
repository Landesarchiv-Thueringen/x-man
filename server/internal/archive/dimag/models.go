package dimag

import (
	"encoding/xml"
	"lath/xman/internal/db"
)

const SoapNs string = "http://schemas.xmlsoap.org/soap/envelope/"
const DimagNs string = "http://dimag.la-bw.de/WebService.wsdl"

type ArchivePackageData struct {
	IOTitle          string
	IOLifetime       string
	REPTitle         string
	PrimaryDocuments []db.PrimaryDocument
	CollectionID     string
}

type DimagControl struct {
	XMLName    xml.Name `xml:"verzeichnungseinheit"`
	RootID     string   `xml:"rootid"`
	IndexItems []IndexItem
}

type IndexItem struct {
	XMLName    xml.Name `xml:"verz-obj"`
	IndexID    string   `xml:"aid"`
	Lifetime   string   `xml:"entstehungs-zeitraum"`
	FilePath   string   `xml:"sftp-dateiname,omitempty"`
	Title      string   `xml:"titel"`
	ItemType   string   `xml:"typ"`
	IndexItems []IndexItem
}

type EnvelopeHeader struct {
	XMLName xml.Name `xml:"soapenv:Header"`
}

// ImportDoc

type EnvelopeImportDoc struct {
	XMLName xml.Name `xml:"soapenv:Envelope"`
	SoapNs  string   `xml:"xmlns:soapenv,attr"`
	DimagNs string   `xml:"xmlns:web,attr"`
	Header  EnvelopeHeader
	Body    EnvelopeBodyImportDoc
}

type EnvelopeBodyImportDoc struct {
	XMLName   xml.Name `xml:"soapenv:Body"`
	ImportDoc ImportDoc
}

type ImportDoc struct {
	XMLName         xml.Name `xml:"web:importDoc"`
	UserName        string   `xml:"username"`
	Password        string   `xml:"password"`
	ControlFilePath string   `xml:"ControlFile"`
}

type EnvelopeImportDocResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	SoapNs  string   `xml:"xmlns:SOAP-ENV,attr"`
	DimagNs string   `xml:"xmlns:ns1,attr"`
	Body    ImportDocResponseBody
}

type ImportDocResponseBody struct {
	XMLName           xml.Name `xml:"Body"`
	ImportDocResponse ImportDocResponse
}

type ImportDocResponse struct {
	XMLName xml.Name `xml:"importDocResponse"`
	Status  uint     `xml:"status"`
	Message string   `xml:"msg"`
}

// GetAIDforKeyValue

type EnvelopeGetAIDforKeyValue struct {
	XMLName xml.Name `xml:"soapenv:Envelope"`
	SoapNs  string   `xml:"xmlns:soapenv,attr"`
	DimagNs string   `xml:"xmlns:web,attr"`
	Header  EnvelopeHeader
	Body    EnvelopeBodyGetAIDforKeyValue
}

type EnvelopeBodyGetAIDforKeyValue struct {
	XMLName           xml.Name `xml:"soapenv:Body"`
	GetAIDforKeyValue GetAIDforKeyValue
}

type GetAIDforKeyValue struct {
	XMLName  xml.Name `xml:"web:getAIDforKeyValue"`
	Username string   `xml:"username"`
	Password string   `xml:"password"`
	Key      string   `xml:"key"`
	Value    string   `xml:"value"`
	Typ      string   `xml:"typ"`
}

type EnvelopeGetAIDforKeyValueResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	SoapNs  string   `xml:"xmlns:SOAP-ENV,attr"`
	DimagNs string   `xml:"xmlns:ns1,attr"`
	Body    GetAIDforKeyValueResponseBody
}

type GetAIDforKeyValueResponseBody struct {
	XMLName                   xml.Name `xml:"Body"`
	GetAIDforKeyValueResponse GetAIDforKeyValueResponse
}

type GetAIDforKeyValueResponse struct {
	XMLName xml.Name `xml:"getAIDforKeyValueResponse"`
	Status  uint     `xml:"status"`
	Message string   `xml:"msg"`
	Hits    uint     `xml:"hits"`
	AIDList string   `xml:"aIDList"`
}
