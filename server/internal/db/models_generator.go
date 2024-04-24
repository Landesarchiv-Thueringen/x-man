package db

import (
	"encoding/xml"

	"github.com/google/uuid"
)

type GeneratorMessage0502 struct {
	XMLName          xml.Name                   `xml:"xdomea:Aussonderung.Bewertungsverzeichnis.0502"`
	MessageHead      GeneratorMessageHead0502   `xml:"xdomea:Kopf"`
	AppraisedObjects []GeneratorAppraisedObject `xml:"xdomea:BewertetesObjekt"`
	XdomeaXmlNs      string                     `xml:"xmlns:xdomea,attr"`
	XsiXmlNs         string                     `xml:"xmlns:xsi,attr"`
}

type GeneratorMessage0504 struct {
	XMLName     xml.Name             `gorm:"-" xml:"xdomea:Aussonderung.AnbietungEmpfangBestaetigen.0504" json:"-"`
	MessageHead GeneratorMessageHead `xml:"xdomea:Kopf" json:"messageHead"`
	XdomeaXmlNs string               `xml:"xmlns:xdomea,attr"`
	XsiXmlNs    string               `xml:"xmlns:xsi,attr"`
}

type GeneratorMessage0506 struct {
	XMLName             xml.Name                      `gorm:"-" xml:"xdomea:Aussonderung.AussonderungImportBestaetigen.0506" json:"-"`
	MessageHead         GeneratorMessageHead          `xml:"xdomea:Kopf" json:"messageHead"`
	XdomeaXmlNs         string                        `xml:"xmlns:xdomea,attr"`
	XsiXmlNs            string                        `xml:"xmlns:xsi,attr"`
	ArchivingInfoPre300 *GeneratorArchivingInfoPre300 `xml:"xdomea:ErfolgOderMisserfolg"`
	ArchivedRecordInfo  []GeneratorArchivedRecordInfo `xml:"xdomea:AusgesondertesSGO"`
}

type GeneratorAppraisedObject struct {
	XMLName         xml.Name                 `xml:"xdomea:BewertetesObjekt"`
	XdomeaID        uuid.UUID                `xml:"xdomea:ID"`
	ObjectAppraisal GeneratorObjectAppraisal `xml:"xdomea:Aussonderungsart"`
}

type GeneratorObjectAppraisal struct {
	XMLName             xml.Name       `xml:"xdomea:Aussonderungsart"`
	AppraisalCode       *GeneratorCode `xml:"xdomea:Aussonderungsart"`
	AppraisalCodePre300 *string        `xml:"code"`
}

type GeneratorMessageHead0502 struct {
	ProcessID        string                 `xml:"xdomea:ProzessID"`
	MessageType      GeneratorCode          `xml:"xdomea:Nachrichtentyp"`
	CreationTime     string                 `xml:"xdomea:Erstellungszeitpunkt"`
	Sender           GeneratorContact       `xml:"xdomea:Absender"`
	Receiver         GeneratorContact       `xml:"xdomea:Empfaenger"`
	SendingSystem    GeneratorSendingSystem `xml:"xdomea:SendendesSystem"`
	ReceiptRequested bool                   `xml:"xdomea:Empfangsbestaetigung"`
}

type GeneratorMessageHead struct {
	ProcessID     string                 `xml:"xdomea:ProzessID"`
	MessageType   GeneratorCode          `xml:"xdomea:Nachrichtentyp"`
	CreationTime  string                 `xml:"xdomea:Erstellungszeitpunkt"`
	Sender        GeneratorContact       `xml:"xdomea:Absender"`
	Receiver      GeneratorContact       `xml:"xdomea:Empfaenger"`
	SendingSystem GeneratorSendingSystem `xml:"xdomea:SendendesSystem"`
}

type GeneratorSendingSystem struct {
	XMLName        xml.Name `xml:"xdomea:SendendesSystem"`
	ProductName    *string  `xml:"xdomea:Produktname"`
	ProductVersion *string  `xml:"xdomea:Version"`
}

type GeneratorContact struct {
	AgencyIdentification *GeneratorAgencyIdentification `xml:"xdomea:Behoerdenkennung"`
	Institution          *GeneratorInstitution          `xml:"xdomea:Institution"`
}

type GeneratorAgencyIdentification struct {
	Code   *GeneratorCode `xml:"xdomea:Behoerdenschluessel"`
	Prefix *GeneratorCode `xml:"xdomea:Praefix"`
}

type GeneratorInstitution struct {
	Name         *string `xml:"xdomea:Name"`
	Abbreviation *string `xml:"xdomea:Kurzbezeichnung"`
}

type GeneratorCode struct {
	Code string `xml:"code"`
}

type GeneratorArchivingInfoPre300 struct {
	Success              bool                            `xml:"xdomea:Erfolgreich"`
	RecordArchiveMapping []GeneratorRecordArchiveMapping `xml:"xdomea:Rueckgabeparameter"`
}

type GeneratorRecordArchiveMapping struct {
	RecordID  string `xml:"xdomea:ID"`
	ArchiveID string `xml:"xdomea:Archivkennung"`
}

type GeneratorArchivedRecordInfo struct {
	RecordID  string  `xml:"xdomea:IDSGO"`
	Success   bool    `xml:"xdomea:Erfolgreich"`
	ArchiveID *string `xml:"xdomea:Archivkennung"`
}
