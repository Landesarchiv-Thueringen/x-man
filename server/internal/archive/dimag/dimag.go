package dimag

import (
	"context"
	"errors"
	"fmt"
	"lath/xman/internal/db"
	"time"
)

const pollIntervalMin = time.Second * 1
const pollIntervalMax = time.Second * 10

type jobFailedError struct {
	action    string
	jobID     int
	jobStatus getJobStatusResponse
}

func (err *jobFailedError) Error() string {
	message := err.jobStatus.Message
	if message == "" {
		message = err.jobStatus.ReturnMessage
	}
	return fmt.Sprintf(
		"DIMAG %s job %d: status %d: %s",
		err.action, err.jobID, err.jobStatus.Status, message,
	)
}

func IsJobFailedError(err error) bool {
	target := &jobFailedError{}
	return errors.As(err, &target)
}

// StartImport starts a DIMAG job for archiving a record object in DIMAG.
func StartImport(
	ctx context.Context,
	process db.SubmissionProcess,
	message db.Message,
	aip *db.ArchivePackage,
	c Connection,
) (jobID int, err error) {
	bagit := createArchiveBagit(process, message, *aip)
	uploadDir, err := uploadBagit(ctx, c, bagit)
	if err != nil {
		return 0, err
	}
	jobID, err = importBag(ctx, uploadDir)
	if err != nil {
		return 0, err
	}
	// We remove the BagIt when it was processed without errors. Otherwise, we
	// leave it for debugging purposes.
	bagit.Remove()
	return jobID, nil
}

// WaitForArchiveJob periodically polls DIMAG for the status of an import job
// and fills in the package ID for the AIP when done.
func WaitForArchiveJob(
	ctx context.Context,
	jobID int,
	aip *db.ArchivePackage,
) error {
	var pollInterval time.Duration
	var jobStatus getJobStatusResponse
	var err error
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		jobStatus, err = getJobStatus(jobID)
		if err != nil {
			return err
		}
		if jobStatus.Status == 100 {
			pollInterval = updatePollInterval(pollInterval)
			time.Sleep(pollInterval)
			continue
		} else if jobStatus.Status == 200 {
			break
		} else {
			return &jobFailedError{"importBag", jobID, jobStatus}
		}
	}
	packageID, err := packageID(jobStatus)
	if err != nil {
		fmt.Printf("%#v\n", jobStatus)
		return err
	}
	aip.PackageID = packageID
	ok := db.ReplaceArchivePackage(aip)
	if !ok {
		return fmt.Errorf("failed to set PackageID for archive package %v", aip.ID.Hex())
	}
	return nil
}

// updatePollInterval returns a new poll interval given the previous one.
//
// It gradually increases the interval starting at pollIntervalMin and going no
// further than pollIntervalMax.
func updatePollInterval(prev time.Duration) time.Duration {
	if prev == 0 {
		return pollIntervalMin
	}
	return min(time.Duration(float64(prev)*float64(1.5)), pollIntervalMax)
}
