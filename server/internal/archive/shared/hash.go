package shared

import (
	"crypto/sha512"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Sha512Sum creates lines for a a checksum file for the given file or
// directory and all subdirectories, if any.
//
// Parameters:
// - rootPath: the root for paths in the checksum file
// - subPath: the file or directory to be checked, relative to rootPath
// - isDir: whether the entry at subPath is a directory (otherwise, file is assumed)
// Returns the contents of the checksum file.
func Sha512Sum(rootPath string, subPath string, isDir bool) []byte {
	path := filepath.Join(rootPath, subPath)
	if isDir {
		var sums []byte
		entries, err := os.ReadDir(path)
		if err != nil {
			panic(err)
		}
		for _, entry := range entries {
			entrySubPath := filepath.Join(subPath, entry.Name())
			entrySums := Sha512Sum(rootPath, entrySubPath, entry.IsDir())
			sums = append(sums, entrySums...)
		}
		return sums
	} else {
		return fmt.Appendf([]byte{}, "%x  %s\n", sha512Sum(path), subPath)
	}
}

// sha512Sum calculates the sha512 sum for a single file.
func sha512Sum(path string) []byte {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	h := sha512.New()
	_, err = io.Copy(h, f)
	if err != nil {
		panic(err)
	}
	return h.Sum(nil)
}
