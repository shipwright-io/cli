package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/shipwright-io/cli/pkg/shp/cmd/build"
	"github.com/shipwright-io/cli/pkg/shp/cmd/buildrun"
	"github.com/shipwright-io/cli/pkg/shp/cmd/types"
	"github.com/shipwright-io/cli/pkg/shp/suggestion"
)

var (
	rootCmdLongDescription = templates.LongDesc(`
		Shipwright Client

		Shipwright is an extensible framework for building container images on Kubernetes  
		which supports popular tools such as Kaniko, Cloud Native Buildpacks, Buildah, and more!

		Shipwright is based around four elements for each build:  
    		- Source code - “what” you are trying to build  
    		- Output image - “where” you are trying to deliver your application  
    		- Build strategy - “how” your application is assembled  
    		- Invocation - “when” you want to build your application

		For more information about Shipwright, see the documentation at https://shipwright.io/docs
	`)
	rootCmdExamples = templates.Examples(`
		Create a new build  
		$ shp build create my-build --source-url=https://github.com/shipwright-io/sample-go --source-context-dir=docker-build --output-image=my-image

		Get a list of existing builds  
		$ shp build list

		Run an existing build  
		$ shp build run my-build

		Cancel a running build  
		$ shp build cancel my-build

		Delete an existing build  
		$ shp build delete my-build
	`)
)
var rootCmd = &cobra.Command{
	Use:           "shp [resource] [command] [flags]",
	Short:         "Shipwright Client",
	Long:          rootCmdLongDescription,
	Example:       rootCmdExamples,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// NewCmdSHP create a new SHP root command, linking together all sub-commands organized by groups.
func NewRootCmd(ctx context.Context, ioStreams *genericclioptions.IOStreams, clients *types.ClientSets) *cobra.Command {

	//p.AddFlags(rootCmd.PersistentFlags())

	rootCmd.AddCommand(build.Command(ctx, ioStreams, clients))
	rootCmd.AddCommand(buildrun.Command(ctx, ioStreams, clients))

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
