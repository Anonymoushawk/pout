package socket

import (
	"encoding/json"

	"github.com/codeuk/pout/client/cmd/system"
)

// `ClientSocket.ExtractProcesses` retrieves the list of processes running on the local machine and returns it in a JSON format.
func (client *ClientSocket) ExtractProcesses() ([]byte, error) {
	// Retrieve the list of processes.
	processes := system.ListProcesses()

	// Encode the list of processes as JSON.
	jsonData, err := json.Marshal(processes)
	if err != nil {
		// If an error occurs while encoding the list of processes, write the error message to the server.
		return []byte{}, err
	}

	// Send the encoded JSON data to the server.
	return jsonData, nil
}
