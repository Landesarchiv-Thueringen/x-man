package xdomea

import (
	"io"
	"lath/xman/internal/db"
	"log"
	"os"

	"github.com/lestrrat-go/libxml2"
	"github.com/lestrrat-go/libxml2/types"
	"github.com/lestrrat-go/libxml2/xsd"
)

// ValidateXdomeaXmlFile performs a xsd schema validation against the XML file of a xdomea message.
func ValidateXdomeaXmlFile(xmlPath string, version db.XdomeaVersion) error {
	schema, err := xsd.ParseFromFile(version.XSDPath)
	if err != nil {
		log.Println(err)
		return err
	}
	defer schema.Free()
	xmlFile, err := os.Open(xmlPath)
	if err != nil {
		log.Println(err)
		return err
	}
	defer xmlFile.Close()
	xmlBytes, err := io.ReadAll(xmlFile)
	if err != nil {
		log.Println(err)
		return err
	}
	xml, err := libxml2.Parse(xmlBytes)
	if err != nil {
		log.Println(err)
		return err
	}
	defer xml.Free()
	err = validateXdomeaXml(schema, xml)
	return nil
}

// ValidateXdomeaXmlString performs a xsd schema validation against the XML code of a xdomea message.
func ValidateXdomeaXmlString(xmlText string, version db.XdomeaVersion) error {
	schema, err := xsd.ParseFromFile(version.XSDPath)
	if err != nil {
		log.Println(err)
		return err
	}
	defer schema.Free()
	xml, err := libxml2.ParseString(xmlText)
	if err != nil {
		log.Println(err)
		return err
	}
	defer xml.Free()
	err = validateXdomeaXml(schema, xml)
	return err
}

// ValidateXdomeaXml performs a xsd schema validation of a parsed xdomea XML file.
func validateXdomeaXml(schema *xsd.Schema, xml types.Document) error {
	err := schema.Validate(xml)
	// Print all schema errors.
	if err != nil {
		validationError, ok := err.(xsd.SchemaValidationError)
		if ok {
			for _, e := range validationError.Errors() {
				log.Printf("error: %s", e.Error())
			}
		}
		return err
	}
	return nil
}
