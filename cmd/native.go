package cmd

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"procguard/internal/config"
	"procguard/internal/webblocklist"
	"time"

	"github.com/spf13/cobra"
)

// NativeHostCmd is the hidden command that handles native messaging.
var NativeHostCmd = &cobra.Command{
	Use:    "native-host",
	Short:  "Runs the native messaging host for browser communication.",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logDir, err := config.GetAppDataDir()
		if err != nil {
			return fmt.Errorf("failed to get app data dir: %w", err)
		}
		logFile, err := os.OpenFile(filepath.Join(logDir, "native-host.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		defer logFile.Close()

		logToFile(logFile, "Native messaging host started.")

		reader := bufio.NewReader(os.Stdin)

		for {
			// Read the 4-byte message length
			lengthBytes := make([]byte, 4)
			_, err := io.ReadFull(reader, lengthBytes)
			if err != nil {
				if err == io.EOF {
					logToFile(logFile, "Stdin closed, exiting native host.")
					break // Exit loop cleanly
				}
				logToFile(logFile, fmt.Sprintf("Error reading message length: %v", err))
				return err
			}

			// Decode the length (assuming native endianness, which is typical)
			length := binary.LittleEndian.Uint32(lengthBytes)

			// Read the message body
			body := make([]byte, length)
			_, err = io.ReadFull(reader, body)
			if err != nil {
				logToFile(logFile, fmt.Sprintf("Error reading message body: %v", err))
				return err
			}

			logToFile(logFile, fmt.Sprintf("Received message: %s", string(body)))

			// Process the message
			var request struct {
				Command string `json:"command"`
			}
			if err := json.Unmarshal(body, &request); err != nil {
				logToFile(logFile, fmt.Sprintf("Error unmarshalling request: %v", err))
				continue // Don't crash
			}

			// Handle commands
			if request.Command == "getBlocklist" {
				logToFile(logFile, "Received getBlocklist command.")
				domains, err := webblocklist.Load()
				if err != nil {
					logToFile(logFile, fmt.Sprintf("Error loading webblocklist: %v", err))
					// Send an empty list back on error
					domains = []string{}
				}

				response := map[string]interface{}{
					"type":    "blocklist",
					"domains": domains,
				}

				if err := sendMessage(response); err != nil {
					logToFile(logFile, fmt.Sprintf("Error sending blocklist: %v", err))
					return err
				}
				logToFile(logFile, "Successfully sent blocklist to extension.")
			} else {
				logToFile(logFile, fmt.Sprintf("Received unknown command: %s", request.Command))
			}
		}

		return nil
	},
}

// sendMessage formats and sends a message to the browser.
func sendMessage(messageData interface{}) error {
	msgBytes, err := json.Marshal(messageData)
	if err != nil {
		return fmt.Errorf("failed to marshal response message: %w", err)
	}

	// Write the 4-byte length prefix
	if err := binary.Write(os.Stdout, binary.LittleEndian, uint32(len(msgBytes))); err != nil {
		return fmt.Errorf("failed to write message length: %w", err)
	}

	// Write the message body
	_, err = os.Stdout.Write(msgBytes)
	return err
}

func logToFile(logFile *os.File, message string) {
	fmt.Fprintf(logFile, "[%s] %s\n", time.Now().Format(time.RFC3339), message)
}
