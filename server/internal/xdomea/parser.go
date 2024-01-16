package xdomea

import (
	"encoding/xml"
	"errors"
	"io"
	"lath/xman/internal/db"
	"log"
	"os"
)

func ExtractVersionFromMessage(message db.Message) (db.XdomeaVersion, error) {
	var xdomeaVersion db.XdomeaVersion
	xmlFile, err := os.Open(message.MessagePath)
	if err != nil {
		log.Println(err)
		return xdomeaVersion, err
	}
	defer xmlFile.Close()
	// Read the opened xmlFile as a byte array.
	xmlBytes, err := io.ReadAll(xmlFile)
	if err != nil {
		log.Println(err)
		return xdomeaVersion, err
	}
	switch message.MessageType.Code {
	case "0501":
		var messageBody db.MessageBody0501
		err := xml.Unmarshal(xmlBytes, &messageBody)
		if err != nil {
			log.Println(err)
			return xdomeaVersion, err
		}
		return extractVersion(messageBody.XMLName.Space)
	case "0503":
		var messageBody db.MessageBody0503
		err := xml.Unmarshal(xmlBytes, &messageBody)
		if err != nil {
			log.Println(err)
			return xdomeaVersion, err
		}
		return extractVersion(messageBody.XMLName.Space)
	case "0505":
		var messageBody db.MessageBody0505
		err := xml.Unmarshal(xmlBytes, &messageBody)
		if err != nil {
			log.Println(err)
			return xdomeaVersion, err
		}
		return extractVersion(messageBody.XMLName.Space)
	default:
		errorMessage := "message type can't be parsed"
		log.Println(errorMessage)
		return xdomeaVersion, errors.New(errorMessage)
	}
}

func ParseMessage(message db.Message) (db.Message, error) {
	xmlFile, err := os.Open(message.MessagePath)
	if err != nil {
		return message, err
	}
	defer xmlFile.Close()
	// Read the opened xmlFile as a byte array.
	xmlBytes, err := io.ReadAll(xmlFile)
	if err != nil {
		return message, err
	}
	switch message.MessageType.Code {
	case "0501":
		return parse0501Message(message, xmlBytes)
	case "0503":
		return parse0503Message(message, xmlBytes)
	case "0505":
		return parse0505Message(message, xmlBytes)
	default:
		errorMessage := "message type can't be parsed"
		return message, errors.New(errorMessage)
	}
}

func parse0501Message(message db.Message, xmlBytes []byte) (db.Message, error) {
	var m db.Message0501
	err := xml.Unmarshal(xmlBytes, &m)
	if err != nil {
		return message, err
	}
	version, err := extractVersion(m.XMLName.Space)
	if err != nil {
		return message, err
	}
	CorrectXdomeaVersionDifferences(m.RecordObjects)
	message.MessageHead = m.MessageHead
	message.RecordObjects = m.RecordObjects
	message.XdomeaVersion = version.Code
	return message, nil
}

func parse0503Message(message db.Message, xmlBytes []byte) (db.Message, error) {
	var m db.Message0503
	err := xml.Unmarshal(xmlBytes, &m)
	if err != nil {
		return message, err
	}
	version, err := extractVersion(m.XMLName.Space)
	if err != nil {
		return message, err
	}
	CorrectXdomeaVersionDifferences(m.RecordObjects)
	message.MessageHead = m.MessageHead
	message.RecordObjects = m.RecordObjects
	message.XdomeaVersion = version.Code
	return message, nil
}

func parse0505Message(message db.Message, xmlBytes []byte) (db.Message, error) {
	var m db.Message0505
	err := xml.Unmarshal(xmlBytes, &m)
	if err != nil {
		return message, err
	}
	version, err := extractVersion(m.XMLName.Space)
	if err != nil {
		return message, err
	}
	message.MessageHead = m.MessageHead
	message.XdomeaVersion = version.Code
	return message, nil
}

func extractVersion(namespace string) (db.XdomeaVersion, error) {
	var version string
	var xdomeaVersion db.XdomeaVersion
	if namespaceRegex.MatchString(namespace) {
		version = namespaceRegex.FindStringSubmatch(namespace)[1]
	} else {
		errorMessage := "xdomea version can't be extracted from xdomea namespace"
		log.Println(errorMessage)
		return xdomeaVersion, errors.New(errorMessage)
	}
	xdomeaVersion, err := db.GetXdomeaVersionByCode(version)
	if err != nil {
		errorMessage := "unsupported xdomea version"
		log.Println(errorMessage)
		return xdomeaVersion, errors.New(errorMessage)
	}
	return xdomeaVersion, nil
}

// CorrectXdomeaVersionDifferences corrects all differences between xdomea versions for further processing.
func CorrectXdomeaVersionDifferences(recordObjects []db.RecordObject) {
	for recordObjectIndex := range recordObjects {
		if recordObjects[recordObjectIndex].FileRecordObject != nil {
			recordObjects[recordObjectIndex].FileRecordObject.SetVersionIndependentSubFiles()
		}
	}
}
