package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version string

// Command returns Version subcommand of Shipwright CLI
// for retrieving the shp version
func Command() *cobra.Command {
	command := &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "version",
		Annotations: map[string]string{
			"commandType": "main",
		},
		Run: func(cmd *cobra.Command, args []string) {
			if version == "" {
				version = "development"
			}

			fmt.Printf("version: %s\n", version)
		},
	}
	return command
}
