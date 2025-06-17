// Package routines calls reoccurring tasks in regular intervals.
package routines

import (
	"context"
	"fmt"
	"lath/xman/internal/core"
	"lath/xman/internal/db"
	"lath/xman/internal/errors"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// interval is the time interval between scheduled routine runs.
const interval = 1 * time.Hour

// Init schedules regular execution for all routines.
func Init() {
	// Run periodically
	go func() {
		for {
			log.Println("Starting cleanup routines...")
			cleanupArchivedProcesses()
			cleanupErrors()
			log.Println("Cleanup routines done")
			time.Sleep(interval)
		}
	}()
}

// cleanupArchivedProcesses deletes submission processes that have been archived
// successfully in the past.
//
// The time after which submission processes are deleted can be configured via
// the environment variable `DELETE_ARCHIVED_SUBMISSIONS_AFTER_DAYS`.
func cleanupArchivedProcesses() {
	defer errors.HandlePanic("cleanupArchivedProcesses", nil)
	deleteDeltaDays, err := strconv.Atoi(os.Getenv("DELETE_ARCHIVED_SUBMISSIONS_AFTER_DAYS"))
	if err != nil {
		panic("missing or improper env variable DELETE_ARCHIVED_SUBMISSIONS_AFTER_DAYS")
	}
	deleteBeforeTime := time.Now().Add(-1 * time.Hour * 24 * time.Duration(deleteDeltaDays))
	processes := db.FindProcesses(context.Background())
	for _, process := range processes {
		if process.ProcessState.Archiving.Complete &&
			process.ProcessState.Archiving.CompletedAt.Before(deleteBeforeTime) {
			deleteProcess(process)
		}
	}
}

// cleanupErrors deletes solved processing errors that are not associated with a
// still existing submission process.
//
// Processing errors that _are_ associated with a submission process will be
// deleted with the submission process (except application errors).
//
// The time after which errors are deleted can be configured with the
// environment variable `DELETE_ERRORS_AFTER_DAYS`
func cleanupErrors() {
	defer errors.HandlePanic("cleanupErrors", nil)
	deleteDeltaDays, err := strconv.Atoi(os.Getenv("DELETE_ERRORS_AFTER_DAYS"))
	if err != nil {
		panic("missing or improper env variable DELETE_ERRORS_AFTER_DAYS")
	}
	deleteBeforeTime := time.Now().Add(-1 * time.Hour * 24 * time.Duration(deleteDeltaDays))
	// Delete resolved process errors, that are not associated with a still
	// existing submission process.
	processIDs := make(map[uuid.UUID]bool)
	for _, p := range db.FindProcesses(context.Background()) {
		processIDs[p.ProcessID] = true
	}
	for _, e := range db.FindResolvedProcessingErrorsOlderThan(context.Background(), deleteBeforeTime) {
		if e.ProcessID == uuid.Nil || !processIDs[e.ProcessID] {
			log.Println("Deleting processing error", e.Title)
			db.DeleteProcessingError(e.ID)
		}
	}
}

// deleteProcess deletes the given process and all associated data from the
// database and removes all associated message files from the message store.
func deleteProcess(process db.SubmissionProcess) {
	found := core.DeleteProcess(process.ProcessID)
	if !found {
		panic(fmt.Sprintf("failed to delete process %v: not found", process.ProcessID))
	}
}
