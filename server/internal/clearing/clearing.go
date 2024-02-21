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
func Resolve(processingError db.ProcessingError, resolution db.ProcessingErrorResolution) error {
	switch resolution {
	case db.ErrorResolutionReimportMessage:
		deleted, err := messagestore.DeleteMessage(*processingError.MessageID, true)
		if err != nil {
			return err
		} else if !deleted {
			return fmt.Errorf("failed to delete message %v", processingError.MessageID)
		}
	case db.ErrorResolutionDeleteMessage:
		deleted, err := messagestore.DeleteMessage(*processingError.MessageID, false)
		if err != nil {
			return err
		} else if !deleted {
			return fmt.Errorf("failed to delete message %v", processingError.MessageID)
		}
	default:
		return fmt.Errorf("unknown resolution: %s", resolution)
	}
	processingError.Resolved = true
	processingError.Resolution = resolution
	return db.UpdateProcessingError(processingError)
}
