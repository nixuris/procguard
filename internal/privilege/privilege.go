//go:build !windows

package privilege

// RunAsAdmin is a no-op on non-Windows platforms.
func RunAsAdmin(command string, args ...string) error {
	// On Linux, this would be implemented using pkexec or gksudo.
	// For now, it does nothing.
	return nil
}
