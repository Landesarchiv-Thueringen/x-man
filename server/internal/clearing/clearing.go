package clearing

import (
	"fmt"
	"lath/xman/internal/db"
	"lath/xman/internal/messagestore"
)

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
