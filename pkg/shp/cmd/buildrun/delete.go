package buildrun

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/shipwright-io/cli/pkg/shp/cmd/types"
)

var (
	// Long description for the "buildrun delete" command
	buildRunDeleteLongDescription = templates.LongDesc(`
		Deletes a BuildRun
	`)

	// Examples for using the "buildrun delete" command
	buildRunDeleteExamples = templates.Examples(`
		$ shp buildrun delete my-buildrun
	`)
)

// BuildRunDeleteOptions stores data passed to the command via command line flags
type BuildRunDeleteOptions struct {
	types.SharedOptions

	Name string
}

// newBuildRunDeleteCmd creates the "buildrun delete" command
func newBuildRunDeleteCmd(ctx context.Context, ioStreams *genericclioptions.IOStreams, clients *types.ClientSets, o *BuildRunDeleteOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <name>",
		Short:   "Delete an existing BuildRun",
		Long:    buildRunDeleteLongDescription,
		Example: buildRunDeleteExamples,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(args))
			cmdutil.CheckErr(o.Run())
		},
	}

	return cmd
}

// NewBuildRunDeleteCmd is a wrapper for newBuildRunDeleteCmd
func NewBuildRunDeleteCmd(ctx context.Context, ioStreams *genericclioptions.IOStreams, clients *types.ClientSets) *cobra.Command {
	o := &BuildRunDeleteOptions{
		SharedOptions: types.SharedOptions{
			Clients: clients,
			Context: ctx,
			Streams: ioStreams,
		},
	}

	return newBuildRunDeleteCmd(ctx, ioStreams, clients, o)
}

// Complete processes any data that is needed before Run executes
func (o *BuildRunDeleteOptions) Complete(args []string) error {
	// Guard against index out of bound errors
	if len(args) == 0 {
		return errors.New("argument list is empty")
	}

	o.Name = args[0]

	return nil
}

// Run executes the command logic
func (o *BuildRunDeleteOptions) Run() error {
	if err := o.Clients.ShipwrightClientSet.ShipwrightV1alpha1().BuildRuns(o.Clients.Namespace).Delete(o.Context, o.Name, metav1.DeleteOptions{}); err != nil {
		return err
	}

	fmt.Fprintf(o.Streams.Out, "BuildRun deleted '%v'\n", o.Name)

	return nil
}
