package buildrun

import (
	"fmt"

	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/params"
)

// DeleteCommand contains data input from user for delete sub-command
type DeleteCommand struct {
	cmd *cobra.Command

	name string
}

func deleteCmd() runner.SubCommand {
	return &DeleteCommand{
		cmd: &cobra.Command{
			Use:   "delete <name>",
			Short: "Delete BuildRun",
			Args:  cobra.ExactArgs(1),
		},
	}
}

// Cmd returns cobra command object
func (c *DeleteCommand) Cmd() *cobra.Command {
	return c.cmd
}

// Complete fills in data provided by user
func (c *DeleteCommand) Complete(_ *params.Params, _ *genericclioptions.IOStreams, args []string) error {
	c.name = args[0]

	return nil
}

// Validate validates data input by user
func (c *DeleteCommand) Validate() error {
	return nil
}

// Run executes delete sub-command logic
func (c *DeleteCommand) Run(params *params.Params, ioStreams *genericclioptions.IOStreams) error {
	clientset, err := params.ShipwrightClientSet()
	if err != nil {
		return err
	}

	if err = clientset.ShipwrightV1beta1().BuildRuns(params.Namespace()).Delete(c.cmd.Context(), c.name, metav1.DeleteOptions{}); err != nil {
		return err
	}

	fmt.Fprintf(ioStreams.Out, "BuildRun deleted '%v'\n", c.name)

	return nil
}
