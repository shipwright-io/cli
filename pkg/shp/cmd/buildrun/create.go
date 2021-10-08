package buildrun

import (
	"fmt"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/flags"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/util"
)

// CreateCommand reprents the build's create subcommand.
type CreateCommand struct {
	cmd *cobra.Command // cobra command instance

	name         string                      // buildrun name
	buildRunSpec *buildv1alpha1.BuildRunSpec // stores command-line flags
}

const buildRunCreateLongDesc = `
Creates a new BuildRun instance using the given name, and requires --buildref-name to
find the Build object. Example:

	$ shp buildrun create my-app-build --buildref-name="..."
`

// Cmd returns cobra.Command object of the create sub-command.
func (c *CreateCommand) Cmd() *cobra.Command {
	return c.cmd
}

// Complete checks if the arguments is informing the BuildRun name.
func (c *CreateCommand) Complete(params *params.Params, io *genericclioptions.IOStreams, args []string) error {
	switch len(args) {
	case 1:
		c.name = args[0]
	default:
		return fmt.Errorf("wrong amount of arguments, expected only one")
	}
	return nil
}

// Validate makes sure a name is informed.
func (c *CreateCommand) Validate() error {
	if c.name == "" {
		return fmt.Errorf("name is not informed")
	}
	return nil
}

// Run executes the creation of BuildRun object.
func (c *CreateCommand) Run(params *params.Params, ioStreams *genericclioptions.IOStreams) error {
	br := &buildv1alpha1.BuildRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.name,
		},
		Spec: *c.buildRunSpec,
	}

	envs, err := c.cmd.Flags().GetStringArray("env")
	if err != nil {
		return err
	}
	br.Spec.Env = append(br.Spec.Env, util.StringSliceToEnvVarSlice(envs)...)

	flags.SanitizeBuildRunSpec(&br.Spec)

	clientset, err := params.ShipwrightClientSet()
	if err != nil {
		return err
	}
	if _, err = clientset.ShipwrightV1alpha1().BuildRuns(params.Namespace()).Create(c.cmd.Context(), br, metav1.CreateOptions{}); err != nil {
		return err
	}
	fmt.Fprintf(ioStreams.Out, "BuildRun created %q for Build %q\n", c.name, br.Spec.BuildRef.Name)
	return nil
}

// createCmd instantiate a new CreateCommand, by wiring it as a cobra.Command and registering the
// flags and marking flags required.
func createCmd() runner.SubCommand {
	cmd := &cobra.Command{
		Use:   "create <name> [flags]",
		Short: "Creates a BuildRun instance.",
		Long:  buildRunCreateLongDesc,
	}

	// instantiating command-line flags, using an actual BuildRunSpec object to receive the flags
	// issued on command-line, also marking flags as required
	buildRunSpecFlags := flags.BuildRunSpecFromFlags(cmd.Flags())
	if err := cmd.MarkFlagRequired(flags.BuildrefNameFlag); err != nil {
		panic(err)
	}

	return &CreateCommand{
		cmd:          cmd,
		buildRunSpec: buildRunSpecFlags,
	}
}
