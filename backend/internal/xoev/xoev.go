// Package xoev provides utility functions for processing messages based on the XOEV standards
// xdomea and XJustiz.
package xoev

import (
	"path/filepath"
	"slices"
	"strings"
)

// fileExtensions contains all valid filename extensions for XOEV container files.
// Extensions must be specified without a leading dot.
var fileExtensions = []string{"zip", "xdomea"}

// IsXoevMessage checks if a file could be a supported XOEV message. The names for the container
// files of XOEV messages aren't reliable. Therefore, verification can only be performed based on
// the file extension.
func IsXoevMessage(path string) bool {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	return slices.Contains(fileExtensions, ext)
}
