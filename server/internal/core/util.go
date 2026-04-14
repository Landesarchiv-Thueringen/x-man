package core

import (
	"errors"
	"lath/xman/internal/db"
	"path/filepath"
	"regexp"
)

const uuidRegexString = "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}"

// MessageSuffixByType contains the suffix for messages before and after xdomea version 4.0.0 .
var MessageSuffixByType = map[db.MessageType][2]string{
	db.MessageType0501: {"_Aussonderung.Anbieteverzeichnis.0501", "_0501"},
	db.MessageType0502: {"_Aussonderung.Bewertungsverzeichnis.0502", "_0502"},
	db.MessageType0503: {"_Aussonderung.Aussonderung.0503", "_0503"},
	db.MessageType0504: {"_Aussonderung.AnbietungEmpfangBestaetigen.0504", "_0504"},
	db.MessageType0505: {"_Aussonderung.BewertungEmpfangBestaetigen.0505", "_0505"},
	db.MessageType0506: {"_Aussonderung.AussonderungImportBestaetigen.0506", "_0506"},
	db.MessageType0507: {"_Aussonderung.AussonderungEmpfangBestaetigen.0507", "_0507"},
}
var uuidRegex = regexp.MustCompile(uuidRegexString)
var message0501Regex = regexp.MustCompile(uuidRegexString +
	`_(?:Aussonderung\.Anbieteverzeichnis\.)?0501\.(?:zip|xdomea)`)
var message0503Regex = regexp.MustCompile(uuidRegexString +
	`_(?:Aussonderung\.Aussonderung\.)?0503\.(?:zip|xdomea)`)
var message0505Regex = regexp.MustCompile(uuidRegexString +
	`_(?:Aussonderung\.BewertungEmpfangBestaetigen\.)?0505\.(?:zip|xdomea)`)
var namespaceRegex = regexp.MustCompile(`^urn:xoev-de:xdomea:schema:([0-9]\.[0-9]\.[0-9])$`)

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

func getMessageName(id string, messageType db.MessageType, preV4 bool) string {
	return id + getMessageSuffix(messageType, preV4) + ".xml"
}

func getMessageSuffix(messageType db.MessageType, preV4 bool) string {
	messageSuffix, ok := MessageSuffixByType[messageType]
	if !ok {
		panic("unsupported message type: " + messageType)
	}
	if preV4 {
		return messageSuffix[0]
	}
	return messageSuffix[1]
}
