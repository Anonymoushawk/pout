package system

import (
	"os/exec"
)

// `ExecuteShellCommand` executes a shell command and returns its output as a byte slice.
func ExecuteShellCommand(command string) ([]byte, error) {
	// Create a new command using the appropriate shell command for the current operating system.
	shell := "cmd"
	args := []string{"/C", command}
	cmd := exec.Command(shell, args...)

	// Run the command and retrieve its combined output.
	output, err := cmd.CombinedOutput()

	return output, err
}
