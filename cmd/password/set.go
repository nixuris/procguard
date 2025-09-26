package password

import (
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"procguard/internal/auth"
	"procguard/internal/config"
)

func init() {
	PasswordCmd.AddCommand(setCmd)
}

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set or change the master password",
	Run:   runSet,
}

func runSet(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading config:", err)
		os.Exit(1)
	}

	// If a password already exists, require the old one first.
	if cfg.PasswordHash != "" {
		fmt.Print("Enter old password: ")
		oldPass, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Fprintln(os.Stderr, "\nError reading password:", err)
			os.Exit(1)
		}
		fmt.Println()

		if !auth.CheckPasswordHash(string(oldPass), cfg.PasswordHash) {
			fmt.Fprintln(os.Stderr, "Incorrect old password.")
			os.Exit(1)
		}
	}

	// Get the new password.
	fmt.Print("Enter new password: ")
	newPass, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Fprintln(os.Stderr, "\nError reading password:", err)
		os.Exit(1)
	}
	fmt.Println()

	// Confirm the new password.
	fmt.Print("Confirm new password: ")
	confirmPass, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Fprintln(os.Stderr, "\nError reading password:", err)
		os.Exit(1)
	}
	fmt.Println()

	if string(newPass) != string(confirmPass) {
		fmt.Fprintln(os.Stderr, "Passwords do not match.")
		os.Exit(1)
	}

	// Hash the new password and save it.
	hash, err := auth.HashPassword(string(newPass))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error hashing password:", err)
		os.Exit(1)
	}

	cfg.PasswordHash = hash
	if err := cfg.Save(); err != nil {
		fmt.Fprintln(os.Stderr, "Error saving configuration:", err)
		os.Exit(1)
	}

	fmt.Println("Password updated successfully.")
}