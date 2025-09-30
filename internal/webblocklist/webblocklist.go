package webblocklist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const blockListFile = "web_blocklist.json"

// Load reads the web blocklist file from the user's cache directory.
func Load() ([]string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	p := filepath.Join(cacheDir, "procguard", blockListFile)

	b, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return nil, nil // Return empty list if file doesn't exist
	}
	if err != nil {
		return nil, err
	}

	var list []string
	_ = json.Unmarshal(b, &list)

	// Normalize all entries to lowercase for case-insensitive comparison.
	for i := range list {
		list[i] = strings.ToLower(list[i])
	}
	return list, nil
}

// Save writes the given list of strings to the web blocklist file.
func Save(list []string) error {
	// Normalize all entries to lowercase to ensure consistency.
	for i := range list {
		list[i] = strings.ToLower(list[i])
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	procguardDir := filepath.Join(cacheDir, "procguard")
	if err := os.MkdirAll(procguardDir, 0755); err != nil {
		return err
	}
	p := filepath.Join(procguardDir, blockListFile)

	b, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(p, b, 0644)
}
