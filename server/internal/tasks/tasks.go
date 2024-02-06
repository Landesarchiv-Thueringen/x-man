// Package tasks calls reoccurring tasks in regular intervals.
package tasks

import (
	"fmt"
	"lath/xman/internal/db"
	"lath/xman/internal/messagestore"
	"log"
	"os"
	"strconv"
	"time"
)

// interval is the time to wait until task runs.
const interval = 1 * time.Hour

// const interval = 1 * time.Hour

// Init schedules regular execution for all tasks.
func Init() {
	go func() {
		for {
			cleanupArchivedProcesses()
			time.Sleep(interval)
		}
	}()
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
