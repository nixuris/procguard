package hosts

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
)

const (
	startMarker = "# BEGIN ProcGuard Block"
	endMarker   = "# END ProcGuard Block"
)

// Update writes the given domains to the hosts file.
func Update(domains []string) error {
	path, err := GetPath()
	if err != nil {
		return fmt.Errorf("could not get hosts file path: %w", err)
	}

	// Read the current content of the hosts file
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read hosts file: %w", err)
	}

	var newContent bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(content))
	inBlock := false

	// Copy existing content, but skip the old ProcGuard block
	for scanner.Scan() {
		line := scanner.Text()
		if line == startMarker {
			inBlock = true
			continue
		}
		if line == endMarker {
			inBlock = false
			continue
		}
		if !inBlock {
			newContent.WriteString(line)
			newContent.WriteString("\n")
		}
	}

	// Add the new ProcGuard block if there are domains to block
	if len(domains) > 0 {
		newContent.WriteString(startMarker)
		newContent.WriteString("\n")
		for _, domain := range domains {
			newContent.WriteString(fmt.Sprintf("127.0.0.1 %s\n", domain))
		}
		newContent.WriteString(endMarker)
		newContent.WriteString("\n")
	}

	// Write the new content back to the hosts file
	// This will require administrator privileges.
	return os.WriteFile(path, newContent.Bytes(), 0644)
}
