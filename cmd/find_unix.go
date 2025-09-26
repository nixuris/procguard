//go:build !windows

package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// This is the non-windows version of the find command.

var (
	since string
	until string
)

func init() {
	findCmd.Flags().StringVar(&since, "since", "", "Show logs since a specific time")
	findCmd.Flags().StringVar(&until, "until", "", "Show logs until a specific time")
	rootCmd.AddCommand(findCmd)
}

var findCmd = &cobra.Command{
	Use:   "find <name>",
	Short: "Find log lines by program name",
	Args:  cobra.ExactArgs(1),
	Run:   runFind,
}

func runFind(cmd *cobra.Command, args []string) {
	query := strings.ToLower(args[0])

	var sinceTime, untilTime time.Time
	var err error

	if since != "" {
		sinceTime, _ = parseTime(since)
	}
	if until != "" {
		untilTime, _ = parseTime(until)
	}

	cacheDir, _ := os.UserCacheDir()
	logFile := filepath.Join(cacheDir, "procguard", "events.log")

	file, err := os.Open(logFile)
	if err != nil {
		fmt.Println("[]") // Return empty JSON array on error
		return
	}
	defer file.Close()

	matchingLines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " | ")
		if len(parts) < 4 {
			continue
		}

		logTime, err := time.ParseInLocation("2006-01-02 15:04:05", parts[0], time.Local)
		if err != nil {
			continue
		}

		if !sinceTime.IsZero() && logTime.Before(sinceTime) {
			continue
		}
		if !untilTime.IsZero() && logTime.After(untilTime) {
			continue
		}

		exe := strings.ToLower(parts[1])
		parentExe := strings.ToLower(parts[3])
		if strings.Contains(exe, query) || strings.Contains(parentExe, query) {
			matchingLines = append(matchingLines, line)
		}
	}

	// Format for simple, non-windows UI
	var jsonLines [][]string
	for _, l := range matchingLines {
		if l != "" {
			jsonLines = append(jsonLines, strings.Split(l, " | "))
		}
	}

	jsonBytes, _ := json.Marshal(jsonLines)
	fmt.Println(string(jsonBytes))
}

func parseTime(input string) (time.Time, error) {
	now := time.Now()
	lowerInput := strings.ToLower(strings.TrimSpace(input))

	switch lowerInput {
	case "now":
		return now, nil
	case "1 hour ago":
		return now.Add(-1 * time.Hour), nil
	case "24 hours ago":
		return now.Add(-24 * time.Hour), nil
	case "7 days ago":
		return now.AddDate(0, 0, -7), nil
	}

	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
	}

	for _, layout := range layouts {
		parsedTime, err := time.ParseInLocation(layout, input, time.Local)
		if err == nil {
			return parsedTime, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse time: %s", input)
}
