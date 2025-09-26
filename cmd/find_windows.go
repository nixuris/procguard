//go:build windows

package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"procguard/cmd/windows"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"
)

// This is the Windows-specific version of the find command.

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

// RichLogEntry contains the full, rich information for a single log event.
type RichLogEntry struct {
	AppInfo *windows.AppInfo `json:"appInfo"`
	RawLog  string           `json:"rawLog"`
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

	// Enrich the results with Windows-specific info
	results := []*RichLogEntry{}
	procCache := make(map[string]*windows.AppInfo)

	for _, line := range matchingLines {
		parts := strings.Split(line, " | ")
		exeName := parts[1]

		appInfo, exists := procCache[exeName]
		if !exists {
			path, err := findExePathForName(exeName)
			if err != nil {
				appInfo = &windows.AppInfo{Name: exeName, ExeName: exeName}
			} else {
				appInfo = windows.GetAppInfo(path)
			}
			procCache[exeName] = appInfo
		}

		results = append(results, &RichLogEntry{
			AppInfo: appInfo,
			RawLog:  line,
		})
	}

	jsonBytes, _ := json.Marshal(results)
	fmt.Println(string(jsonBytes))
}

func findExePathForName(exeName string) (string, error) {
	procs, err := process.Processes()
	if err != nil {
		return "", err
	}
	for _, p := range procs {
		name, _ := p.Name()
		if strings.EqualFold(name, exeName) {
			return p.Exe()
		}
	}
	return "", fmt.Errorf("process not found: %s", exeName)
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
