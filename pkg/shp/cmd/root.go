package cmd

import (
	"github.com/spf13/cobra"
)

// rootCmd root cobra command.
var rootCmd = &cobra.Command{
	Use:   "shp [command] [resource] [flags]",
	Short: "Command-line client for Shipwright Build.",
}

// NewCmd ...
func NewCmd() *cobra.Command {
	return rootCmd
}
