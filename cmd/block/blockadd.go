package block

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	BlockAddCmd.Flags().Bool("json", false, "output json for gui")
}

// BlockAddCmd defines the command for adding a program to the blocklist.
var BlockAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add program to block-list (OS-agnostic)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		promptAndVerifyPassword()
		// All entries are stored in lowercase for case-insensitive matching.
		name := strings.ToLower(args[0])
		list, _ := LoadBlockList()

		// Check if the program is already in the blocklist.
		for _, v := range list {
			if v == name {
				isJSON, _ := cmd.Flags().GetBool("json")
				Reply(isJSON, "exists", name)
				return
			}
		}

		// Add the new program to the list and save it.
		list = append(list, name)
		if err := SaveBlockList(list); err != nil {
			fmt.Fprintln(os.Stderr, "save:", err)
			os.Exit(1)
		}

		// Respond to the user with the result of the operation.
		isJSON, _ := cmd.Flags().GetBool("json")
		Reply(isJSON, "added", name)
	},
}
