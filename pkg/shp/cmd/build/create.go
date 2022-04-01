package build

import (
	"fmt"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/flags"
	"github.com/shipwright-io/cli/pkg/shp/params"
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
func (c *CreateCommand) Complete(params *params.Params, io *genericclioptions.IOStreams, args []string) error {
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
	b := &buildv1alpha1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.name,
		},
		Spec: *c.buildSpec,
	}

	flags.SanitizeBuildSpec(&b.Spec)

	// print warning with regards to source bundle image being used
	if b.Spec.Source.BundleContainer != nil && b.Spec.Source.BundleContainer.Image != "" {
		fmt.Fprintf(io.Out, "Build %q uses a source bundle image, which means source code will be transferred to a container registry. It is advised to use private images to ensure the security of the source code being uploaded.\n", c.name)
	}

	clientset, err := params.ShipwrightClientSet()
	if err != nil {
		return err
	}
	if _, err := clientset.ShipwrightV1alpha1().Builds(params.Namespace()).Create(c.cmd.Context(), b, metav1.CreateOptions{}); err != nil {
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
	if err := cmd.MarkFlagRequired(flags.OutputImageFlag); err != nil {
		panic(err)
	}

	return &CreateCommand{
		cmd:       cmd,
		buildSpec: buildSpecFlags,
	}
}
