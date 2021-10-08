package buildrun

import (
	"context"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/shipwright-io/cli/pkg/shp/cmd/types"
)

var (
	buildRunCmdLongDescription = templates.LongDesc(``)
	buildRunCmdExamples        = templates.Examples(``)
)

// Command creates the BuildRun sub-command for managing Shipwright BuildRuns
func Command(ctx context.Context, ioStreams *genericclioptions.IOStreams, clients *types.ClientSets) *cobra.Command {
	command := &cobra.Command{
		Use:     "buildrun",
		Aliases: []string{"br"},
		Short:   "Manage BuildRuns",
		Long:    buildRunCmdLongDescription,
		Example: buildRunCmdExamples,
		Annotations: map[string]string{
			"commandType": "main",
		},
	}

	// TODO: add support for `update` and `get` commands
	command.AddCommand(
		NewBuildRunCreateCmd(ctx, ioStreams, clients),
		NewBuildRunDeleteCmd(ctx, ioStreams, clients),
		NewBuildRunCancelCmd(ctx, ioStreams, clients),
		NewBuildRunListCmd(ctx, ioStreams, clients),
		NewBuildRunLogsCmd(ctx, ioStreams, clients),
	)
	return command
}
