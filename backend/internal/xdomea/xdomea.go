package xdomea

import (
	"lath/xdomea/internal/db"
	"path/filepath"
	"regexp"
)

var uuidRegexString = "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}"
var message0501RegexString = uuidRegexString + "_Aussonderung.Anbieteverzeichnis.0501.zip"
var message0503RegexString = uuidRegexString + "_Aussonderung.Aussonderung.0503.zip"
var uuidRegex = regexp.MustCompile(uuidRegexString)
var message0501Regex = regexp.MustCompile(message0501RegexString)
var message0503Regex = regexp.MustCompile(message0503RegexString)

func InitMessageTypes() {
	messageTypes := []*db.MessageType{
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

func IsMessage(path string) bool {
	fileName := filepath.Base(path)
	return message0501Regex.MatchString(fileName) || message0503Regex.MatchString(fileName)
}

func GetMessageID(path string) string {
	fileName := filepath.Base(path)
	return uuidRegex.FindString(fileName)
}
