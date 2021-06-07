package build

import (
	"fmt"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/flags"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/resource"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// CreateCommand contains data input from user
type CreateCommand struct {
	cmd *cobra.Command // cobra command instance

	name      string                   // build resource's name
	buildSpec *buildv1alpha1.BuildSpec // stores command-line flags
}

const buildCreateLongDesc = `
Creates a new Build instance using the first argument as its name. For example:

	$ shp build create my-app --source-url="..." --output-image="..."
`

// Cmd returns cobra.Command object of the create subcommand.
func (c *CreateCommand) Cmd() *cobra.Command {
	return c.cmd
}

// Complete fills internal subcommand structure for future work with user input
func (c *CreateCommand) Complete(params *params.Params, args []string) error {
	switch len(args) {
	case 1:
		c.name = args[0]
	default:
		return fmt.Errorf("one argument is expected")
	}
	return nil
}

// Validate is used for user input validation of flags and other data.
func (c *CreateCommand) Validate() error {
	if c.name == "" {
		return fmt.Errorf("name must be provided")
	}
	return nil
}

// Run executes the creation of a new Build instance using flags to fill up the details.
func (c *CreateCommand) Run(params *params.Params, io *genericclioptions.IOStreams) error {
	b := &buildv1alpha1.Build{Spec: *c.buildSpec}
	flags.SanitizeBuildSpec(&b.Spec)

	buildResource := resource.GetBuildResource(params)
	if err := buildResource.Create(c.cmd.Context(), c.name, b); err != nil {
		return err
	}
	fmt.Fprintf(io.Out, "Created build %q\n", c.name)
	return nil
}

// createCmd instantiate the "build create" subcommand.
func createCmd() runner.SubCommand {
	cmd := &cobra.Command{
		Use:   "create <name> [flags]",
		Short: "Create Build",
		Long:  buildCreateLongDesc,
	}

	// instantiating command-line flags and the build-spec structure which receives the informed flag
	// values, also marking certain flags as mandatory
	buildSpecFlags := flags.BuildSpecFromFlags(cmd.Flags())
	if err := cmd.MarkFlagRequired(flags.SourceURLFlag); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(flags.OutputImageFlag); err != nil {
		panic(err)
	}

	return &CreateCommand{
		cmd:       cmd,
		buildSpec: buildSpecFlags,
	}
}
