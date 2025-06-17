package core

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"lath/xman/internal/db"
	"os"
)

type message0501 struct {
	XMLName         xml.Name            `xml:"Aussonderung.Anbieteverzeichnis.0501"`
	MessageHead     db.MessageHead      `xml:"Kopf"`
	FileRecords     []db.FileRecord     `xml:"Schriftgutobjekt>Akte"`
	ProcessRecords  []db.ProcessRecord  `xml:"Schriftgutobjekt>Vorgang"`
	DocumentRecords []db.DocumentRecord `xml:"Schriftgutobjekt>Dokument"`
}

type messageBody0501 struct {
	XMLName xml.Name `xml:"Aussonderung.Anbieteverzeichnis.0501"`
}

type message0503 struct {
	XMLName         xml.Name            `xml:"Aussonderung.Aussonderung.0503"`
	MessageHead     db.MessageHead      `xml:"Kopf"`
	FileRecords     []db.FileRecord     `xml:"Schriftgutobjekt>Akte"`
	ProcessRecords  []db.ProcessRecord  `xml:"Schriftgutobjekt>Vorgang"`
	DocumentRecords []db.DocumentRecord `xml:"Schriftgutobjekt>Dokument"`
}

type messageBody0503 struct {
	XMLName xml.Name `xml:"Aussonderung.Aussonderung.0503"`
}

type message0505 struct {
	XMLName     xml.Name       `xml:"Aussonderung.BewertungEmpfangBestaetigen.0505"`
	MessageHead db.MessageHead `xml:"Kopf"`
}

type messageBody0505 struct {
	XMLName xml.Name `xml:"Aussonderung.BewertungEmpfangBestaetigen.0505"`
}

type parsedMessage struct {
	MessageHead   db.MessageHead
	XdomeaVersion XdomeaVersion
	RootRecords   *db.RootRecords
}

func extractXdomeaVersion(messageType db.MessageType, messagePath string) (XdomeaVersion, error) {
	xmlFile, err := os.Open(messagePath)
	if err != nil {
		panic(err)
	}
	defer xmlFile.Close()
	// Read the opened xmlFile as a byte array.
	xmlBytes, err := io.ReadAll(xmlFile)
	if err != nil {
		panic(err)
	}
	switch messageType {
	case "0501":
		var messageBody messageBody0501
		err = xml.Unmarshal(xmlBytes, &messageBody)
		if err != nil {
			return XdomeaVersion{}, err
		}
		return extractVersion(messageBody.XMLName.Space)
	case "0503":
		var messageBody messageBody0503
		err = xml.Unmarshal(xmlBytes, &messageBody)
		if err != nil {
			return XdomeaVersion{}, err
		}
		return extractVersion(messageBody.XMLName.Space)
	case "0505":
		var messageBody messageBody0505
		err = xml.Unmarshal(xmlBytes, &messageBody)
		if err != nil {
			return XdomeaVersion{}, err
		}
		return extractVersion(messageBody.XMLName.Space)
	default:
		panic("unknown message type: " + messageType)
	}
}

func parseMessage(messagePath string, messageType db.MessageType) (result parsedMessage, err error) {
	xmlFile, err := os.Open(messagePath)
	if err != nil {
		panic(err)
	}
	defer xmlFile.Close()
	// Read the opened xmlFile as a byte array.
	xmlBytes, err := io.ReadAll(xmlFile)
	if err != nil {
		panic(err)
	}
	switch messageType {
	case "0501":
		return parse0501Message(xmlBytes)
	case "0503":
		return parse0503Message(xmlBytes)
	case "0505":
		return parse0505Message(xmlBytes)
	default:
		panic("unknown message type: " + messageType)
	}
}

func parse0501Message(xmlBytes []byte) (result parsedMessage, err error) {
	var m message0501
	err = xml.Unmarshal(xmlBytes, &m)
	if err != nil {
		return
	}
	result.MessageHead = m.MessageHead
	result.RootRecords = &db.RootRecords{
		Files:     m.FileRecords,
		Processes: m.ProcessRecords,
		Documents: m.DocumentRecords,
	}
	result.XdomeaVersion, err = extractVersion(m.XMLName.Space)
	return
}

func parse0503Message(xmlBytes []byte) (result parsedMessage, err error) {
	var m message0503
	err = xml.Unmarshal(xmlBytes, &m)
	if err != nil {
		return
	}
	result.MessageHead = m.MessageHead
	result.RootRecords = &db.RootRecords{
		Files:     m.FileRecords,
		Processes: m.ProcessRecords,
		Documents: m.DocumentRecords,
	}
	result.XdomeaVersion, err = extractVersion(m.XMLName.Space)
	return
}

func parse0505Message(xmlBytes []byte) (result parsedMessage, err error) {
	var m message0505
	err = xml.Unmarshal(xmlBytes, &m)
	if err != nil {
		return
	}
	result.MessageHead = m.MessageHead
	result.XdomeaVersion, err = extractVersion(m.XMLName.Space)
	return
}

func extractVersion(namespace string) (XdomeaVersion, error) {
	var version string
	var xdomeaVersion XdomeaVersion
	if namespaceRegex.MatchString(namespace) {
		version = namespaceRegex.FindStringSubmatch(namespace)[1]
	} else {
		return XdomeaVersion{}, errors.New("xdomea version can't be extracted from xdomea namespace")
	}
	xdomeaVersion, ok := XdomeaVersions[version]
	if !ok {
		return XdomeaVersion{}, fmt.Errorf("unsupported xdomea version: %s", version)
	}
	return xdomeaVersion, nil
}
