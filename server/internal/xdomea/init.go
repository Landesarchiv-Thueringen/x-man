package xdomea

import (
	"fmt"
	"lath/xman/internal/db"
	"os"
)

func InitCodeLists() {
	initMessageTypes()
	initXdomeaVersions()
	initRecordObjectAppraisals()
	initConfidentialityLevelCodelist()
	initMediumCodelist()
}

func InitTestSetup() {
	initAgencies()
}

func initMessageTypes() {
	messageTypes := []*db.MessageType{
		{Code: "0000"}, // unknown message type
		{Code: "0501"},
		{Code: "0502"},
		{Code: "0503"},
		{Code: "0504"},
		{Code: "0505"},
		{Code: "0506"},
		{Code: "0507"},
	}
	db.InitMessageTypes(messageTypes)
}

func initXdomeaVersions() {
	versions := []*db.XdomeaVersion{
		{
			Code:    "2.3.0",
			URI:     "urn:xoev-de:xdomea:schema:2.3.0",
			XSDPath: "xsd/2.3.0/xdomea-Nachrichten-AussonderungDurchfuehren.xsd",
		},
		{
			Code:    "2.4.0",
			URI:     "urn:xoev-de:xdomea:schema:2.4.0",
			XSDPath: "xsd/2.4.0/xdomea-Nachrichten-AussonderungDurchfuehren.xsd",
		},
		{
			Code:    "3.0.0",
			URI:     "urn:xoev-de:xdomea:schema:3.0.0",
			XSDPath: "xsd/3.0.0/xdomea-Nachrichten-AussonderungDurchfuehren.xsd",
		},
		{
			Code:    "3.1.0",
			URI:     "urn:xoev-de:xdomea:schema:3.1.0",
			XSDPath: "xsd/3.1.0/xdomea-Nachrichten-AussonderungDurchfuehren.xsd",
		},
	}
	db.InitXdomeaVersions(versions)
}

func initRecordObjectAppraisals() {
	appraisals := []*db.RecordObjectAppraisal{
		{Code: "A", ShortDesc: "Archivieren", Desc: "Das Schriftgutobjekt ist archivwürdig."},
		{Code: "B", ShortDesc: "Durchsicht", Desc: "Das Schriftgutobjekt ist zum Bewerten markiert."},
		{Code: "V", ShortDesc: "Vernichten", Desc: "Das Schriftgutobjekt ist zum Vernichten markiert."},
	}
	db.InitRecordObjectAppraisals(appraisals)
}

func initConfidentialityLevelCodelist() {
	confidentialityLevelCodelist := []*db.ConfidentialityLevel{
		{ID: "001", ShortDesc: "Geheim", Desc: "Geheim: Das Schriftgutobjekt ist als geheim eingestuft."},
		{ID: "002", ShortDesc: "NfD", Desc: "NfD: Das Schriftgutobjekt ist als \"nur für den Dienstgebrauch (nfD)\" eingestuft."},
		{ID: "003", ShortDesc: "Offen", Desc: "Offen: Das Schriftgutobjekt ist nicht eingestuft."},
		{ID: "004", ShortDesc: "Streng geheim", Desc: "Streng geheim: Das Schriftgutobjekt ist als streng geheim eingestuft."},
		{ID: "005", ShortDesc: "Vertraulich", Desc: "Vertraulich: Das Schriftgutobjekt ist als vertraulich eingestuft."},
	}
	db.InitConfidentialityLevelCodelist(confidentialityLevelCodelist)
}

func initMediumCodelist() {
	mediumCodelist := []*db.Medium{
		{ID: "001", ShortDesc: "Elektronisch", Desc: "Elektronisch: Das Schriftgutobjekt liegt ausschließlich in elektronischer Form vor."},
		{ID: "002", ShortDesc: "Hybrid", Desc: "Hybrid: Das Schriftgutobjekt liegt teilweise in elektronischer Form und teilweise als Papier vor."},
		{ID: "003", ShortDesc: "Papier", Desc: "Papier: Das Schriftgutobjekt liegt ausschließlich als Papier vor."},
	}
	db.InitMediumCodelist(mediumCodelist)
}

func initAgencies() {
	agencies := []db.Agency{
		{
			Name:           "Thüringer Ministerium für Inneres und Kommunales",
			Abbreviation:   "TMIK",
			TransferDirURL: "file:///xman/transfer_dir",
		},
		{
			Name:           "Thüringer Staatskanzlei",
			Abbreviation:   "TSK",
			TransferDirURL: fmt.Sprintf("dav://%s:%s@webdav/", os.Getenv("WEBDAV_USERNAME"), os.Getenv("WEBDAV_PASSWORD")),
		},
	}
	db.InitAgencies(agencies)
}
