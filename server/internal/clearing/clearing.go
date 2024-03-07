// The package clearing provides methods to handle processing errors.
//
// The basic idea is to pass an error of type db.ProcessingError up the chain
// and to finally call HandleError on it. This allows the caller act on an error
// while also saving the error to the database and inform administrators.
//
// HandleError should be called by the highest-level function that still needs
// to know about the error.
package clearing

import (
	"fmt"
	"lath/xman/internal/auth"
	"lath/xman/internal/db"
	"lath/xman/internal/mail"
	"lath/xman/internal/messagestore"
	"log"
)

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
		log.Printf("Processing error for message %s: %s\n", e.MessageID, e.Description)
		db.AddProcessingError(db.ProcessingError(e))
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
	case db.ErrorResolutionReimportMessage:
		messagestore.DeleteMessage(*processingError.MessageID, true)
	case db.ErrorResolutionDeleteMessage:
		messagestore.DeleteMessage(*processingError.MessageID, false)
	default:
		panic(fmt.Sprintf("unknown resolution: %s", resolution))
	}
	processingError, found := db.GetProcessingError(processingError.ID)
	if found {
		db.UpdateProcessingError(db.ProcessingError{
			ID:         processingError.ID,
			Resolved:   true,
			Resolution: resolution,
		})
	} else {
		// The processing error has already been deleted.
	}
}

func sendEmailNotifications(e db.ProcessingError) {
	users := auth.ListUsers()
	for _, user := range users {
		if user.Permissions.Admin {
			preferences := db.GetUserInformation(user.ID).Preferences
			if preferences.ErrorEmailNotifications {
				mailAddr := auth.GetMailAddress(user.ID)
				mail.SendMailProcessingError(mailAddr, e)
			}
		}
	}
}

func augmentProcessingError(e db.ProcessingError) db.ProcessingError {
	if e.Process == nil && e.ProcessID != nil {
		process, found := db.GetProcess(*e.ProcessID)
		if found {
			e.Process = &process
		}
	}
	if e.AgencyID == nil && e.Agency == nil {
		if e.Process != nil {
			e.AgencyID = &e.Process.AgencyID
			e.Agency = &e.Process.Agency
		}
	}
	if e.Message == nil && e.MessageID != nil {
		message, err := db.GetMessageByID(*e.MessageID)
		if err == nil {
			e.Message = &message
		}
	}
	if e.TransferPath == nil && e.Message != nil {
		e.TransferPath = &e.Message.TransferDirMessagePath
	}
	if e.Message != nil && e.Process != nil && e.ProcessStep == nil && e.ProcessStepID == nil {
		switch e.Message.MessageType.Code {
		case "0501":
			e.ProcessStep = &e.Process.ProcessState.Receive0501
		case "0503":
			e.ProcessStep = &e.Process.ProcessState.Receive0503
		case "0505":
			e.ProcessStep = &e.Process.ProcessState.Receive0505
		}
	}
	return e
}
