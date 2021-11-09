package buildrun

import (
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/params"
)

// Command represents "shp buildrun" sub-command.
func Command(p *params.Params, ioStreams *genericclioptions.IOStreams) *cobra.Command {
	command := &cobra.Command{
		Use:     "buildrun",
		Aliases: []string{"br"},
		Short:   "Manage BuildRuns",
		Annotations: map[string]string{
			"commandType": "main",
		},
	}

	// TODO: add support for `update` and `get` commands
	command.AddCommand(
		runner.NewRunner(p, ioStreams, listCmd()).Cmd(),
		runner.NewRunner(p, ioStreams, logsCmd()).Cmd(),
		runner.NewRunner(p, ioStreams, createCmd()).Cmd(),
		runner.NewRunner(p, ioStreams, cancelCmd()).Cmd(),
		runner.NewRunner(p, ioStreams, deleteCmd()).Cmd(),
	)
	return command
}
