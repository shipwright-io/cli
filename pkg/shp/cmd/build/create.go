package build

import (
	"errors"
	"fmt"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/resource"
)

// CreateCommand contains data input from user
type CreateCommand struct {
	url      string
	strategy string
	name     string

	image string

	build *buildv1alpha1.Build

	cmd *cobra.Command
}

func createCmd() runner.SubCommand {
	createCommand := &CreateCommand{
		cmd: &cobra.Command{
			Use:   "create <name> [flags]",
			Short: "Create Build",
			Args:  cobra.ExactArgs(1),
		},
	}

	createCommand.cmd.Flags().StringVarP(&createCommand.image, "output-image", "i", "", "Output image created by build")
	createCommand.cmd.Flags().StringVarP(&createCommand.strategy, "strategy", "", "buildah", "Build strategy")
	createCommand.cmd.Flags().StringVarP(&createCommand.image, "source-url", "u", "", "Source URL to run the build from")

	createCommand.cmd.MarkFlagRequired("source-url")

	return createCommand
}

// Cmd returns cobra Command object of the create subcommand
func (c *CreateCommand) Cmd() *cobra.Command {
	return c.cmd
}

// Complete fills internal subcommand structure for future work with user input
func (c *CreateCommand) Complete(params *params.Params, args []string) error {
	c.name = args[0]

	return nil
}

func (c *CreateCommand) initializeBuild() {
	strategyKind := buildv1alpha1.ClusterBuildStrategyKind

	c.build = &buildv1alpha1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.name,
		},
		Spec: buildv1alpha1.BuildSpec{
			Strategy: &buildv1alpha1.Strategy{
				Name: c.strategy,
				Kind: &strategyKind,
			},
			Source: buildv1alpha1.Source{
				URL: c.url,
			},
		},
	}

	if c.image != "" {
		c.build.Spec.Output = buildv1alpha1.Image{
			Image: c.image,
		}
	}
}

// Validate is used for user input validation of flags and other data
func (c *CreateCommand) Validate() error {
	if c.strategy != "buildah" {
		return errors.New("incorrect strategy, must be 'buildah'")
	}

	return nil
}

// Run contains main logic of the create subcommand
func (sc *CreateCommand) Run(params *params.Params, io *genericclioptions.IOStreams) error {
	sc.initializeBuild()
	buildResource := resource.GetBuildResource(params)

	if err := buildResource.Create(sc.cmd.Context(), sc.name, sc.build); err != nil {
		return err
	}

	fmt.Fprintf(io.Out, "Build created %q\n", sc.name)
	return nil
}
