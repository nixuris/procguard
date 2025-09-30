package cmd

import (
	"net/http"

	"procguard/cmd/block"
	"procguard/cmd/daemon"
	"procguard/cmd/gui"
	"procguard/cmd/internal"
	"procguard/cmd/uninstall"

	"github.com/spf13/cobra"
)

// HandleDefaultStartup implements the main startup logic for GUI mode.
func HandleDefaultStartup() {
	const defaultPort = "58141"
	guiAddress := "127.0.0.1:" + defaultPort
	guiUrl := "http://" + guiAddress

	// Check if a server is already running
	_, err := http.Get(guiUrl + "/ping")
	if err == nil {
		// Instance is already running. Just open the browser and exit.
		openBrowser(guiUrl)
		return
	}

	// No instance is running. This is the first instance.
	// Set up autostart for Windows if applicable.
	daemon.EnsureAutostartTask()

	// Start the daemon in the background
	go daemon.Start()

	// Open the browser
	openBrowser(guiUrl)

	// Start the web server (this is a blocking call)
	gui.StartWebServer(guiAddress)
}

var rootCmd = &cobra.Command{
	Use:   "procguard",
	Short: "Process monitor and control program",
}

func Execute() { cobra.CheckErr(rootCmd.Execute()) }

func init() {
	rootCmd.AddCommand(daemon.DaemonCmd)
	rootCmd.AddCommand(block.BlockCmd)
	rootCmd.AddCommand(gui.GuiCmd)
	rootCmd.AddCommand(uninstall.UninstallCmd)
	rootCmd.AddCommand(internal.UpdateHostsCmd)
}
