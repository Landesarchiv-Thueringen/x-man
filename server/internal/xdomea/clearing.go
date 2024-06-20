package xdomea

import (
	"fmt"
	"lath/xman/internal/db"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Resolve resolves the given processing error with the given resolution.
//
// If successful, it marks the processing error as resolved. Otherwise, it
// returns an error.
func Resolve(e db.ProcessingError, r db.ProcessingErrorResolution, user string) {
	var err error
	switch r {
	case db.ErrorResolutionMarkSolved:
		// Do nothing
	case db.ErrorResolutionMarkDone:
		err = db.UpdateProcessStepCompletion(e.ProcessID, e.ProcessStep, true, user)
	case db.ErrorResolutionReimportMessage:
		err = DeleteMessage(e.ProcessID, e.MessageType, true)
	case db.ErrorResolutionDeleteMessage:
		err = DeleteMessage(e.ProcessID, e.MessageType, false)
	case db.ErrorResolutionDeleteTransferFile:
		RemoveFileFromTransferDir(*e.Agency, e.TransferPath)
	case db.ErrorResolutionIgnoreTransferFile:
		for _, f := range e.Data.(primitive.A) {
			db.InsertTransferFile(e.Agency.ID, uuid.Nil, f.(string))
		}
	case db.ErrorResolutionDeleteTransferFiles:
		for _, f := range e.Data.(primitive.A) {
			RemoveFileFromTransferDir(*e.Agency, f.(string))
		}
	default:
		panic(fmt.Sprintf("unknown resolution: %s", r))
	}
	if err != nil {
		panic(err)
	}
	db.UpdateProcessingErrorResolve(e, r)
}
