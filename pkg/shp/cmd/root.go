package cmd

import (
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd/build"
	"github.com/shipwright-io/cli/pkg/shp/params"
)

var rootCmd = &cobra.Command{
	Use:   "shp [command] [resource] [flags]",
	Short: "Command-line client for Shipwright's Build API.",
}

// NewCmdSHP create a new SHP root command, linking together all sub-commands organized by groups.
func NewCmdSHP(ioStreams genericclioptions.IOStreams) *cobra.Command {
	p := params.NewParams()
	p.AddFlags(rootCmd.PersistentFlags())

	rootCmd.AddCommand(build.Command(p))

	return rootCmd
}
