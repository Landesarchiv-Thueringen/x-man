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
	var err error
	switch resolution {
	case db.ErrorResolutionMarkSolved:
		// Do nothing
	case db.ErrorResolutionReimportMessage:
		err = DeleteMessage(processingError.ProcessID, processingError.MessageType, true)
	case db.ErrorResolutionDeleteMessage:
		err = DeleteMessage(processingError.ProcessID, processingError.MessageType, false)
	case db.ErrorResolutionDeleteTransferFile:
		RemoveFileFromTransferDir(*processingError.Agency, processingError.TransferPath)
	default:
		panic(fmt.Sprintf("unknown resolution: %s", resolution))
	}
	if err != nil {
		panic(err)
	}
	db.UpdateProcessingErrorResolve(processingError, resolution)
}
