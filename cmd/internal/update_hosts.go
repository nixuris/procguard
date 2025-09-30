package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"procguard/internal/hosts"

	"github.com/spf13/cobra"
)

// UpdateHostsCmd is a hidden command that updates the hosts file.
var UpdateHostsCmd = &cobra.Command{
	Use:    "internal-update-hosts [file]",
	Short:  "Internal command to update the hosts file",
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("could not read domains file: %w", err)
		}
		defer os.Remove(filePath) // Clean up the temporary file

		var domains []string
		if err := json.Unmarshal(content, &domains); err != nil {
			return fmt.Errorf("could not unmarshal domains: %w", err)
		}

		return hosts.Update(domains)
	},
}
