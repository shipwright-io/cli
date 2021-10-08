package build

import (
	"context"
	"errors"
	"fmt"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/shipwright-io/cli/pkg/shp/cmd/types"
)

var (
	// Long description for the "build delete" command
	buildeDeleteLongDescription = templates.LongDesc(`
		Deletes an existing Build
	`)

	// Examples for using the "build delete" command
	buildDeleteExamples = templates.Examples(`
		$ shp build delete my-build
	`)
)

// BuildDeleteOptions stores data passed to the command via command line flags
type BuildDeleteOptions struct {
	types.SharedOptions

	BuildName string

	DeleteRuns bool
}

// newBuildDeleteCmd creates the "build delete" command
func newBuildDeleteCmd(ctx context.Context, ioStreams *genericclioptions.IOStreams, clients *types.ClientSets, o *BuildDeleteOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <name> [flags]",
		Short:   "Delete an existing Build",
		Long:    buildeDeleteLongDescription,
		Example: buildDeleteExamples,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(args))
			cmdutil.CheckErr(o.Run())
		},
	}

	cmd.Flags().BoolVarP(&o.DeleteRuns, "delete-runs", "r", false, "Delete all BuildRuns associated with this Build. Defaults to false.")

	return cmd
}

// NewBuildDeleteCmd is a wrapper for newBuildDeleteCmd
func NewBuildDeleteCmd(ctx context.Context, ioStreams *genericclioptions.IOStreams, clients *types.ClientSets) *cobra.Command {
	o := &BuildDeleteOptions{
		SharedOptions: types.SharedOptions{
			Clients: clients,
			Context: ctx,
			Streams: ioStreams,
		},
	}

	return newBuildDeleteCmd(ctx, ioStreams, clients, o)
}

// Complete processes any data that is needed before Run executes
func (o *BuildDeleteOptions) Complete(args []string) error {
	// Guard against index out of bound errors
	if len(args) == 0 {
		return errors.New("argument list is empty")
	}

	o.BuildName = args[0]

	return nil
}

// Run executes the command logic
func (o *BuildDeleteOptions) Run() error {
	if err := o.Clients.ShipwrightClientSet.ShipwrightV1alpha1().Builds(o.Clients.Namespace).Delete(o.Context, o.BuildName, v1.DeleteOptions{}); err != nil {
		return fmt.Errorf("error occurred while attempting to delete Build %q: %v", o.BuildName, err)
	} else {
		fmt.Fprintf(o.Streams.Out, "Successfully deleted Build %q\n", o.BuildName)
	}

	if o.DeleteRuns {
		brList, err := o.Clients.ShipwrightClientSet.ShipwrightV1alpha1().BuildRuns(o.Clients.Namespace).List(o.Context, v1.ListOptions{
			LabelSelector: fmt.Sprintf("%v/name=%v", buildv1alpha1.BuildDomain, o.BuildName),
		})
		if err != nil {
			return err
		}

		for _, buildrun := range brList.Items {
			if err := o.Clients.ShipwrightClientSet.ShipwrightV1alpha1().BuildRuns(o.Clients.Namespace).Delete(o.Context, buildrun.Name, v1.DeleteOptions{}); err != nil {
				fmt.Fprintf(o.Streams.ErrOut, "error occurred while attempting to delete BuildRun %q: %v\n", buildrun.Name, err)
			} else {
				fmt.Fprintf(o.Streams.Out, "Successfully deleted BuildRun %q\n", buildrun.Name)
			}
		}
	}

	return nil
}
