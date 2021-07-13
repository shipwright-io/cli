package build

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/cli/pkg/shp/cmd/buildrun"
	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/flags"
	"github.com/shipwright-io/cli/pkg/shp/params"
)

// RunCommand represents the `build run` sub-command, which creates a unique BuildRun instance to run
// the build process, informed via arguments.
type RunCommand struct {
	cmd *cobra.Command

	buildName    string
	buildRunSpec *buildv1alpha1.BuildRunSpec
	buildRunOpts *flags.BuildRunOpts
}

const buildRunLongDesc = `
Creates a unique BuildRun instance for the given Build, which starts the build
process orchestrated by the Shipwright build controller. For example:

	$ shp build run my-app
`

// Cmd returns cobra.Command object of the create sub-command.
func (r *RunCommand) Cmd() *cobra.Command {
	return r.cmd
}

// Complete picks the build resource name from arguments, and instantiate additional components.
func (r *RunCommand) Complete(params *params.Params, args []string) error {
	switch len(args) {
	case 1:
		r.buildName = args[0]

	default:
		return errors.New("Build name is not informed")
	}

	// overwriting build-ref name to use what's on arguments
	return r.Cmd().Flags().Set(flags.BuildrefNameFlag, r.buildName)
}

// Validate the user must inform the build resource name.
func (r *RunCommand) Validate() error {
	if r.buildName == "" {
		return fmt.Errorf("name is not informed")
	}
	return nil
}

// Run creates a BuildRun resource based on Build's name informed on arguments.
func (r *RunCommand) Run(params *params.Params, ioStreams *genericclioptions.IOStreams) error {
	return buildrun.CreateBuildRun(
		r.cmd.Context(),
		"",
		params,
		ioStreams,
		r.buildRunSpec,
		r.buildRunOpts,
	)
}

// runCmd instantiate the "build run" sub-command using common BuildRun flags.
func runCmd() runner.SubCommand {
	cmd := &cobra.Command{
		Use:   "run <name>",
		Short: "Start a build specified by 'name'",
		Long:  buildRunLongDesc,
	}

	runCommand := &RunCommand{
		cmd:          cmd,
		buildRunSpec: flags.BuildRunSpecFromFlags(cmd.Flags()),
		buildRunOpts: flags.BuildRunOptsFromFlags(cmd.Flags()),
	}

	return runCommand
}
