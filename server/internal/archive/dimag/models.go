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

type SoapEnvelopeImportDoc struct {
	XMLName xml.Name `xml:"soapenv:Envelope"`
	SoapNs  string   `xml:"xmlns:soapenv,attr"`
	DimagNs string   `xml:"xmlns:web,attr"`
	Header  SoapEnvelopeHeader
	Body    SoapEnvelopeBodyImportDoc
}

type SoapEnvelopeHeader struct {
	XMLName xml.Name `xml:"soapenv:Header"`
}

type SoapEnvelopeBodyImportDoc struct {
	XMLName   xml.Name `xml:"soapenv:Body"`
	ImportDoc SoapImportDoc
}

type SoapImportDoc struct {
	XMLName         xml.Name `xml:"web:importDoc"`
	UserName        string   `xml:"username"`
	Password        string   `xml:"password"`
	ControlFilePath string   `xml:"ControlFile"`
}
