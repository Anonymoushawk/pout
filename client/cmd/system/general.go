package system

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	wapi "github.com/iamacarpet/go-win64api"
	so "github.com/iamacarpet/go-win64api/shared"
)

// `Process` represents the attributes of a running process on the machine.
type Process struct {
	PID    string `json:"process_id"`
	Name   string `json:"executable_name"`
	Type   string `json:"session_type"`
	Memory string `json:"mem_usage"`
}

// `SystemTable` represents a database of information gathered from the system.
type SystemTable struct {
	Registry RegistryTable      `json:"registry_information"`
	Hardware so.Hardware        `json:"hardware_information"`
	OS       so.OperatingSystem `json:"operating_system_information"`
	Memory   so.Memory          `json:"memory_information"`
	Disk     []so.Disk          `json:"disk_information"`
	Network  []so.Network       `json:"network_information"`
}

// `MetaData` represents a database of the local system information.
type MetaData struct {
	// The JSON values of this struct will have the match the ones server side,
	// as they are sent and read directory into the client's MetaData struct there.
	IP            string      `json:"ip_address"`
	Name          string      `json:"machine_name"`
	Arch          string      `json:"architecture"`
	Access        string      `json:"access_privileges"`
	System        SystemTable `json:"system_information,omitempty"`
	Processes     []Process   `json:"running_processes,omitempty"`
	EncryptionKey []byte      `json:"encryption_key,omitempty"`
}

// `GetSystemInformation` returns the information gathered from the wapi package.
// The wapi part of this function isn't currently working, as it's returning empty data.
func GetSystemInformation() SystemTable {
	var system SystemTable

	// Get valuable information from the Windows registry.
	system.Registry = GetRegistryInformation()

	if GetPrivilegeLevel() == "Admin" {
		// This information will only be available if the client is running with admin privileges.
		hardware, os, memory, disk, network, err := wapi.GetSystemProfile()
		if err != nil {
			return system
		}

		system.Hardware = hardware
		system.OS = os
		system.Memory = memory
		system.Disk = disk
		system.Network = network
	}

	return system
}

// `ListProcesses` retrieves a list of currently running processes on the local machine
// and their attributes using the tasklist command and parsing the output.
func ListProcesses() []Process {
	var processes []Process

	output, err := exec.Command("tasklist").Output()
	if err != nil {
		return processes
	}

	lines := strings.Split(string(output), "\n")[3:] // Only parse after the first 3 lines.

	for _, line := range lines {
		fields := strings.Fields(line)

		// Parse and validated the processes information from the line.
		if len(fields) > 3 {
			name := fields[0]
			pid := strings.Replace(fields[1], ",", "", -1)

			if strings.Contains(name, ".") {
				processes = append(processes, Process{
					PID:    pid,
					Name:   name,
					Type:   fields[2],
					Memory: fields[4],
				})
			}
		}
	}

	return processes
}

// `GetPrivilegeLevel` returns whether the current process has admin privileges or not.
// This isn't the most optimal solution, a better way would be by checking the current process
// privileges using the Win32 API, but there were issues using the required functions and variables in GoLang.
func GetPrivilegeLevel() string {
	var adminStr = "Unknown"

	// Use the ExecuteShellCommand function to run a command line script.
	out, err := ExecuteShellCommand("net session >nul 2>&1 && echo true || echo false")
	if err == nil {
		// Parse the command output and use the value to determine the current access privileges.
		admin := strings.Contains(string(out), "true")
		if admin {
			adminStr = "Admin"
		} else {
			adminStr = "User"
		}
	}

	return adminStr
}

// `GetArchitecture` returns the architecture of the current computer.
func GetArchitecture() string {
	// Use the builtin runtime architecture command to get the raw architecture string.

	return runtime.GOOS
}

// `GetHostname` returns the hostname of the current computer.
// If an error occurs while retrieving the hostname, an empty string is returned.
func GetHostname() string {
	// Use the os.Hostname() function to retrieve the hostname of the computer.
	hostname, err := os.Hostname()
	if err != nil {
		return "Unknown"
	}

	return hostname
}
