package cmd

import (
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd/build"
	"github.com/shipwright-io/cli/pkg/shp/cmd/buildrun"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/suggestion"
)

var rootCmd = &cobra.Command{
	Use:           "shp [command] [resource] [flags]",
	Short:         "Command-line client for Shipwright's Build API.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// NewCmdSHP create a new SHP root command, linking together all sub-commands organized by groups.
func NewCmdSHP(ioStreams *genericclioptions.IOStreams) *cobra.Command {
	p := params.NewParams()
	p.AddFlags(rootCmd.PersistentFlags())

	rootCmd.AddCommand(build.Command(p, ioStreams))
	rootCmd.AddCommand(buildrun.Command(p, ioStreams))

	visitCommands(rootCmd, reconfigureCommandWithSubcommand)

	return rootCmd
}

func reconfigureCommandWithSubcommand(cmd *cobra.Command) {
	if len(cmd.Commands()) == 0 {
		return
	}

	if cmd.Args == nil {
		cmd.Args = cobra.ArbitraryArgs
	}

	if cmd.RunE == nil {
		cmd.RunE = suggestion.SubcommandsRequiredWithSuggestions
	}
}

func visitCommands(cmd *cobra.Command, f func(*cobra.Command)) {
	f(cmd)
	for _, child := range cmd.Commands() {
		visitCommands(child, f)
	}
}
