package xdomea

import (
	"fmt"
	"lath/xman/internal/db"
	"os"
)

type XdomeaVersion struct {
	Code    string
	URI     string
	XSDPath string
}

var XdomeaVersions = map[string]XdomeaVersion{
	"2.3.0": {
		Code:    "2.3.0",
		URI:     "urn:xoev-de:xdomea:schema:2.3.0",
		XSDPath: "xsd/2.3.0/xdomea-Nachrichten-AussonderungDurchfuehren.xsd",
	},
	"2.4.0": {
		Code:    "2.4.0",
		URI:     "urn:xoev-de:xdomea:schema:2.4.0",
		XSDPath: "xsd/2.4.0/xdomea-Nachrichten-AussonderungDurchfuehren.xsd",
	},
	"3.0.0": {
		Code:    "3.0.0",
		URI:     "urn:xoev-de:xdomea:schema:3.0.0",
		XSDPath: "xsd/3.0.0/xdomea-Nachrichten-AussonderungDurchfuehren.xsd",
	},
	"3.1.0": {
		Code:    "3.1.0",
		URI:     "urn:xoev-de:xdomea:schema:3.1.0",
		XSDPath: "xsd/3.1.0/xdomea-Nachrichten-AussonderungDurchfuehren.xsd",
	},
}

func InitTestSetup() {
	initAgencies()
}

func initAgencies() {
	db.InsertAgency(db.Agency{
		Name:           "Thüringer Ministerium für Inneres und Kommunales",
		Abbreviation:   "TMIK",
		TransferDirURL: "file:///xman/transfer_dir",
	})
	db.InsertAgency(db.Agency{
		Name:           "Thüringer Staatskanzlei",
		Abbreviation:   "TSK",
		TransferDirURL: fmt.Sprintf("dav://%s:%s@webdav/", os.Getenv("WEBDAV_USERNAME"), os.Getenv("WEBDAV_PASSWORD")),
	})
}
