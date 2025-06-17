package report

import (
	"context"
	"fmt"
	"lath/xman/internal/core"
	"lath/xman/internal/db"

	"github.com/google/uuid"
)

type AppraisalStructure struct {
	Title             string
	Children          []AppraisalStructure
	AppraisalDecision db.AppraisalDecisionOption
	AppraisalNote     string
}

func appraisalInfo(
	ctx context.Context,
	process db.SubmissionProcess,
) []AppraisalStructure {
	rootRecords := db.FindAllRootRecords(ctx, process.ProcessID, db.MessageType0501)
	records := core.AppraisableRecords(&rootRecords)
	appraisals := getAppraisalsMap(process.ProcessID)
	return appraisalInfoNodes("", rootRecords.Files, rootRecords.Processes, records, appraisals)
}

func appraisalInfoNodes(
	parentType db.RecordType,
	files []db.FileRecord,
	processes []db.ProcessRecord,
	records core.AppraisableRecordsMap,
	appraisals appraisalMap,
) []AppraisalStructure {
	var result []AppraisalStructure
	for _, file := range files {
		children := appraisalInfoNodes(db.RecordTypeFile, file.Subfiles, file.Processes, records, appraisals)
		// If there are any children of the node which have a different
		// appraisal decision than the node itself, we include all children in
		// the report.
		//
		// Otherwise, we only include those children with descendants that have a
		// different appraisal decision.
		if !hasDivergentAppraisals(file.RecordID, records, appraisals) {
			children = filterHasChildren(children)
		}
		result = append(result, AppraisalStructure{
			Title:             core.FileRecordTitle(file, parentType == db.RecordTypeFile),
			AppraisalDecision: appraisals[file.RecordID].Decision,
			AppraisalNote:     appraisals[file.RecordID].Note,
			Children:          children,
		})
	}
	for _, process := range processes {
		children := appraisalInfoNodes(db.RecordTypeProcess, []db.FileRecord{}, process.Subprocesses, records, appraisals)
		if !hasDivergentAppraisals(process.RecordID, records, appraisals) {
			children = filterHasChildren(children)
		}
		result = append(result, AppraisalStructure{
			Title:             core.ProcessRecordTitle(process, parentType == db.RecordTypeProcess),
			AppraisalDecision: appraisals[process.RecordID].Decision,
			AppraisalNote:     appraisals[process.RecordID].Note,
			Children:          children,
		})
	}
	return result
}

// filterHasChildren returns all nodes that have children.
func filterHasChildren(nodes []AppraisalStructure) []AppraisalStructure {
	var result []AppraisalStructure
	for _, node := range nodes {
		if len(node.Children) > 0 {
			result = append(result, node)
		}
	}
	return result
}

// hasDivergentAppraisals returns true if the given record has any direct
// children with a different appraisal decision than itself.
func hasDivergentAppraisals(
	recordID uuid.UUID,
	records core.AppraisableRecordsMap,
	appraisals appraisalMap,
) bool {
	fmt.Println("hasDivergentAppraisals", recordID, records[recordID].Type, appraisals[recordID].Decision)
	if appraisals[recordID].Decision != db.AppraisalDecisionA {
		// All sub records of discarded records are discarded as well.
		return false
	}
	for _, childID := range records[recordID].Children {
		fmt.Println("  child", childID, appraisals[childID].Decision)
		if appraisals[childID].Decision != db.AppraisalDecisionA {
			return true
		}
	}
	return false
}
