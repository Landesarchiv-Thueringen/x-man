package report

import (
	"lath/xman/internal/db"
	"regexp"
)

type RecordObjectType = string

const (
	File       RecordObjectType = "file"
	SubFile    RecordObjectType = "subFile"
	Process    RecordObjectType = "process"
	SubProcess RecordObjectType = "subProcess"
)

type ArchivePackageData struct {
	Title    string
	Lifetime struct {
		Start string
		End   string
	}
	AppraisalNote string
	TotalFileSize uint64
	PackageID     string
}

func getArchivePackages(process db.Process) (result []ArchivePackageData) {
	aips := db.GetArchivePackagesWithAssociations(process.ID)
	result = make([]ArchivePackageData, len(aips))
	for i, a := range aips {
		result[i] = ArchivePackageData{
			Title:         a.IOTitle,
			Lifetime:      getLifetime(a),
			AppraisalNote: getAppraisalNote(a),
			TotalFileSize: getTotalFileSize(a),
			PackageID:     a.PackageID,
		}
	}
	return
}

// getLifetime decomposes the combined lifetime string of an archive package of
// the form
//
//	"2099-01-31 - 2099-02-28"
//
// where all components are optional, but panics if the sting does not satisfy
// this form.
func getLifetime(a db.ArchivePackage) struct {
	Start string
	End   string
} {
	re := regexp.MustCompile(`^(?:(\d{4}-\d{2}-\d{2})\s+)?-?(?:\s+(\d{4}-\d{2}-\d{2}))?$`)
	match := re.FindStringSubmatch(a.IOLifetimeCombined)
	return struct {
		Start string
		End   string
	}{Start: match[1], End: match[2]}
}

// getAppraisalNote returns the appraisal note of the first record object that
// belongs to the archive package.
func getAppraisalNote(aip db.ArchivePackage) string {
	if len(aip.FileRecordObjects) > 0 {
		a := db.GetAppraisal(aip.Process.ID, aip.FileRecordObjects[0].XdomeaID)
		return a.InternalNote
	}
	if len(aip.ProcessRecordObjects) > 0 {
		a := db.GetAppraisal(aip.Process.ID, aip.ProcessRecordObjects[0].XdomeaID)
		return a.InternalNote
	}
	return ""
}

// getTotalFileSize returns the total file size in bytes of all files that
// belong to the given archive package.
func getTotalFileSize(a db.ArchivePackage) uint64 {
	var totalFileSize uint64
	for _, d := range a.PrimaryDocuments {
		totalFileSize += d.FileSize
	}
	return totalFileSize
}
