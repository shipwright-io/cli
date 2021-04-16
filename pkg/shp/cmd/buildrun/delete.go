package buildrun

import (
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"

	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/resource"
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
func (c *DeleteCommand) Complete(params *params.Params, args []string) error {
	c.name = args[0]

	return nil
}

// Validate validates data input by user
func (c *DeleteCommand) Validate() error {
	return nil
}

// Run executes delete sub-command logic
func (c *DeleteCommand) Run(params *params.Params, ioStreams *genericclioptions.IOStreams) error {
	brr := resource.GetBuildRunResource(params)

	if err := brr.Delete(c.cmd.Context(), c.name); err != nil {
		return err
	}

	klog.Infof("Deleted buildrun %q", c.name)

	return nil
}
