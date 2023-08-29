package xdomea

import (
	filepath "path/filepath"
	"regexp"
)

var uuidRegexString = "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}"
var message0501RegexString = uuidRegexString + "_Aussonderung.Anbieteverzeichnis.0501.zip"
var message0503RegexString = uuidRegexString + "_Aussonderung.Aussonderung.0503.zip"
var uuidRegex = regexp.MustCompile(uuidRegexString)
var message0501Regex = regexp.MustCompile(message0501RegexString)
var message0503Regex = regexp.MustCompile(message0503RegexString)

func IsMessage(path string) bool {
	fileName := filepath.Base(path)
	return message0501Regex.MatchString(fileName) || message0503Regex.MatchString(fileName)
}

func GetMessageID(path string) string {
	fileName := filepath.Base(path)
	return uuidRegex.FindString(fileName)
}
