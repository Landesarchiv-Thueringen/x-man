package dimag

import (
	"encoding/xml"
)

type DimagControl struct {
	XMLName    xml.Name `xml:"verzeichnungseinheit"`
	RootID     string   `xml:"rootid"`
	IndexItems []IndexItem
}

type IndexItem struct {
	XMLName    xml.Name `xml:"verz-obj"`
	IndexID    string   `xml:"aid"`
	ItemType   string   `xml:"typ"`
	Title      string   `xml:"titel"`
	FilePath   string   `xml:"sftp-dateiname"`
	IndexItems []IndexItem
}

type SoapEnvelopeImportDoc struct {
	XMLName xml.Name `xml:"soapenv:Envelope"`
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
