package archive

import (
	"context"
	"encoding/json"
	"lath/xman/internal/db"
)

const ProtocolFilename = "xman_protocol.json"

func GenerateProtocol(process db.SubmissionProcess) string {
	processStateBytes, err := json.MarshalIndent(process.ProcessState, "", " ")
	if err != nil {
		panic(err)
	}
	protocol := string(processStateBytes)
	processingErrors := db.FindProcessingErrorsForProcess(context.Background(), process.ProcessID)
	if len(processingErrors) > 0 {
		errorsBytes, err := json.MarshalIndent(processingErrors, "", " ")
		if err != nil {
			panic(err)
		}
		protocol += "\n" + string(errorsBytes)
	}
	return protocol
}
