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
