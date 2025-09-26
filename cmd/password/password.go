package password

import "github.com/spf13/cobra"

// PasswordCmd is the parent command for all password-related subcommands.
var PasswordCmd = &cobra.Command{
	Use:   "password",
	Short: "Manage the application password",
}
