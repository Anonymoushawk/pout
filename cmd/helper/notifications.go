package helper

import (
	"fmt"

	toast "gopkg.in/toast.v1"
)

// `ServerStartedNotification` pushes a Windows toast notification for the server starting.
func ServerStartedNotification(port string) error {
	notification := toast.Notification{
		AppID:   "Pout C2",
		Title:   "Server Started!",
		Message: fmt.Sprintf("Pout is listening for connections! (port: %s)", port),
	}

	return notification.Push()
}

// `NewClientNotification` pushes a Windows toast notification for each new client.
func NewClientNotification(rawAddr string) error {
	notification := toast.Notification{
		AppID:   "Pout C2",
		Title:   "New Connection!",
		Message: fmt.Sprintf("A new client has connected to Pout: (%s)", rawAddr),
	}

	return notification.Push()
}

// Windows toast notification variables.
const (
	App = "Pout C2"
	// Icon = "icon.png"
)
