package buildstrategy

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/params"
)

// Command represents "shp buildstrategy".
func Command(p *params.Params, io *genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "buildstrategy",
		Aliases: []string{"bs"},
		Short:   "Manage namespaced BuildStrategies",
		Annotations: map[string]string{
			"commandType": "main",
		},
	}

	cmd.AddCommand(
		runner.NewRunner(p, io, listCmd()).Cmd(),
		runner.NewRunner(p, io, deleteCmd()).Cmd(),
	)

	return cmd
}
