package db

import (
	"encoding/xml"

	"github.com/google/uuid"
)

type GeneratorMessage0502 struct {
	XMLName          xml.Name                   `xml:"xdomea:Aussonderung.Bewertungsverzeichnis.0502"`
	MessageHead      GeneratorMessageHead       `xml:"xdomea:Kopf"`
	AppraisedObjects []GeneratorAppraisedObject `xml:"xdomea:BewertetesObjekt"`
	XdomeaXmlNs      string                     `xml:"xmlns:xdomea,attr"`
}

type GeneratorAppraisedObject struct {
	XMLName         xml.Name                 `xml:"xdomea:BewertetesObjekt"`
	XdomeaID        uuid.UUID                `xml:"xdomea:ID"`
	ObjectAppraisal GeneratorObjectAppraisal `xml:"xdomea:Aussonderungsart"`
}

type GeneratorObjectAppraisal struct {
	XMLName       xml.Name               `xml:"xdomea:Aussonderungsart"`
	AppraisalCode GeneratorAppraisalCode `xml:"xdomea:Aussonderungsart"`
}

type GeneratorAppraisalCode struct {
	Code string `xml:"code"`
}

type GeneratorMessageHead struct {
	ProcessID    string           `xml:"xdomea:ProzessID"`
	CreationTime string           `xml:"xdomea:Erstellungszeitpunkt"`
	Sender       GeneratorContact `xml:"xdomea:Absender"`
	Receiver     GeneratorContact `xml:"xdomea:Empfaenger"`
}

type GeneratorContact struct {
	AgencyIdentification *GeneratorAgencyIdentification `xml:"xdomea:Behoerdenkennung"`
	Institution          *GeneratorInstitution          `xml:"xdomea:Institution"`
}

type GeneratorAgencyIdentification struct {
	Code   *GeneratorCode `xml:"xdomea:Behoerdenschluessel"`
	Prefix *GeneratorCode `xml:"xdomea:Praefix"`
}

type GeneratorCode struct {
	Code *string `xml:"code"`
	Name *string `xml:"name"`
}

type GeneratorInstitution struct {
	Name         *string `xml:"xdomea:Name"`
	Abbreviation *string `xml:"xdomea:Kurzbezeichnung"`
}
