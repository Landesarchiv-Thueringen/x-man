package dimag

import (
	"encoding/xml"
)

const SoapNs string = "http://schemas.xmlsoap.org/soap/envelope/"
const DimagNs string = "http://dimag.la-bw.de/WebService.wsdl"

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

type EnvelopeImportDoc struct {
	XMLName xml.Name `xml:"soapenv:Envelope"`
	SoapNs  string   `xml:"xmlns:soapenv,attr"`
	DimagNs string   `xml:"xmlns:web,attr"`
	Header  EnvelopeHeader
	Body    EnvelopeBodyImportDoc
}

type EnvelopeHeader struct {
	XMLName xml.Name `xml:"soapenv:Header"`
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
