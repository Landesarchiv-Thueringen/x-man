package archive

import (
	"encoding/json"
	"lath/xman/internal/db"
)

const ProtocolFilename = "xman_protocol.json"

func GenerateProtocol(process db.Process) string {
	processStateBytes, err := json.MarshalIndent(process.ProcessState, "", " ")
	if err != nil {
		panic(err)
	}
	protocol := string(processStateBytes)
	if len(process.ProcessingErrors) > 0 {
		errorsBytes, err := json.MarshalIndent(process.ProcessingErrors, "", " ")
		if err != nil {
			panic(err)
		}
		protocol += "\n" + string(errorsBytes)
	}
	return protocol
}
