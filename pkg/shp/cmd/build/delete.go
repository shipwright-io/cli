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
)

// DeleteCommand contains data provided by user to the delete subcommand
type DeleteCommand struct {
	name string

	cmd        *cobra.Command
	deleteRuns bool
}

func deleteCmd() runner.SubCommand {
	deleteCommand := &DeleteCommand{
		cmd: &cobra.Command{
			Use:   "delete <name> [flags]",
			Short: "Delete Build",
			Args:  cobra.ExactArgs(1),
		},
	}

	deleteCommand.cmd.Flags().BoolVarP(&deleteCommand.deleteRuns, "delete-runs", "r", false, "Also delete all of the buildruns")

	return deleteCommand
}

// Cmd returns cobra command object of the delete subcommand
func (c *DeleteCommand) Cmd() *cobra.Command {
	return c.cmd
}

// Complete fills DeleteSubCommand structure with data obtained from cobra command
func (c *DeleteCommand) Complete(p params.Interface, io *genericclioptions.IOStreams, args []string) error {
	if len(args) != 1 {
		return errors.New("build name is not informed")
	}
	c.name = args[0]

	return nil
}

// Validate is used for validation of user input data
func (c *DeleteCommand) Validate() error {
	return nil
}

// Run contains main logic of delete subcommand
func (c *DeleteCommand) Run(p params.Interface, io *genericclioptions.IOStreams) error {
	clientset, err := p.ShipwrightClientSet()
	if err != nil {
		return err
	}
	if err := clientset.ShipwrightV1alpha1().Builds(p.Namespace()).Delete(c.Cmd().Context(), c.name, metav1.DeleteOptions{}); err != nil {
		return err
	}

	if c.deleteRuns {
		var brList *buildv1alpha1.BuildRunList
		if brList, err = clientset.ShipwrightV1alpha1().BuildRuns(p.Namespace()).List(c.cmd.Context(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%v/name=%v", buildv1alpha1.BuildDomain, c.name),
		}); err != nil {
			return err
		}

		for _, buildrun := range brList.Items {
			if err := clientset.ShipwrightV1alpha1().BuildRuns(p.Namespace()).Delete(c.cmd.Context(), buildrun.Name, metav1.DeleteOptions{}); err != nil {
				fmt.Fprintf(io.ErrOut, "Error deleting BuildRun %q: %v\n", buildrun.Name, err)
			}
		}
	}

	fmt.Fprintf(io.Out, "Build deleted %q\n", c.name)

	return nil
}
