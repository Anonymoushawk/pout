package helper

import (
	"fmt"
	"strings"
	"time"
)

// `Logs` represents a list of log entries.
type Logs struct {
	Entries []string
}

// `Logs.Add` appends the passed log to the Logs.Entries field.
func (logs *Logs) Add(log string) {
	// Current time prefix for each log.
	currentTime := time.Now().Format("03:04:05")

	logs.Entries = append(logs.Entries,
		fmt.Sprintf("[%s] %s", currentTime, strings.ReplaceAll(log, "\n", "")))
}

// `Logs.Clear` empties all logs from the Logs.Entries field.
func (logs *Logs) Clear() {
	logs.Entries = []string{}
}
