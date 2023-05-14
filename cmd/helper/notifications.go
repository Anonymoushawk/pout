package helper

import (
	"fmt"

	toast "gopkg.in/toast.v1"
)

// `ServerStartedNotification` pushes a Windows toast notification for the server starting.
func ServerStartedNotification(port string) error {
	notification := toast.Notification{
		AppID:   APP_NAME,
		Title:   "Server Started!",
		Message: fmt.Sprintf("Pout is listening for connections! (port: %s)", port),
	}

	return notification.Push()
}

// `NewClientNotification` pushes a Windows toast notification for each new client.
func NewClientNotification(rawAddr string) error {
	notification := toast.Notification{
		AppID:   APP_NAME,
		Title:   "New Connection!",
		Message: fmt.Sprintf("A new client has connected to Pout: (%s)", rawAddr),
	}

	return notification.Push()
}

// Toast notification variables.
const (
	APP_NAME = "Pout C2"
	// Icon = "icon.png"
)
