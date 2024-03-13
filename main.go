//go:build prod

// main is a special name declaring an executable rather than a library
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gen2brain/beeep"
)

type Meeting struct {
	Topic string `json:"topic`
	Link  string `json:"link"`
	When  string `json:"when"`
}

// entry point for the executable
func main() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("Error finding config directory:", err)
		return
	}
	file, err := os.Open(filepath.Join(configDir, "meetings.json"))

	if err != nil {
		fmt.Println("Error while opening JSON file:", err)
		return
	}
	defer file.Close()

	var meetings []Meeting
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&meetings)
	if err != nil {
		fmt.Println("Error decoding JSON file:", err)
		return
	}

	now := time.Now()

	for {
		now = time.Now()
		hour, minute := now.Hour(), now.Minute()

		for _, meeting := range meetings {
			fmt.Println(meeting.Topic)
			mt, err := time.Parse("15:04", meeting.When)
			if err != nil {
				fmt.Println("Error parsing time:", err)
				continue
			}

			mh, mm := mt.Hour(), mt.Minute()
			if hour == mh && minute == mm {
				err := beeep.Notify("GoMeet", "Your meeting link is set to launch shortly", "assets/information.png")
				if err != nil {
					fmt.Println("Error sending notification:", err)
					return
				}

				err = openBrowser(meeting.Link)
				// retry to open it if you failed
				if err != nil {
					continue
				}
				// do not launch it again if you succeeded
				time.Sleep(time.Minute)
			}
		}

		// sleep but wake up before minute ends
		time.Sleep(20 * time.Second)
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
