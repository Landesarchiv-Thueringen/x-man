package core

import (
	"context"
	"lath/xman/internal/db"
)

// PackagingDecision is the computed decision of how to package a record, based
// on the selected packaging choice of its root record.
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

// Packaging looks up the selected packaging options for all objects of the
// given submission process and calculates the resulting packaging.
//
// - `decisions` indicates the calculated packaging decision for each record.
// - `stats` gives information about the content of each packaged record.
// - `options` contains the packaging options as selected by the user.
func Packaging(processID string) (
	decisions map[string]PackagingDecision,
	stats map[string]PackagingStats,
	choices map[string]db.PackagingChoice,
) {
	choices = make(map[string]db.PackagingChoice)
	for _, c := range db.FindPackagingChoicesForProcess(context.Background(), processID) {
		choices[c.RecordID] = c.PackagingChoice
	}
	rootRecords := db.FindAllRootRecords(context.Background(), processID, db.MessageType0503)
	decisions = make(map[string]PackagingDecision)
	stats = make(map[string]PackagingStats)
	for _, f := range rootRecords.Files {
		stats[f.RecordID] = packagingFileRecord(f, db.PackagingChoiceRoot, 0, choices, decisions)
	}
	// Add an entry for the message root to the stats map, so stats are included
	// when calculating the combined stats.
	stats["root"] = PackagingStats{
		Processes: packagingProcessRecords(rootRecords.Processes, decisions),
		Other:     packagingDocumentRecords(rootRecords.Documents),
	}
	return decisions, stats, choices
}

func PackagingStatsForChoices(rootRecords []db.FileRecord) map[db.PackagingChoice]PackagingStats {
	result := make(map[db.PackagingChoice]PackagingStats)
	dummyDecisions := make(map[string]PackagingDecision)
	for _, c := range []db.PackagingChoice{
		db.PackagingChoiceRoot,
		db.PackagingChoiceLevel1,
		db.PackagingChoiceLevel2,
	} {
		var stats PackagingStats
		for _, f := range rootRecords {
			stats.add(packagingFileRecord(f, c, 0, nil, dummyDecisions))
		}
		result[c] = stats
	}
	return result
}

// packagingFileRecord calculates the packaging decisions for the
// given file record. Additionally, it returns statistics about the number of
// packages to be created for the respective record types.
func packagingFileRecord(
	record db.FileRecord,
	choice db.PackagingChoice,
	level int,
	choices map[string]db.PackagingChoice,
	decisions map[string]PackagingDecision,
) PackagingStats {
	if choices[record.RecordID] != "" {
		choice = choices[record.RecordID]
	}
	switch {
	case choice == db.PackagingChoiceRoot && level == 0:
		decisions[record.RecordID] = PackagingDecisionSingle
		return PackagingStats{Files: 1, DeepestLevelHasItems: true}
	case level == choiceLevel(choice):
		decisions[record.RecordID] = PackagingDecisionSingle
		return PackagingStats{Subfiles: 1, DeepestLevelHasItems: true}
	case level < choiceLevel(choice):
		var stats PackagingStats
		for _, f := range record.Subfiles {
			stats.add(packagingFileRecord(f, choice, level+1, choices, decisions))
		}
		stats.Processes += packagingProcessRecords(record.Processes, decisions)
		if stats.Total() == 0 {
			decisions[record.RecordID] = PackagingDecisionSingle
			return PackagingStats{Subfiles: 1, DeepestLevelHasItems: false}
		}
		decisions[record.RecordID] = PackagingDecisionSub
		stats.Other += packagingDocumentRecords(record.Documents)
		stats.DeepestLevelHasItems = stats.DeepestLevelHasItems || level+1 == choiceLevel(choice)
		return stats
	default:
		panic("unreachable")
	}
}

func choiceLevel(c db.PackagingChoice) int {
	switch c {
	case db.PackagingChoiceRoot:
		return 0
	case db.PackagingChoiceLevel1:
		return 1
	case db.PackagingChoiceLevel2:
		return 2
	default:
		panic("unreachable")
	}
}

// processFileRecordsForSubPackaging calculates the packaging decisions for the
// given process records. Returns the number of packages to be created.
func packagingProcessRecords(
	records []db.ProcessRecord,
	decisions map[string]PackagingDecision,
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
