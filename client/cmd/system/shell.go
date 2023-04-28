package system

import "C"

import (
	"os/exec"
)

// `ExecuteShellCommand` executes a shell command and returns its output as a byte slice.
// The function uses the "cmd" command on Windows and "/bin/sh" on Unix-based systems.
// The output includes both standard output and standard error streams, unless hideErrors is passed.
// If an error occurs while executing the command, it is returned along with an empty byte slice.
func ExecuteShellCommand(command string) ([]byte, error) {
	// Create a new command using the appropriate shell command for the current operating system.
	var shell string
	var args []string

	shell = "cmd"
	args = []string{"/C", command}
	cmd := exec.Command(shell, args...)

	// Run the command and retrieve its combined output.
	output, err := cmd.CombinedOutput()

	return output, err
}
