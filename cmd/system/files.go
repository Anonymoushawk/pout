package system

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// `File` represents the attributes of a file on the host machine.
type File struct {
	Name string
	Path string
}

// `BasePath` derives the file base from the path and removes any whitespace from it.
//
//	"C:/Path/example file.exe" -> "examplefile.exe".
func BasePath(path string) string {
	// Get the file (including the extension) from the path.
	path = filepath.Base(path)

	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			// If the character is a space, drop it.
			return -1
		}
		// Otherwise, keep it in the string.
		return r
	}, path)
}

// `GetFileContent` reads the passed filepaths content and returns it as an array of bytes.
func GetFileContent(filePath string) ([]byte, error) {
	// Get the passed files content.
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

func GetFiles(folder string) (files []File) {
	// Scrape passed directory for filenames and filepaths.
	directory := UserPath + "\\" + folder
	fileScrape, _ := os.ReadDir(directory)

	// Loop through list of files and store the filename and path in a File struct.
	for _, file := range fileScrape {
		files = append(files, File{
			Name: file.Name(),
			Path: CleanPath(directory + "\\" + file.Name()),
		})
	}

	return files
}

func (file *File) Move(destination string) bool {
	// Move a copy of the file to the passed desination path.

	return CopyFileToDirectory(file.Path, destination)
}

func (file *File) WriteString(data string) bool {
	// Append the supplied data (string) to the files content (File).
	f, err := os.OpenFile(file.Name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return false
	}
	defer f.Close()

	str := data + "\n"
	_, err = f.WriteString(str)

	return err == nil
}

func (file *File) WriteJson(data interface{}) bool {
	// Append the supplied data to the files content.
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	jsonData = append(jsonData, []byte("\n")...)

	f, err := os.OpenFile(file.Path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return false
	}
	defer f.Close()

	_, err = f.Write(jsonData)

	return err == nil
}

func CopyFileToDirectory(pathSourceFile string, pathDestFile string) bool {
	// Copies the supplied source file to a destination directory.
	pathSourceFile = CleanPath(pathSourceFile)
	pathDestFile = CleanPath(pathDestFile)

	// Open the source file.
	dataSourceFile, err := os.Open(pathSourceFile)
	if err != nil {
		return false
	}
	defer dataSourceFile.Close()

	dataDestFile, err := os.Create(pathDestFile)
	if err != nil {
		return false
	}
	defer dataDestFile.Close()

	_, err = io.Copy(dataDestFile, dataSourceFile)

	return err == nil
}

func FileExists(filePath string) bool {
	// Check if a filepath exists on the machine.
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

func CleanPath(filePath string) string {
	// Make sure filepath slashes do not collide.
	return strings.ReplaceAll(filepath.Clean(filePath), `\`, `/`)
}

// Common local system paths (for data storage purposes).
var (
	UserPath   = os.Getenv("USERPROFILE")
	DataPath   = ".\\data\\"
	AssetsPath = ".\\assets\\"
)
