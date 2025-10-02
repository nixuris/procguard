//go:build windows

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"procguard/internal/config"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// HandleFirstRunSetup checks if the application is installed. If not, it performs
// the installation and relaunches the application. It returns true if setup was
// performed and the current process should exit.
func HandleFirstRunSetup() (bool, error) {
	// 1. Get paths
	currentExePath, err := os.Executable()
	if err != nil {
		return false, fmt.Errorf("could not get current executable path: %w", err)
	}

	installDir, err := config.GetAppDataDir()
	if err != nil {
		return false, fmt.Errorf("could not get app data directory: %w", err)
	}
	destExePath := filepath.Join(installDir, filepath.Base(currentExePath))

	// 2. Check if we are already running from the install directory.
	if strings.EqualFold(currentExePath, destExePath) {
		// Already installed, do nothing.
		return false, nil
	}

	// --- Not installed, perform setup ---
	fmt.Println("Performing first-time setup...")

	// 3. Create directory and copy executable (GetAppDataDir already creates the dir)
	if err := copyFile(currentExePath, destExePath); err != nil {
		return false, fmt.Errorf("failed to copy executable: %w", err)
	}

	// 4. Create the Native Messaging Host manifest
	manifestPath := filepath.Join(installDir, nativeHostName+".json")
	manifest := map[string]interface{}{
		"name":        nativeHostName,
		"description": "ProcGuard Native Messaging Host",
		"path":        destExePath,
		"type":        "stdio",
		"allowed_origins": []string{
			"chrome-extension://__EXTENSION_ID__/", // Placeholder!
		},
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return false, fmt.Errorf("could not create manifest JSON: %w", err)
	}
	if err := os.WriteFile(manifestPath, manifestBytes, 0644); err != nil {
		return false, fmt.Errorf("could not write manifest file: %w", err)
	}

	// 5. Create the registry key
	regKeyPath := `SOFTWARE\Google\Chrome\NativeMessagingHosts\` + nativeHostName
	key, _, err := registry.CreateKey(registry.CURRENT_USER, regKeyPath, registry.SET_VALUE)
	if err != nil {
		return false, fmt.Errorf("could not create registry key: %w", err)
	}
	defer key.Close()
	if err := key.SetStringValue("", manifestPath); err != nil {
		return false, fmt.Errorf("could not set registry key value: %w", err)
	}

	// 6. Relaunch the application from the new location
	cmd := exec.Command(destExePath)
	if err := cmd.Start(); err != nil {
		return false, fmt.Errorf("failed to relaunch application from new location: %w", err)
	}

	fmt.Println("Installation complete. Relaunching...")
	// Tell the calling process to exit
	return true, nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	return err
}
