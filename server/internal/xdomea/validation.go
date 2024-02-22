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
// Returns (true, nil) if the schema validation didn't find an error.
// Returns (false, SchemaValidationError) if the schema validation found errors.
// All schema errors can be extracted from the SchemaValidationError.
// Returns (false, error) if another error happened.
func ValidateXdomeaXmlFile(xmlPath string, version db.XdomeaVersion) (bool, error) {
	schema, err := xsd.ParseFromFile(version.XSDPath)
	if err != nil {
		log.Println(err)
		return false, err
	}
	defer schema.Free()
	xmlFile, err := os.Open(xmlPath)
	if err != nil {
		log.Println(err)
		return false, err
	}
	defer xmlFile.Close()
	xmlBytes, err := io.ReadAll(xmlFile)
	if err != nil {
		log.Println(err)
		return false, err
	}
	xml, err := libxml2.Parse(xmlBytes)
	if err != nil {
		log.Println(err)
		return false, err
	}
	defer xml.Free()
	return validateXdomeaXml(schema, xml)
}

// ValidateXdomeaXmlString performs a xsd schema validation against the XML code of a xdomea message.
// Returns (true, nil) if the schema validation didn't find an error.
// Returns (false, SchemaValidationError) if the schema validation found errors.
// All schema errors can be extracted from the SchemaValidationError.
// Returns (false, error) if another error happened.
func ValidateXdomeaXmlString(xmlText string, version db.XdomeaVersion) (bool, error) {
	schema, err := xsd.ParseFromFile(version.XSDPath)
	if err != nil {
		log.Println(err)
		return false, err
	}
	defer schema.Free()
	xml, err := libxml2.ParseString(xmlText)
	if err != nil {
		log.Println(err)
		return false, err
	}
	defer xml.Free()
	return validateXdomeaXml(schema, xml)
}

// ValidateXdomeaXml performs a xsd schema validation of a parsed xdomea XML file.
// Returns (true, nil) if the schema validation didn't find an error.
// Returns (false, SchemaValidationError) if the schema validation found errors.
// All schema errors can be extracted from the SchemaValidationError.
func validateXdomeaXml(schema *xsd.Schema, xml types.Document) (bool, error) {
	err := schema.Validate(xml)
	// log all schema errors.
	if err != nil {
		return false, err
	}
	return true, nil
}
