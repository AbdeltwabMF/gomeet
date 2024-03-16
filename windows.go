//go:build windows
// +build windows

package main

import (
	"github.com/go-toast/toast"
)

// Sends a notification with the specified meeting topic and URL,
// allowing users to directly open the meeting from the notification
func notifyMeeting(topic string, url string) error {
	notification := toast.Notification{
		AppID:    "gomeet",
		Title:    "Join Meeting: " + topic,
		Message:  "Click to join the meeting now.",
		Actions:  []toast.Action{{Type: "protocol", Label: "Join", Arguments: url}},
		Duration: toast.Long,
	}

	return notification.Push()
}
