package build

import (
	"context"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/shipwright-io/cli/pkg/shp/cmd/types"
)

var (
	buildCmdLongDescription = templates.LongDesc(``)
	buildCmdExamples        = templates.Examples(``)
)

// Command creates the Build sub-command for managing Shipwright Builds
func Command(ctx context.Context, ioStreams *genericclioptions.IOStreams, clients *types.ClientSets) *cobra.Command {
	command := &cobra.Command{
		Use:     "build",
		Aliases: []string{"bd"},
		Short:   "Manage Builds",
		Long:    buildCmdLongDescription,
		Example: buildCmdExamples,
		Annotations: map[string]string{
			"commandType": "main",
		},
	}

	// TODO: add support for `update` and `get` commands
	command.AddCommand(
		NewBuildCreateCmd(ctx, ioStreams, clients),
		NewBuildDeleteCmd(ctx, ioStreams, clients),
		NewBuildListCmd(ctx, ioStreams, clients),
		NewBuildRunCmd(ctx, ioStreams, clients),
	)
	return command
}
