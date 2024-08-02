package xdomea

import (
	"context"
	"lath/xman/internal/db"

	"github.com/google/uuid"
)

type PackagingDecision string

const (
	// Do not create a package for the given record or its sub records. However,
	// the given record might be packaged with a higher-level record.
	PackagingDecisionNone PackagingDecision = ""
	// Package the given record into a single archive package.
	PackagingDecisionSingle PackagingDecision = "single"
	// Package the given record's sub or sub-sub records with the packaging
	// decision "single". Additionally create a single package for all direct
	// sub records with packaging decision "none".
	PackagingDecisionSub PackagingDecision = "sub"
)

func PackagingDecisions(processID uuid.UUID) map[uuid.UUID]PackagingDecision {
	options := make(map[uuid.UUID]db.PackagingOption)
	for _, o := range db.FindRecordOptionsForProcess(context.Background(), processID) {
		options[o.RecordID] = o.Packaging
	}
	rootRecords := db.FindRootRecords(context.Background(), processID, db.MessageType0503)
	decisions := make(map[uuid.UUID]PackagingDecision)
	for _, f := range rootRecords.Files {
		if options[f.RecordID] == db.PackagingOptionDefault {
			decisions[f.RecordID] = PackagingDecisionSingle
		} else {
			nSubPackages := processFileRecordsForSubPackaging(f.Subfiles, options[f.RecordID], options, decisions)
			nSubPackages += processProcessRecordsForSubPackaging(f.Processes, options[f.RecordID], decisions)
			if nSubPackages > 0 {
				decisions[f.RecordID] = PackagingDecisionSub
			} else {
				decisions[f.RecordID] = PackagingDecisionSingle
			}
		}
	}
	return decisions
}

func processFileRecordsForSubPackaging(
	records []db.FileRecord,
	option db.PackagingOption,
	options map[uuid.UUID]db.PackagingOption,
	decisions map[uuid.UUID]PackagingDecision,
) (nSubPackages int) {
	for _, f := range records {
		nSubPackages++
		if options[f.RecordID] != db.PackagingOptionDefault {
			option = options[f.RecordID]
		}
		switch option {
		case db.PackagingOptionSubFile:
			decisions[f.RecordID] = PackagingDecisionSingle
		case db.PackagingOptionProcess:
			decisions[f.RecordID] = PackagingDecisionSub
			nSubPackages += processFileRecordsForSubPackaging(f.Subfiles, option, options, decisions)
			nSubPackages += processProcessRecordsForSubPackaging(f.Processes, option, decisions)
		}
	}
	return nSubPackages
}

func processProcessRecordsForSubPackaging(
	records []db.ProcessRecord,
	option db.PackagingOption,
	decisions map[uuid.UUID]PackagingDecision,
) (nSubPackages int) {
	for _, p := range records {
		if option == db.PackagingOptionProcess {
			nSubPackages++
			decisions[p.RecordID] = PackagingDecisionSingle
		}
	}
	return nSubPackages
}
