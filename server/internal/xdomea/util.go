package xdomea

import (
	"errors"
	"lath/xman/internal/db"
	"path/filepath"
	"regexp"

	"github.com/google/uuid"
)

var uuidRegexString = "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}"
var Message0501MessageSuffix = "_Aussonderung.Anbieteverzeichnis.0501"
var Message0502MessageSuffix = "_Aussonderung.Bewertungsverzeichnis.0502"
var Message0503MessageSuffix = "_Aussonderung.Aussonderung.0503"
var Message0504MessageSuffix = "_Aussonderung.AnbietungEmpfangBestaetigen.0504"
var Message0505MessageSuffix = "_Aussonderung.BewertungEmpfangBestaetigen.0505"
var Message0506MessageSuffix = "_Aussonderung.AussonderungImportBestaetigen.0506"
var message0501RegexString = uuidRegexString + Message0501MessageSuffix + ".zip"
var message0503RegexString = uuidRegexString + Message0503MessageSuffix + ".zip"
var message0505RegexString = uuidRegexString + Message0505MessageSuffix + ".zip"
var uuidRegex = regexp.MustCompile(uuidRegexString)
var message0501Regex = regexp.MustCompile(message0501RegexString)
var message0503Regex = regexp.MustCompile(message0503RegexString)
var message0505Regex = regexp.MustCompile(message0505RegexString)
var namespaceRegex = regexp.MustCompile("^urn:xoev-de:xdomea:schema:([0-9].[0-9].[0=9])$")

func isMessage(path string) bool {
	fileName := filepath.Base(path)
	return message0501Regex.MatchString(fileName) ||
		message0503Regex.MatchString(fileName) ||
		message0505Regex.MatchString(fileName)
}

func getMessageTypeImpliedByPath(path string) (db.MessageType, error) {
	fileName := filepath.Base(path)
	if message0501Regex.MatchString(fileName) {
		return db.MessageType0501, nil
	} else if message0503Regex.MatchString(fileName) {
		return db.MessageType0503, nil
	} else if message0505Regex.MatchString(fileName) {
		return db.MessageType0505, nil
	}
	return "", errors.New("unknown message type: " + path)
}

func getProcessID(path string) (uuid.UUID, error) {
	fileName := filepath.Base(path)
	s := uuidRegex.FindString(fileName)
	processID, err := uuid.Parse(s)
	if err != nil {
		return processID, errors.New("failed to parse process id: " + err.Error())
	}
	return processID, nil
}

func getMessageName(id uuid.UUID, messageType db.MessageType) string {
	messageSuffix := ""
	switch messageType {
	case "0501":
		messageSuffix = Message0501MessageSuffix
	case "0503":
		messageSuffix = Message0503MessageSuffix
	case "0505":
		messageSuffix = Message0505MessageSuffix
	default:
		panic("message type not supported: " + messageType)
	}
	return id.String() + messageSuffix + ".xml"
}
