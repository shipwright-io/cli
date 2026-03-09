package clusterbuildstrategy

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/params"
)

// Command represents "shp clusterbuildstrategy".
func Command(p *params.Params, ioStreams *genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clusterbuildstrategy",
		Aliases: []string{"cbs"},
		Short:   "Manage cluster-scoped BuildStrategies",
		Annotations: map[string]string{
			"commandType": "main",
		},
	}

	cmd.AddCommand(
		runner.NewRunner(p, ioStreams, listCmd()).Cmd(),
		runner.NewRunner(p, ioStreams, deleteCmd()).Cmd(),
	)

	return cmd
}
