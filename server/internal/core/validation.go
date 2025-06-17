package core

import (
	"io"
	"os"

	"github.com/lestrrat-go/libxml2"
	"github.com/lestrrat-go/libxml2/types"
	"github.com/lestrrat-go/libxml2/xsd"
)

// validateXdomeaXmlFile performs a xsd schema validation against the XML file of a xdomea message.
// Returns nil if the schema validation didn't find an error.
// Returns SchemaValidationError if the schema validation found errors.
// All schema errors can be extracted from the SchemaValidationError.
// Returns error if another error happened.
func validateXdomeaXmlFile(xmlPath string, version XdomeaVersion) error {
	schema, err := xsd.ParseFromFile(version.XSDPath)
	if err != nil {
		return err
	}
	defer schema.Free()
	xmlFile, err := os.Open(xmlPath)
	if err != nil {
		return err
	}
	defer xmlFile.Close()
	xmlBytes, err := io.ReadAll(xmlFile)
	if err != nil {
		return err
	}
	xml, err := libxml2.Parse(xmlBytes)
	if err != nil {
		return err
	}
	defer xml.Free()
	return validateXdomeaXml(schema, xml)
}

// ValidateXdomeaXmlString performs a xsd schema validation against the XML code of a xdomea message.
// Returns nil if the schema validation didn't find an error.
// Returns SchemaValidationError if the schema validation found errors.
// All schema errors can be extracted from the SchemaValidationError.
// Returns error if another error happened.
func ValidateXdomeaXmlString(xmlText string, version XdomeaVersion) error {
	schema, err := xsd.ParseFromFile(version.XSDPath)
	if err != nil {
		return err
	}
	defer schema.Free()
	xml, err := libxml2.ParseString(xmlText)
	if err != nil {
		return err
	}
	defer xml.Free()
	return validateXdomeaXml(schema, xml)
}

// ValidateXdomeaXml performs a xsd schema validation of a parsed xdomea XML file.
// Returns nil if the schema validation didn't find an error.
// Returns SchemaValidationError if the schema validation found errors.
// All schema errors can be extracted from the SchemaValidationError.
func validateXdomeaXml(schema *xsd.Schema, xml types.Document) error {
	return schema.Validate(xml)
}
