package dimag

// This file contains types and methods that enable usage of the BagIt format.
// It is not specific to any DIMAG functionality.

import (
	"crypto/sha512"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

const bagItTxt = `BagIt-Version: 1.0
Tag-File-Character-Encoding: UTF-8
`

// bagitHandle represents a BagIt structure for upload to an archive system.
//
// The BagIt is constructed on the filesystem under the path obtained by `Path`.
//
// When done, call `Remove` to clean up the filesystem.
type bagitHandle struct {
	id uuid.UUID
}

func makeBagit() bagitHandle {
	bagIt := bagitHandle{
		id: uuid.New(),
	}
	err := os.MkdirAll(bagIt.Path(), 0755)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(filepath.Join(bagIt.Path(), "bagit.txt"), []byte(bagItTxt), 0644)
	if err != nil {
		panic(err)
	}
	return bagIt
}

func (h *bagitHandle) ID() uuid.UUID {
	return h.id
}

// Path returns the BagIt's path on the local filesystem.
func (h *bagitHandle) Path() string {
	if os.Getenv("DEBUG_MODE") == "true" {
		return "/debug-data/bagit_" + h.id.String()
	} else {
		return filepath.Join(os.TempDir(), "bagit_"+h.id.String())
	}
}

// CreateFile creates a file and adds it to the BagIt.
func (h *bagitHandle) CreateFile(bagitPath string, content []byte) {
	dstPath := filepath.Join(h.Path(), bagitPath)
	err := os.MkdirAll(filepath.Dir(dstPath), 0755)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(dstPath, content, 0644)
	if err != nil {
		panic(err)
	}
}

// CopyFile copies an existing file from the local filesystem to the BagIt.
func (h *bagitHandle) CopyFile(bagitPath string, srcPath string) {
	dstPath := filepath.Join(h.Path(), bagitPath)
	src, err := os.Open(srcPath)
	if err != nil {
		panic(err)
	}
	defer src.Close()
	dst, err := os.Create(dstPath)
	if err != nil {
		panic(err)
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	if err != nil {
		panic(err)
	}
}

// Finalize calculates and saves the BagIt's checksums, making the BagIt ready
// for transmission.
func (h *bagitHandle) Finalize() {
	h.createManifest()
	h.createTagManifest()
}

// Remove deletes the BagIt structure from the filesystem.
//
// After calling `remove`, the BagItHandle cannot be used anymore.
func (h *bagitHandle) Remove() {
	err := os.RemoveAll(h.Path())
	if err != nil {
		panic(err)
	}
}

func (h *bagitHandle) createManifest() {
	h.CreateFile("manifest-sha512.txt", sha512SumRecords(h.Path(), "data", true))
}

func (h *bagitHandle) createTagManifest() {
	var records []byte
	entries, err := os.ReadDir(h.Path())
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		if entry.Name() == "data" {
			continue
		}
		entryRecords := sha512SumRecords(h.Path(), entry.Name(), entry.IsDir())
		records = append(records, entryRecords...)
	}
	h.CreateFile("tagmanifest-sha512.txt", records)
}

// sha512SumRecords creates lines for a a checksum file for the given file or
// directory and all subdirectories, if any.
//
// Parameters:
// - rootPath: the root for paths in the checksum file
// - subPath: the file or directory to be checked, relative to rootPath
// - isDir: whether the entry at subPath is a directory (otherwise, file is assumed)
// Returns the contents of the checksum file.
func sha512SumRecords(rootPath string, subPath string, isDir bool) []byte {
	path := filepath.Join(rootPath, subPath)
	if isDir {
		var sums []byte
		entries, err := os.ReadDir(path)
		if err != nil {
			panic(err)
		}
		for _, entry := range entries {
			entrySubPath := filepath.Join(subPath, entry.Name())
			entrySums := sha512SumRecords(rootPath, entrySubPath, entry.IsDir())
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
