package build

import (
	"errors"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"

	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/params"
)

// DeleteCommand contains data provided by user to the delete subcommand
type DeleteCommand struct {
	name string

	cmd *cobra.Command
}

func deleteCmd() runner.SubCommand {
	deleteCommand := &DeleteCommand{
		cmd: &cobra.Command{
			Use:   "delete [flags] name",
			Short: "Delete Build",
		},
	}

	return deleteCommand
}

// Cmd returns cobra command object of the delete subcommand
func (c *DeleteCommand) Cmd() *cobra.Command {
	return c.cmd
}

// Complete fills DeleteSubCommand structure with data obtained from cobra command
func (c *DeleteCommand) Complete(params *params.Params, args []string) error {
	if len(args) == 0 {
		return errors.New("missing 'name' argument")
	}

	c.name = args[0]

	return nil
}

// Validate is used for validation of user input data
func (c *DeleteCommand) Validate() error {
	return nil
}

// Run contains main logic of delete subcommand
func (c *DeleteCommand) Run(params *params.Params) error {
	var b buildv1alpha1.Build

	if err := buildResource.Get(c.name, &b); err != nil {
		return err
	}

	return buildResource.Delete(c.name)
}
