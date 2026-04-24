// Package xoev provides utility functions for processing messages based on the XÖV standards
// xdomea and XJustiz.
package xoev

import (
	"path/filepath"
	"slices"
	"strings"
)

// fileExtensions contains all valid filename extensions for XÖV container files.
// Extensions must be specified without a leading dot.
var fileExtensions = []string{"zip", "xdomea"}

// IsXoevMessage checks if a file could be a supported XÖV message. The names for the container
// files of XÖV messages aren't reliable. Therefore, verification can only be performed based on
// the file extension.
func IsXoevMessage(path string) bool {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	return slices.Contains(fileExtensions, ext)
}
