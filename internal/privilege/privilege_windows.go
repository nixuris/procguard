//go:build windows

package privilege

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/windows"
)

// RunAsAdmin executes a command with elevated (administrator) privileges.
func RunAsAdmin(command string, args ...string) error {
	verb := "runas"
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	argsStr := strings.Join(args, " ")

	verbPtr, err := windows.UTF16PtrFromString(verb)
	if err != nil {
		return err
	}
	exePtr, err := windows.UTF16PtrFromString(exe)
	if err != nil {
		return err
	}
	argsPtr, err := windows.UTF16PtrFromString(argsStr)
	if err != nil {
		return err
	}

	var showCmd int32 = windows.SW_SHOWNORMAL // or SW_HIDE to hide the window

	// ShellExecuteW is used to request elevation via the UAC prompt.
	err = windows.ShellExecute(0, verbPtr, exePtr, argsPtr, nil, showCmd)
	if err != nil {
		return fmt.Errorf("ShellExecute failed: %w", err)
	}

	return nil
}
