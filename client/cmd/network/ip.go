package network

import (
	"io/ioutil"
	"net/http"
)

var (
	// API to get the external IPV4 address from.
	// This can be changed if the formatting of the IP address returned is the same.
	// String form: http://api.ipify.org
	IP_API = []byte{104, 116, 116, 112, 58, 47, 47, 97, 112, 105, 46, 105, 112, 105, 102, 121, 46, 111, 114, 103}

	// IP API request counter.
	// This is used for checking if the MAX_REQ_ATTEMPTS has been reached.
	RequestsToAPI int
)

// `GetIPAddress` returns the external IP address of the current computer.
// The function makes a request to the "http://api.ipify.org" API to retrieve the IP address.
// If an error occurs during the retrieval process, an empty string is returned.
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

	// Read the response body to retrieve the IP address as a string.
	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if RequestsToAPI >= MAX_REQ_ATTEMPTS {
			return "Unknown"
		} else {
			return GetIPAddress()
		}
	}

	// Return the IP address as a string.
	return string(ip)
}

// Maximum retry attempts before returning an empty string from `GetIPAddress`.
// This is in place to deter constant retries incase the API we're calling is down
// or the request is blocked, as it could be flagged.
const MAX_REQ_ATTEMPTS = 2
