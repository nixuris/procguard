package block

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	BlockRmCmd.Flags().Bool("json", false, "output json for gui")
}

// BlockRmCmd defines the command for removing a program from the blocklist.
var BlockRmCmd = &cobra.Command{
	Use:   "rm <exe>",
	Short: "Remove program from block-list",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		promptAndVerifyPassword()
		// All entries are stored in lowercase for case-insensitive matching.
		base := strings.ToLower(args[0])
		list, _ := LoadBlockList()

		// Find the index of the program to be removed.
		idx := slices.Index(list, base)
		if idx == -1 {
			isJSON, _ := cmd.Flags().GetBool("json")
			Reply(isJSON, "not found", base)
			return
		}

		// Remove the element from the list.
		list = slices.Delete(list, idx, idx+1)
		if err := SaveBlockList(list); err != nil {
			fmt.Fprintln(os.Stderr, "save:", err)
			os.Exit(1)
		}

		// Respond to the user with the result of the operation.
		isJSON, _ := cmd.Flags().GetBool("json")
		Reply(isJSON, "removed", base)
	},
}
