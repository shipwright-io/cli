package build

import (
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/params"
)

// Command returns Build subcommand of Shipwright CLI
// for interaction with shipwright builds
func Command(p *params.Params, ioStreams *genericclioptions.IOStreams) *cobra.Command {
	command := &cobra.Command{
		Use:     "build",
		Aliases: []string{"bd"},
		Short:   "Manage Builds",
		Annotations: map[string]string{
			"commandType": "main",
		},
	}

	// TODO: add support for `update` and `get` commands
	command.AddCommand(
		runner.NewRunner(p, ioStreams, createCmd()).Cmd(),
		runner.NewRunner(p, ioStreams, listCmd()).Cmd(),
		runner.NewRunner(p, ioStreams, deleteCmd()).Cmd(),
		runner.NewRunner(p, ioStreams, runCmd()).Cmd(),
		runner.NewRunner(p, ioStreams, uploadCmd()).Cmd(),
	)
	return command
}
