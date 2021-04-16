package build

import (
	"errors"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"

	"github.com/shipwright-io/cli/pkg/shp/cmd/buildrun"
	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/resource"
)

// RunCommand contains data input from user for run sub-command
type RunCommand struct {
	cmd *cobra.Command

	buildName string
}

// Cmd returns cobra command object
func (c *RunCommand) Cmd() *cobra.Command {
	return c.cmd
}

func runCmd() runner.SubCommand {
	runCommand := &RunCommand{
		cmd: &cobra.Command{
			Use:   "run <name>",
			Short: "Start a build specified by 'name'",
			Args:  cobra.ExactArgs(1),
		},
	}

	return runCommand
}

// Complete fills in data provided by user
func (c *RunCommand) Complete(params *params.Params, args []string) error {
	if len(args) < 1 {
		return errors.New("'name' argument is empty")
	}

	c.buildName = args[0]

	return nil
}

// Validate validates data input by user
func (c *RunCommand) Validate() error {
	return nil
}

// Run executes run sub-command logic
func (c *RunCommand) Run(params *params.Params, ioStreams *genericclioptions.IOStreams) error {
	br := resource.GetBuildResource(params)
	brr := resource.GetBuildRunResource(params)

	var build buildv1alpha1.Build
	if err := br.Get(c.cmd.Context(), c.buildName, &build); err != nil {
		return err
	}

	buildRun := buildrun.NewBuildRun(&build, build.Name)

	if err := brr.Create(c.cmd.Context(), "", buildRun); err != nil {
		return err
	}

	klog.Infof("Created buildrun %q", buildRun.Name)
	return nil
}
