//go:build !windows

package cmd

// HandleFirstRunSetup is a no-op on non-Windows platforms.
func HandleFirstRunSetup() (bool, error) {
	// No setup is needed, proceed with normal execution.
	return false, nil
}
