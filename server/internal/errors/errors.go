// Package errors provides functions to deal with runtime errors.
//
// We distinguish between unexpected errors and expected errors.
//
// Unexpected errors should be thrown using `panic`. `HandlePanic` is used to
// recover from unexpected errors and insert a processing error into the
// database to be shown to administrators.
//
// Expected errors should be propagated via return values. Calling functions can
// complement the error's data and insert a processing error into the database
// using `AddProcessingErrorWithData`. Calling functions can decide to abort
// after adding a processing error or to proceed.
package errors

import (
	"context"
	"fmt"
	"lath/xman/internal/auth"
	"lath/xman/internal/db"
	"lath/xman/internal/mail"
	"log"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

// FromError create a minimal processing error from a standard return error.
func FromError(title string, err error) db.ProcessingError {
	return db.ProcessingError{
		Title: title,
		Info:  err.Error(),
	}
}

// AddProcessingErrorWithData inserts a new processing error into the database.
//
// The inserted error is based on `err`, which can be a standard return error or
// a processing error. `data` is used to complement `err`. Fields of `data` take
// precedence over fields of `err`.
//
// If `err` is not a ProcessingError, `data` should at least populate the field
// Title. Failing to do so will be treated as application error.
func AddProcessingErrorWithData(err error, data db.ProcessingError) {
	e := withData(err, data)
	AddProcessingError(e)
}

// AddProcessingError inserts a new processing error into the database.
func AddProcessingError(e db.ProcessingError) {
	e.CreatedAt = time.Now()
	e.Stack = string(debug.Stack())
	e = augmentProcessingError(e)
	log.Printf("New processing error: %s\n", printableErrorInfo(e))
	db.InsertProcessingError(e)
	sendEmailNotifications(e)
}

func printableErrorInfo(e db.ProcessingError) string {
	i := e.Title
	if e.Agency != nil {
		i += "\n\tAgency: " + e.Agency.Name
	}
	if e.ProcessID != uuid.Nil {
		i += "\n\tProcess ID: " + e.ProcessID.String()
	}
	if e.Info != "" {
		i += "\n\t" + e.Info
	}
	return i
}

// HandlePanic checks for a panic in the current go routine and recovers from
// it.
//
// Call as `defer HandlePanic()` at the start of every go routine.
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
func HandlePanic(taskDescription string, data *db.ProcessingError, cb ...func(r interface{})) {
	if r := recover(); r != nil {
		e := db.ProcessingError{
			Title:     "Anwendungsfehler",
			ErrorType: "application-error",
			Info:      fmt.Sprintf("%s\n\n%v", taskDescription, r),
		}
		log.Printf("panic: %v\n", r)
		debug.PrintStack()
		if data != nil {
			e = withData(&e, *data)
		}
		// AddProcessingError (called below) could still panic. Prevent an
		// application crash and try to record panic.
		defer func() {
			if r2 := recover(); r2 != nil {
				e2 := db.ProcessingError{
					Title:     "Anwendungsfehler",
					ErrorType: "application-error",
					Info:      fmt.Sprintf("%s\n\n%v", "HandlePanic", r2),
					CreatedAt: time.Now(),
					Stack:     string(debug.Stack()),
				}
				db.InsertProcessingErrorFailsafe(e2)
			}
		}()
		AddProcessingError(e)
		for _, f := range cb {
			f(r)
		}
	}
}

// withData applies context information to a concrete error object.
func withData(err error, data db.ProcessingError) db.ProcessingError {
	e, ok := err.(*db.ProcessingError)
	if !ok {
		if data.Title == "" {
			e = &db.ProcessingError{
				Title:     "Anwendungsfehler",
				ErrorType: "application-error",
				Info: "Fehler ohne ausreichende Kontext-Informationen\n\n" +
					err.Error(),
			}
		} else {
			e = &db.ProcessingError{
				Info: err.Error(),
			}
		}
	}
	if data.Title != "" {
		e.Title = data.Title
	}
	if data.Info != "" {
		e.Info = data.Info
	}
	if data.ErrorType != "" {
		e.ErrorType = data.ErrorType
	}
	if e.Agency == nil && data.Agency != nil {
		e.Agency = data.Agency
	}
	if data.MessageType != "" {
		e.MessageType = data.MessageType
	}
	if e.ProcessID == uuid.Nil && data.ProcessID != uuid.Nil {
		e.ProcessID = data.ProcessID
	}
	if data.ProcessStep != "" {
		e.ProcessStep = data.ProcessStep
	}
	if data.TransferPath != "" {
		e.TransferPath = data.TransferPath
	}
	return *e
}

// augmentProcessingError fills in missing values of the processing error where
// possible.
func augmentProcessingError(e db.ProcessingError) db.ProcessingError {
	if e.Agency == nil && e.ProcessID != uuid.Nil {
		process, found := db.TryFindProcess(context.Background(), e.ProcessID)
		if found {
			e.Agency = &process.Agency
		}
	}
	if e.TransferPath == "" && e.ProcessID != uuid.Nil && e.MessageType != "" {
		message, found := db.TryFindMessage(context.Background(), e.ProcessID, e.MessageType)
		if found {
			e.TransferPath = message.TransferFile
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
	if e.MessageType == "" && e.ProcessStep != "" {
		switch e.ProcessStep {
		case db.ProcessStepReceive0501, db.ProcessStepAppraisal:
			e.MessageType = db.MessageType0501
		case db.ProcessStepReceive0503, db.ProcessStepFormatVerification:
			e.MessageType = db.MessageType0503
		case db.ProcessStepReceive0505:
			e.MessageType = db.MessageType0505
		}
	}
	return e
}

// sendEmailNotifications sends a notification for the processing error to all
// administrators that have subscribed for error notifications.
func sendEmailNotifications(e db.ProcessingError) {
	users := auth.ListUsers()
	for _, user := range users {
		if user.Permissions.Admin {
			preferences := db.FindUserPreferencesWithDefault(context.Background(), user.ID)
			if preferences.ErrorEmailNotifications {
				mailAddr, err := auth.GetMailAddress(user.ID)
				errorData := db.ProcessingError{
					Title:     "Fehler beim Versenden einer E-Mail-Benachrichtigung",
					CreatedAt: time.Now(),
					Stack:     string(debug.Stack()),
				}
				if err != nil {
					errorData.Info = err.Error()
					db.InsertProcessingErrorFailsafe(errorData)
				} else {
					err = mail.SendMailProcessingError(mailAddr, e)
					if err != nil {
						errorData.Info = err.Error()
						db.InsertProcessingErrorFailsafe(errorData)
					}
				}
			}
		}
	}
}
