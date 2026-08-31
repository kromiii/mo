//go:build nogui || (!darwin && !windows && !linux) || !cgo

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var guiCmd = &cobra.Command{
	Use:   "gui [file ...]",
	Short: "Launch mo desktop GUI application (not supported in this build)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("desktop GUI is not supported in this build (requires CGO and GUI dependencies)")
	},
}

func init() {
	rootCmd.AddCommand(guiCmd)
}

func runDesktopGUI(port int, args []string) error {
	return fmt.Errorf("desktop GUI is not supported in this build (requires CGO and GUI dependencies)")
}

func isGUIAvailable() bool {
	return false
}
