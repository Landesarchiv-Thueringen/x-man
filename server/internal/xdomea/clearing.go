package xdomea

import (
	"fmt"
	"lath/xman/internal/db"
)

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
