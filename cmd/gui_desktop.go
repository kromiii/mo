//go:build !nogui && (darwin || windows || linux) && cgo

package cmd

import (
	"github.com/k1LoW/mo/desktop"
	"github.com/spf13/cobra"
)

var guiCmd = &cobra.Command{
	Use:   "gui [file ...]",
	Short: "Launch mo desktop GUI application",
	Long:  "Launch mo as a native desktop application window.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return desktop.Run(port, args)
	},
}

func init() {
	guiCmd.Flags().IntVarP(&port, "port", "p", 6275, "Server port")
	rootCmd.AddCommand(guiCmd)
}

func runDesktopGUI(port int, args []string) error {
	return desktop.Run(port, args)
}

func isGUIAvailable() bool {
	return true
}
