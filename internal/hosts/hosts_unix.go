//go:build linux || darwin

package hosts

// GetPath returns the path to the hosts file on Unix-like systems.
func GetPath() (string, error) {
	return "/etc/hosts", nil
}
