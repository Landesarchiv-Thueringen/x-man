// The package clearing provides methods to handle processing errors.
//
// The basic idea is to pass an error of type db.ProcessingError up the chain
// and to finally call HandleError on it. This allows the caller act on an error
// while also saving the error to the database and inform administrators.
//
// HandleError should be called by the highest-level function that still needs
// to know about the error.
package xdomea

import (
	"context"
	"fmt"
	"lath/xman/internal/auth"
	"lath/xman/internal/db"
	"lath/xman/internal/mail"
	"log"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"
)

func CreateProcessingErrorPanic(info map[string]any) {
	var b strings.Builder
	for key, value := range info {
		fmt.Fprintf(&b, "%s: %v\n", key, value)
	}
	fmt.Fprintf(&b, "\n%s\n", debug.Stack())

	HandleError(db.ProcessingError{
		Type:           db.ProcessingErrorPanic,
		Description:    fmt.Sprintf("Anwendungsfehler"),
		AdditionalInfo: b.String(),
	})
}

// HandlePanic checks for a panic in the current go routine and recovers from
// it.
//
// It should be called as `defer HandlePanic()` at the start of a go routine or
// at the point from which one out you want to recover from panics.
//
// If there is a panic, it prints the stack trace to the console and creates a
// processing error that will be shown on the clearing page to administrators.
//
// For functions invoked by API requests you usually don't need HandlePanic
// since panics are handled by GIN middleware. HandlePanic is useful when API
// requests start go routines or when go routines are started by the server
// without an API request, e.g., as reaction on a newly found message file or
// timed invocation of a routine. In these situations, it should be used when
// there is a chance of panics caused by runtime conditions like a broken
// network connection or a failure of an external service. For panics indicating
// fatal misconfiguration, handling is not required.
func HandlePanic(taskDescription string) {
	if r := recover(); r != nil {
		log.Printf("panic: %v\n", r)
		debug.PrintStack()
		info := map[string]any{
			"Fehler":  r,
			"Aufgabe": taskDescription,
		}
		CreateProcessingErrorPanic(info)
	}
}

// HandleError handles an error object.
//
// If it is a ProcessingError, it adds it to the database and sends notification
// e-mail to subscribed clearing personnel. It fills some missing fields if
// sufficient information is provided.
//
// If it is any other error, it panics.
func HandleError(err error) {
	if err == nil {
		return
	} else if e, ok := err.(db.ProcessingError); ok {
		e = augmentProcessingError(e)
		if e.ProcessID != uuid.Nil {
			log.Printf("Processing error for submission process %s: %s\n", e.ProcessID.String(), e.Description)
		} else {
			log.Printf("Processing error: %s\n", e.Description)
		}
		db.InsertProcessingError(e)
		sendEmailNotifications(e)
	} else {
		panic(fmt.Sprintf("unhandled error: %v", err))
	}
}

// Resolve resolves the given processing error with the given resolution.
//
// If successful, it marks the processing error as resolved. Otherwise, it
// returns an error.
func Resolve(processingError db.ProcessingError, resolution db.ProcessingErrorResolution) {
	switch resolution {
	case db.ErrorResolutionMarkSolved:
		// Do nothing
	case db.ErrorResolutionReimportMessage:
		DeleteMessage(processingError.ProcessID, processingError.MessageType, true)
	case db.ErrorResolutionDeleteMessage:
		DeleteMessage(processingError.ProcessID, processingError.MessageType, false)
	default:
		panic(fmt.Sprintf("unknown resolution: %s", resolution))
	}
	db.UpdateProcessingErrorResolve(processingError, resolution)
}

func sendEmailNotifications(e db.ProcessingError) {
	users := auth.ListUsers()
	for _, user := range users {
		if user.Permissions.Admin {
			preferences := db.FindUserPreferences(context.Background(), user.ID)
			if preferences.ErrorEmailNotifications {
				mailAddr := auth.GetMailAddress(user.ID)
				mail.SendMailProcessingError(mailAddr, e)
			}
		}
	}
}

func augmentProcessingError(e db.ProcessingError) db.ProcessingError {
	e.CreatedAt = time.Now()
	if e.Agency == nil && e.ProcessID != uuid.Nil {
		process, found := db.FindProcess(context.Background(), e.ProcessID)
		if found {
			e.Agency = &process.Agency
		}
	}
	if e.TransferPath == "" && e.ProcessID != uuid.Nil && e.MessageType != "" {
		message, found := db.FindMessage(context.Background(), e.ProcessID, e.MessageType)
		if found {
			e.TransferPath = message.TransferDirPath
		}
	}
	if e.ProcessStep == "" && e.MessageType != "" {
		switch e.MessageType {
		case db.MessageType0501:
			e.ProcessStep = db.ProcessStepReceive0501
		case db.MessageType0503:
			e.ProcessStep = db.ProcessStepReceive0503
		case db.MessageType0505:
			e.ProcessStep = db.ProcessStepReceive0505
		}
	}
	return e
}
