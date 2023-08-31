package xdomea

import (
	"encoding/xml"
	"io/ioutil"
	"lath/xdomea/internal/db"
	"log"
	"os"
)

func ParseMessage(message db.Message) {
	xmlFile, err := os.Open(message.MessagePath)
	if err != nil {
		log.Fatal(err)
	}
	defer xmlFile.Close()
	// read our opened xmlFile as a byte array.
	xmlBytes, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		log.Fatal(err)
	}
	var messageEl db.Message0501
	err = xml.Unmarshal(xmlBytes, &messageEl)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(messageEl)
}
