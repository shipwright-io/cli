package build // nolint:revive

import (
	"fmt"

	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/flags"
	"github.com/shipwright-io/cli/pkg/shp/params"
)

// UpdateCommand contains data provided by user to update an existing Build
type UpdateCommand struct {
	cmd *cobra.Command

	name         string
	buildSpec    *buildv1beta1.BuildSpec
	dockerfile   *string
	builderImage *string
}

func updateCmd() runner.SubCommand {
	cmd := &cobra.Command{
		Use:   "update <name> [flags]",
		Short: "Update an existing Build",
		Args:  cobra.ExactArgs(1),
	}

	buildSpecFlags, dockerfileFlag, builderImageFlag := flags.BuildSpecFromFlags(cmd.Flags())

	return &UpdateCommand{
		cmd:          cmd,
		buildSpec:    buildSpecFlags,
		dockerfile:   dockerfileFlag,
		builderImage: builderImageFlag,
	}
}

// Cmd returns cobra command object
func (c *UpdateCommand) Cmd() *cobra.Command {
	return c.cmd
}

// Complete fills object with user input data
func (c *UpdateCommand) Complete(_ *params.Params, _ *genericclioptions.IOStreams, args []string) error {
	c.name = args[0]
	return nil
}

// Validate checks user input data
func (c *UpdateCommand) Validate() error {
	if c.name == "" {
		return fmt.Errorf("name must be provided")
	}
	return nil
}
