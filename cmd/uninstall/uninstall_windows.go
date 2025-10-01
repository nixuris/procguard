//go:build windows

package uninstall

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"procguard/internal/blocklist"
	"strings"
)

const taskName = "ProcGuardDaemon"

func platformUninstall() error {
	// Unblock all files
	if err := unblockAll(); err != nil {
		return err
	}

	// Remove Task Scheduler task
	if err := removeTask(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not remove task scheduler entry: %v\n", err)
	}

	// Remove data files
	if err := removeDataFiles(); err != nil {
		return err
	}

	// Remove backup executable
	return removeBackup()
}

func unblockAll() error {
	list, err := blocklist.Load()
	if err != nil {
		return fmt.Errorf("could not load blocklist: %w", err)
	}

	for _, name := range list {
		// On Windows, unblocking means renaming the file.
		if strings.HasSuffix(name, ".blocked") {
			newName := strings.TrimSuffix(name, ".blocked")
			if err := os.Rename(name, newName); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not unblock %s: %v\n", name, err)
			}
		}
	}

	return nil
}

func removeTask() error {
	fmt.Println("Removing Task Scheduler task...")
	// The /f flag is to force the deletion.
	return exec.Command("schtasks", "/delete", "/tn", taskName, "/f").Run()
}

func removeDataFiles() error {
	fmt.Println("Removing data files...")
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	procguardDir := filepath.Join(cacheDir, "procguard")
	logsDir := filepath.Join(procguardDir, "logs")

	if err := os.RemoveAll(logsDir); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not remove logs directory: %v\n", err)
	}
	return os.RemoveAll(procguardDir)
}

func removeBackup() error {
	fmt.Println("Removing backup executable...")
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return fmt.Errorf("could not find LOCALAPPDATA directory")
	}
	return os.RemoveAll(filepath.Join(localAppData, "ProcGuard"))
}


