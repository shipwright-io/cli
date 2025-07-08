package buildstrategy

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/params"
)

// ListCommand contains data input from user for list sub-command
type ListCommand struct {
	cmd      *cobra.Command
	noHeader bool
}

func listCmd() runner.SubCommand {
	c := &ListCommand{
		cmd: &cobra.Command{
			Use:   "list [flags]",
			Short: "List BuildStrategies in the current namespace",
		},
	}
	c.cmd.Flags().BoolVar(&c.noHeader, "no-header", false, "Do not print the table header")
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
	k8s, err := p.ClientSet()
	if err != nil {
		return err
	}

	ns := p.Namespace()
	if _, err = k8s.CoreV1().Namespaces().Get(c.cmd.Context(), ns, metav1.GetOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			fmt.Fprintf(io.Out, "Namespace '%s' not found.\n", ns)
			return nil
		}
		return err
	}

	list, err := cs.ShipwrightV1beta1().BuildStrategies(ns).List(c.cmd.Context(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	if len(list.Items) == 0 {
		fmt.Fprintf(io.Out, "No BuildStrategies found in namespace '%s'.\n", ns)
		return nil
	}

	now := time.Now()
	for _, bs := range list.Items {
		age := duration.ShortHumanDuration(now.Sub(bs.CreationTimestamp.Time))
		fmt.Fprintf(w, "%s\t%s\n", bs.Name, age)
	}
	return w.Flush()
}
