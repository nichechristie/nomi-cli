package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/chzyer/readline"
)

// startChat initiates a chat session with a Nomi by name
func startChat(name string) {
	// Ensure the screen is cleared when the program exits
	defer clearScreen()

	// Find the UUID for the given name
	nomiID, err := findNomiByName(name)
	if err != nil {
		fmt.Println(err)
		return
	}

	client := &http.Client{}
	url := fmt.Sprintf("%s/nomis/%s/chat", baseURL, nomiID)

	// Clear the terminal at the start of the chat
	clearScreen()

	fmt.Printf("\n%s=== Chat Session with %s ===%s\n", colorYellow, name, colorReset)
	fmt.Printf("%s• Type your message and press Enter to send\n", colorBlue)
	fmt.Printf("%s• Type 'exit' to end the session\n", colorBlue)
	fmt.Printf("%s• Use arrow keys to navigate within your text%s\n\n", colorBlue, colorReset)

	// Initialize readline with proper terminal settings
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          fmt.Sprintf("%sYou%s: ", colorGreen, colorReset),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		// Disable persistent history (in-memory history still works during the session)
		DisableAutoSaveHistory: true,
	})
	if err != nil {
		fmt.Printf("Error initializing input reader: %v\n", err)
		return
	}
	defer rl.Close()

	// Set auto-completion function if needed later
	// rl.Config.AutoComplete = completer

	for {
		input, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(input) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		// Check for exit command
		if strings.ToLower(strings.TrimSpace(input)) == "exit" {
			fmt.Println("Chat session ended.")
			break
		}

		// Prepare the request payload
		chatRequest := ChatRequest{MessageText: input}
		requestBody, err := json.Marshal(chatRequest)
		if err != nil {
			fmt.Println("Error encoding request body:", err)
			continue
		}

		// Create the HTTP request
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
		if err != nil {
			fmt.Println("Error creating request:", err)
			continue
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")

		// Start the spinner
		stopChan := make(chan bool)
		go spinner(stopChan)

		// Send the request
		resp, err := client.Do(req)

		// Stop the spinner
		close(stopChan)
		fmt.Print("\r") // Clear the spinner line

		if err != nil {
			fmt.Println("Error sending message:", err)
			continue
		}
		defer resp.Body.Close()

		// Check for successful response
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Error: %s\n", resp.Status)
			continue
		}

		// Decode the response
		var chatResponse ChatResponse
		if err := json.NewDecoder(resp.Body).Decode(&chatResponse); err != nil {
			fmt.Println("Error decoding response:", err)
			continue
		}

		// Display the reply
		fmt.Printf("%s%s%s: %s\n", colorBlue, name, colorReset, chatResponse.ReplyMessage.Text)
	}
}