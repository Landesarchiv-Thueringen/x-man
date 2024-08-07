package report

import (
	"context"
	"lath/xman/internal/db"
	"os"
	"path"
	"slices"
	"sort"
)

type FileStats struct {
	TotalFiles  uint
	TotalBytes  uint64
	PUIDEntries []PUIDEntry
}

type PUIDEntry struct {
	PUID    string
	Entries []DocumentsEntry
}

type DocumentsEntry struct {
	MimeType      string
	FormatVersion string
	Valid         *bool
	NumberFiles   uint
}

// processDocument adds the given document to the FileStats' PUIDEntries array.
func (f *FileStats) processDocument(document db.PrimaryDocumentData) {
	if document.FormatVerification == nil {
		return
	}
	PUID := document.FormatVerification.Summary.PUID
	idx := slices.IndexFunc(f.PUIDEntries, func(e PUIDEntry) bool { return e.PUID == PUID })
	if idx == -1 {
		PUIDEntry := PUIDEntry{PUID: PUID, Entries: make([]DocumentsEntry, 0)}
		PUIDEntry.processDocument(document)
		f.PUIDEntries = append(f.PUIDEntries, PUIDEntry)
	} else {
		f.PUIDEntries[idx].processDocument(document)
	}
}

// sort sorts the FileStats' PUIDEntries array.
func (f *FileStats) sort() {
	sort.Slice(f.PUIDEntries, func(i, j int) bool {
		return f.PUIDEntries[i].PUID < f.PUIDEntries[j].PUID
	})
	for _, e := range f.PUIDEntries {
		e.sort()
	}
}

// processDocument adds the given document to the PUIDEntry's Entries array.
func (p *PUIDEntry) processDocument(document db.PrimaryDocumentData) {
	mimeType := document.FormatVerification.Summary.MimeType
	formatVersion := document.FormatVerification.Summary.FormatVersion
	var valid *bool
	if document.FormatVerification.Summary.Valid {
		v := true
		valid = &v
	} else if document.FormatVerification.Summary.Invalid {
		v := false
		valid = &v
	}
	idx := slices.IndexFunc(p.Entries, func(e DocumentsEntry) bool {
		return e.MimeType == mimeType &&
			e.FormatVersion == formatVersion &&
			boolPointerMatches(e.Valid, valid)
	})
	if idx != -1 {
		p.Entries[idx].NumberFiles += 1
	} else {
		p.Entries = append(p.Entries, DocumentsEntry{
			MimeType:      mimeType,
			FormatVersion: formatVersion,
			Valid:         valid,
			NumberFiles:   1,
		})
	}
}

// sort sorts the PUIDEntry's Entries array.
func (p *PUIDEntry) sort() {
	sort.Slice(p.Entries, func(i, j int) bool {
		lhs := p.Entries[i]
		rhs := p.Entries[j]
		if lhs.MimeType != rhs.MimeType {
			return lhs.MimeType < rhs.MimeType
		} else if lhs.FormatVersion != rhs.FormatVersion {
			return lhs.FormatVersion < rhs.FormatVersion
		} else if lhs.Valid != nil && rhs.Valid != nil {
			return *lhs.Valid
		} else {
			return lhs.Valid == nil
		}
	})
}

func getFileStats(ctx context.Context, process db.SubmissionProcess) (fileStats FileStats) {
	documents := db.FindPrimaryDocumentsDataForProcess(ctx, process.ProcessID)
	message0503, _ := db.FindMessage(ctx, process.ProcessID, db.MessageType0503)
	fileStats.PUIDEntries = make([]PUIDEntry, 0)
	for _, document := range documents {
		fileStats.processDocument(document)
		fileStats.TotalFiles += 1
		fileSize := getFileSize(path.Join(message0503.StoreDir, document.Filename))
		fileStats.TotalBytes += fileSize
	}
	fileStats.sort()
	return
}

func getFileSize(path string) uint64 {
	fi, err := os.Stat(path)
	if err != nil {
		panic(err)
	}
	return uint64(fi.Size())
}

func boolPointerMatches(lhs, rhs *bool) bool {
	if lhs == nil || rhs == nil {
		return lhs == rhs
	} else {
		return *lhs == *rhs
	}
}
