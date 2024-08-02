package xdomea

import (
	"context"
	"lath/xman/internal/db"

	"github.com/google/uuid"
)

// PackagingDecision is the computed decision of how to package a record, based
// on the chosen packaging option of its root record.
type PackagingDecision string

const (
	// Do not create a package for the given record or its sub records. The
	// given record will be packaged with a higher-level record.
	PackagingDecisionNone PackagingDecision = ""
	// Package the given record into a single archive package.
	PackagingDecisionSingle PackagingDecision = "single"
	// Package the given record's sub or sub-sub records with the packaging
	// decision "single". Additionally create a single package for all direct
	// sub records with packaging decision "none".
	PackagingDecisionSub PackagingDecision = "sub"
)

type PackagingStats struct {
	Files                int  `json:"files"`
	Subfiles             int  `json:"subfiles"`
	Processes            int  `json:"processes"`
	Other                int  `json:"other"`
	DeepestLevelHasItems bool `json:"deepestLevelHasItems"`
}

func (s *PackagingStats) Total() int {
	return s.Files + s.Subfiles + s.Processes + s.Other
}

func (s *PackagingStats) add(s2 PackagingStats) {
	s.Files += s2.Files
	s.Subfiles += s2.Subfiles
	s.Processes += s2.Processes
	s.Other += s2.Other
	s.DeepestLevelHasItems = s.DeepestLevelHasItems || s2.DeepestLevelHasItems
}

func Packaging(processID uuid.UUID) (
	decisions map[uuid.UUID]PackagingDecision,
	stats map[uuid.UUID]PackagingStats,
	options map[uuid.UUID]db.PackagingOption,
) {
	options = make(map[uuid.UUID]db.PackagingOption)
	for _, o := range db.FindRecordOptionsForProcess(context.Background(), processID) {
		options[o.RecordID] = o.Packaging
	}
	rootRecords := db.FindAllRootRecords(context.Background(), processID, db.MessageType0503)
	decisions = make(map[uuid.UUID]PackagingDecision)
	stats = make(map[uuid.UUID]PackagingStats)
	for _, f := range rootRecords.Files {
		stats[f.RecordID] = packagingFileRecord(f, db.PackagingOptionRoot, 0, options, decisions)
	}
	packagingProcessRecords(rootRecords.Processes, decisions)
	packagingDocumentRecords(rootRecords.Documents)
	return decisions, stats, options
}

func PackagingStatsForOptions(rootRecords []db.FileRecord) map[db.PackagingOption]PackagingStats {
	result := make(map[db.PackagingOption]PackagingStats)
	dummyDecisions := make(map[uuid.UUID]PackagingDecision)
	for _, o := range []db.PackagingOption{
		db.PackagingOptionRoot,
		db.PackagingOptionLevel1,
		db.PackagingOptionLevel2,
	} {
		var stats PackagingStats
		for _, f := range rootRecords {
			stats.add(packagingFileRecord(f, o, 0, nil, dummyDecisions))
		}
		result[o] = stats
	}
	return result
}

// packagingFileRecord calculates the packaging decisions for the
// given file record. Additionally, it returns statistics about the number of
// packages to be created for the respective record types.
func packagingFileRecord(
	record db.FileRecord,
	option db.PackagingOption,
	level int,
	options map[uuid.UUID]db.PackagingOption,
	decisions map[uuid.UUID]PackagingDecision,
) PackagingStats {
	if options[record.RecordID] != "" {
		option = options[record.RecordID]
	}
	switch {
	case option == db.PackagingOptionRoot && level == 0:
		decisions[record.RecordID] = PackagingDecisionSingle
		return PackagingStats{Files: 1, DeepestLevelHasItems: true}
	case level == optionLevel(option):
		decisions[record.RecordID] = PackagingDecisionSingle
		return PackagingStats{Subfiles: 1, DeepestLevelHasItems: true}
	case level < optionLevel(option):
		var stats PackagingStats
		for _, f := range record.Subfiles {
			stats.add(packagingFileRecord(f, option, level+1, options, decisions))
		}
		stats.Processes += packagingProcessRecords(record.Processes, decisions)
		if stats.Total() == 0 {
			decisions[record.RecordID] = PackagingDecisionSingle
			return PackagingStats{Subfiles: 1, DeepestLevelHasItems: false}
		}
		decisions[record.RecordID] = PackagingDecisionSub
		stats.Other += packagingDocumentRecords(record.Documents)
		stats.DeepestLevelHasItems = stats.DeepestLevelHasItems || level+1 == optionLevel(option)
		return stats
	default:
		panic("unreachable")
	}
}

func optionLevel(o db.PackagingOption) int {
	switch o {
	case db.PackagingOptionRoot:
		return 0
	case db.PackagingOptionLevel1:
		return 1
	case db.PackagingOptionLevel2:
		return 2
	default:
		panic("unreachable")
	}
}

// processFileRecordsForSubPackaging calculates the packaging decisions for the
// given process records. Returns the number of packages to be created.
func packagingProcessRecords(
	records []db.ProcessRecord,
	decisions map[uuid.UUID]PackagingDecision,
) (nPackages int) {
	for _, p := range records {
		// Process is the minimal record type building a package. No sub
		// packaging will be performed below this level.
		nPackages++
		decisions[p.RecordID] = PackagingDecisionSingle
	}
	return nPackages
}

// packagingDocumentRecords returns the number of packages to be
// created for document records that are not already part of another package.
func packagingDocumentRecords(
	records []db.DocumentRecord,
) (nPackages int) {
	// Loose documents will be collected into a single archive package.
	if len(records) > 0 {
		return 1
	} else {
		return 0
	}
}
