package core

import (
	"errors"
	"lath/xman/internal/db"
	"path/filepath"
	"regexp"
)

const uuidRegexString = "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}"
const Message0501MessageSuffix = "_Aussonderung.Anbieteverzeichnis.0501"
const Message0502MessageSuffix = "_Aussonderung.Bewertungsverzeichnis.0502"
const Message0503MessageSuffix = "_Aussonderung.Aussonderung.0503"
const Message0504MessageSuffix = "_Aussonderung.AnbietungEmpfangBestaetigen.0504"
const Message0505MessageSuffix = "_Aussonderung.BewertungEmpfangBestaetigen.0505"
const Message0506MessageSuffix = "_Aussonderung.AussonderungImportBestaetigen.0506"
const Message0507MessageSuffix = "_Aussonderung.AussonderungEmpfangBestaetigen.0507"

var uuidRegex = regexp.MustCompile(uuidRegexString)
var message0501Regex = regexp.MustCompile(uuidRegexString + Message0501MessageSuffix + ".zip")
var message0503Regex = regexp.MustCompile(uuidRegexString + Message0503MessageSuffix + ".zip")
var message0505Regex = regexp.MustCompile(uuidRegexString + Message0505MessageSuffix + ".zip")
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

func getProcessID(path string) string {
	fileName := filepath.Base(path)
	processID := uuidRegex.FindString(fileName)
	if processID == "" {
		panic("failed to read process id from: " + fileName)
	}
	return processID
}

func getMessageName(id string, messageType db.MessageType) string {
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
	return id + messageSuffix + ".xml"
}
