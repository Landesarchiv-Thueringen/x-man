package report

import (
	"context"
	"fmt"
	"lath/xman/internal/db"
	"lath/xman/internal/xdomea"
	"reflect"

	"github.com/google/uuid"
)

type RecordObjectType = string

const (
	File       RecordObjectType = "file"
	SubFile    RecordObjectType = "subFile"
	Process    RecordObjectType = "process"
	SubProcess RecordObjectType = "subProcess"
)

// ArchivePackageStructure is a wrapper structure for archive-package data.
//
// ~~Each path to a leaf node contains exactly one node with AIP data.~~
// ~~Usually, leaf nodes will be the ones to contain AIP data, but in cases where
// sub-records of an AIP were appraised individually, further child nodes might
// be possible.~~
// Leaf nodes and only leaf nodes contain AIP data.
type ArchivePackageStructure struct {
	Title    string                    // iff AIP == nil
	Children []ArchivePackageStructure // iff AIP == nil
	AIP      *ArchivePackageData
}

type ArchivePackageData struct {
	Title         string
	Type          db.RecordType
	Lifetime      *db.Lifetime
	TotalFileSize int64
	PackageID     string
}

// archivePackagesInfo returns information about archived packages of the given
// submission process for usage in the report.
func archivePackagesInfo(
	ctx context.Context,
	process db.SubmissionProcess,
) (result []ArchivePackageStructure) {
	rootRecords := db.FindAllRootRecords(ctx, process.ProcessID, db.MessageType0503)
	aips := db.FindArchivePackagesForProcess(ctx, process.ProcessID)
	for _, f := range rootRecords.Files {
		result = append(result, archivePackagesInfoForFile(f, aips[:], []uuid.UUID{}))
	}
	for _, p := range rootRecords.Processes {
		result = append(result, archivePackagesInfoForProcess(p, aips[:], []uuid.UUID{}))
	}
	if len(rootRecords.Documents) > 0 {
		result = append(result,
			archivePackagesInfoForDocuments(rootRecords.Documents[:], aips[:], []uuid.UUID{}),
		)
	}
	return
}

func archivePackagesInfoForFile(
	file db.FileRecord,
	aips []db.ArchivePackage,
	path []uuid.UUID,
) ArchivePackageStructure {
	var subAIPs []db.ArchivePackage
	fullPath := append(path, file.RecordID)
	for _, aip := range aips {
		// File AIPs only contain one record
		if aip.RecordIDs[0] == file.RecordID {
			return ArchivePackageStructure{
				AIP: &ArchivePackageData{
					Title:         aip.IOTitle,
					Type:          db.RecordTypeFile,
					Lifetime:      aip.IOLifetime,
					TotalFileSize: getTotalFileSize(context.Background(), aip),
					PackageID:     aip.PackageID,
				},
			}
		} else if len(aip.RecordPath) >= len(fullPath) &&
			reflect.DeepEqual(aip.RecordPath[:len(fullPath)], fullPath) {
			subAIPs = append(subAIPs, aip)
		}
	}
	if len(subAIPs) == 0 {
		panic("no archive package found for file " + file.RecordID.String())
	}
	var children []ArchivePackageStructure
	for _, s := range file.Subfiles {
		children = append(children, archivePackagesInfoForFile(s, subAIPs[:], fullPath))
	}
	for _, s := range file.Processes {
		children = append(children, archivePackagesInfoForProcess(s, subAIPs[:], fullPath))
	}
	if len(file.Documents) > 0 {
		children = append(children, archivePackagesInfoForDocuments(file.Documents[:], subAIPs[:], fullPath))
	}
	return ArchivePackageStructure{
		Title:    xdomea.FileRecordTitle(file, len(path) > 0),
		Children: children,
	}
}

func archivePackagesInfoForProcess(
	process db.ProcessRecord,
	aips []db.ArchivePackage,
	path []uuid.UUID,
) ArchivePackageStructure {
	var subAIPs []db.ArchivePackage
	fullPath := append(path, process.RecordID)
	for _, aip := range aips {
		// Process AIPs only contain one record
		if aip.RecordIDs[0] == process.RecordID {
			return ArchivePackageStructure{
				AIP: &ArchivePackageData{
					Title:         aip.IOTitle,
					Type:          db.RecordTypeProcess,
					Lifetime:      aip.IOLifetime,
					TotalFileSize: getTotalFileSize(context.Background(), aip),
					PackageID:     aip.PackageID,
				},
			}
		} else if len(aip.RecordPath) >= len(fullPath) &&
			reflect.DeepEqual(aip.RecordPath[:len(fullPath)], fullPath) {
			subAIPs = append(subAIPs, aip)
		}
	}
	if len(subAIPs) == 0 {
		panic("no archive package found for process " + process.RecordID.String())
	}
	var children []ArchivePackageStructure
	for _, s := range process.Subprocesses {
		children = append(children, archivePackagesInfoForProcess(s, subAIPs[:], fullPath))
	}
	if len(process.Documents) > 0 {
		children = append(children, archivePackagesInfoForDocuments(process.Documents[:], subAIPs[:], fullPath))
	}
	return ArchivePackageStructure{
		Title:    xdomea.ProcessRecordTitle(process, false),
		Children: children,
	}
}

func archivePackagesInfoForDocuments(
	documents []db.DocumentRecord,
	aips []db.ArchivePackage,
	path []uuid.UUID,
) ArchivePackageStructure {
	ids := make(map[uuid.UUID]bool)
	for _, d := range documents {
		ids[d.RecordID] = true
	}
	for _, aip := range aips {
		if ids[aip.RecordIDs[0]] {
			return ArchivePackageStructure{
				AIP: &ArchivePackageData{
					Title:         aip.IOTitle,
					Type:          db.RecordTypeDocument,
					Lifetime:      aip.IOLifetime,
					TotalFileSize: getTotalFileSize(context.Background(), aip),
					PackageID:     aip.PackageID,
				},
			}
		}
	}
	panic(fmt.Sprintf("no archive package found for documents in path %s", path))
}

// getTotalFileSize returns the total file size in bytes of all files that
// belong to the given archive package. asdf
func getTotalFileSize(ctx context.Context, a db.ArchivePackage) int64 {
	filenames := make([]string, 0)
	for _, p := range a.PrimaryDocuments {
		filenames = append(filenames, p.Filename)
	}
	return db.CalculateTotalFileSize(ctx, a.ProcessID, filenames)
}
