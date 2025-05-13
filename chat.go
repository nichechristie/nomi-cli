package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type ChatRequest struct {
	MessageText string `json:"messageText"`
}

type Message struct {
	UUID string `json:"uuid"`
	Text string `json:"text"`
	Sent string `json:"sent"`
}

type ChatResponse struct {
	SentMessage  Message `json:"sentMessage"`
	ReplyMessage Message `json:"replyMessage"`
}

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorBlue   = "\033[34m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

// clearScreen clears the terminal screen and attempts to clear the scrollback buffer.
func clearScreen() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		// First attempt using ANSI escape codes
		fmt.Print("\033[H\033[2J\033[3J\033c")

		// Fallback to tput if available
		cmd := exec.Command("tput", "reset")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

// findNomiByName retrieves the UUID of a Nomi by its name.
func findNomiByName(name string) (string, error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s/nomis", baseURL) // Use dynamic baseURL

	// Fetch the list of Nomis
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error fetching Nomis: %s", resp.Status)
	}

	var result struct {
		Nomis []struct {
			UUID string `json:"uuid"`
			Name string `json:"name"`
		} `json:"nomis"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	// Search for the Nomi by name
	for _, nomi := range result.Nomis {
		if strings.EqualFold(nomi.Name, name) { // Case-insensitive comparison
			return nomi.UUID, nil
		}
	}

	return "", fmt.Errorf("no Nomi found with the name: %s", name)
}

// spinner displays a spinning wheel animation while waiting for a response.
func spinner(stopChan chan bool) {
	chars := []string{"-", "\\", "|", "/"} // Simple classic spinner
	for {
		select {
		case <-stopChan:
			return
		default:
			for _, char := range chars {
				select {
				case <-stopChan:
					return
				default:
					fmt.Printf("\r%s%s%s", colorCyan, char, colorReset)
					time.Sleep(100 * time.Millisecond) // Slightly slower rotation
				}
			}
		}
	}
}

var chatCmd = &cobra.Command{
	Use:   "chat [id]",
	Short: "Start a live chat session with a specific Nomi",
	Args:  cobra.ExactArgs(1), // Requires exactly one argument: the Nomi Name
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		startChat(name)
	},
}
