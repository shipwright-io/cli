package clusterbuildstrategy

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/params"
)

type ListCommand struct {
	cmd      *cobra.Command
	noHeader bool
}

func listCmd() runner.SubCommand {
	c := &ListCommand{
		cmd: &cobra.Command{
			Use:   "list [flags]",
			Short: "List ClusterBuildStrategies",
		},
	}
	c.cmd.Flags().BoolVar(&c.noHeader, "no-header", false, "Omit table header")
	return c
}

// Cmd returns cobra command object
func (c *ListCommand) Cmd() *cobra.Command {
	return c.cmd
}

// Complete fills in data provided by user
func (c *ListCommand) Complete(_ *params.Params, _ *genericclioptions.IOStreams, _ []string) error {
	return nil
}

// Validate validates data input by user
func (c *ListCommand) Validate() error {
	return nil
}

// Run executes list sub-command logic
func (c *ListCommand) Run(p *params.Params, io *genericclioptions.IOStreams) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, '\t', 0)
	if !c.noHeader {
		fmt.Fprintln(w, "NAME\tAGE")
	}

	cs, err := p.ShipwrightClientSet()
	if err != nil {
		return err
	}
	list, err := cs.ShipwrightV1beta1().ClusterBuildStrategies().List(c.cmd.Context(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	if len(list.Items) == 0 {
		fmt.Fprintln(io.Out, "No ClusterBuildStrategies found.")
		return nil
	}

	now := time.Now()
	for _, cbs := range list.Items {
		age := duration.ShortHumanDuration(now.Sub(cbs.CreationTimestamp.Time))
		fmt.Fprintf(w, "%s\t%s\n", cbs.Name, age)
	}
	return w.Flush()
}
