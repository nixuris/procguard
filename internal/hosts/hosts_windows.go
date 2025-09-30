//go:build windows

package hosts

import (
	"os"
	"path/filepath"
)

// GetPath returns the path to the hosts file on Windows.
func GetPath() (string, error) {
	return filepath.Join(os.Getenv("SystemRoot"), "System32", "drivers", "etc", "hosts"), nil
}
