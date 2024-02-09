// Package routines calls reoccurring tasks in regular intervals.
package routines

import (
	"fmt"
	"lath/xman/internal/db"
	"lath/xman/internal/messagestore"
	"log"
	"os"
	"strconv"
	"time"
)

// interval is the time interval between scheduled routine runs.
const interval = 1 * time.Hour

// Init schedules regular execution for all routines.
func Init() {
	// Run on application start
	markRunningTasksFailed()
	// Run periodically
	go func() {
		for {
			cleanupArchivedProcesses()
			time.Sleep(interval)
		}
	}()
}

// markRunningTasksFailed searches for tasks that are marked 'running' and marks
// them 'failed'.
func markRunningTasksFailed() {
	tasks, err := db.GetTasks()
	if err != nil {
		log.Println(err)
		return
	}
	for _, t := range tasks {
		if t.State == db.Running {
			t.State = db.Failed
			t.ErrorMessage = "Abgebrochen durch Neustart von X-Man"
			db.UpdateTask(t)
		}
	}
}

// cleanupArchivedProcesses deletes processes that have been archived
// successfully in the past.
//
// The time after which processes are deleted can be configured via the
// environment variable `DELETE_ARCHIVED_PROCESSES_AFTER_DAYS`.
func cleanupArchivedProcesses() {
	log.Println("Running cleanupArchivedProcesses")
	deleteDeltaDays, err := strconv.Atoi(os.Getenv("DELETE_ARCHIVED_PROCESSES_AFTER_DAYS"))
	if err != nil {
		log.Println(err)
		return
	}
	deleteBeforeTime := time.Now().Add(-1 * time.Hour * 24 * time.Duration(deleteDeltaDays))
	processes, err := db.GetProcesses()
	if err != nil {
		log.Println(err)
		return
	}
	for _, process := range processes {
		if process.ProcessState.Archiving.Complete &&
			process.ProcessState.Archiving.CompletionTime.Before(deleteBeforeTime) {
			log.Println("Deleting process", process.XdomeaID)
			err = deleteProcess(process)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

// deleteProcess deletes the given process and all associated data from the
// database and removes all associated message files from the message store.
func deleteProcess(process db.Process) error {
	deleted, err := messagestore.DeleteProcess(process.XdomeaID)
	if err != nil {
		return err
	} else if !deleted {
		return fmt.Errorf("could not delete process", process.XdomeaID)
	}
	return nil
}
