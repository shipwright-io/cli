package build

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/shipwright-io/cli/pkg/shp/cmd/types"
)

var (
	// Long description for the "build list" command
	buildListLongDescription = templates.LongDesc(`
		List existing Builds
	`)

	// Examples for using the "build list" command
	buildListExamples = templates.Examples(`
		$ shp build list
	`)
)

// BuildListOptions stores data passed to the command via command line flags
type BuildListOptions struct {
	types.SharedOptions

	BuildName string

	NoHeader bool
}

// newBuildListCmd creates the "build list" command
func newBuildListCmd(ctx context.Context, ioStreams *genericclioptions.IOStreams, clients *types.ClientSets, o *BuildListOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list [flags]",
		Short:   "List existing Builds",
		Long:    buildListLongDescription,
		Example: buildListExamples,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(args))
			cmdutil.CheckErr(o.Run())
		},
	}

	cmd.Flags().BoolVar(&o.NoHeader, "no-header", false, "Do not show columns header in list output")

	return cmd
}

// NewBuildListCmd is a wrapper for newBuildListCmd
func NewBuildListCmd(ctx context.Context, ioStreams *genericclioptions.IOStreams, clients *types.ClientSets) *cobra.Command {
	o := &BuildListOptions{
		SharedOptions: types.SharedOptions{
			Clients: clients,
			Context: ctx,
			Streams: ioStreams,
		},
	}

	return newBuildListCmd(ctx, ioStreams, clients, o)
}

// Complete processes any data that is needed before Run executes
func (o *BuildListOptions) Complete(args []string) error {
	return nil
}

// Run executes the command logic
func (o *BuildListOptions) Run() error {
	// TODO: Support multiple output formats here, not only tabwriter
	//       find out more in kubectl libraries and use them

	// Initialize tabwriter for command output
	writer := tabwriter.NewWriter(o.Streams.Out, 0, 8, 2, '\t', 0)
	columnNames := "NAME\tOUTPUT\tSTATUS"
	columnTemplate := "%s\t%s\t%s\n"

	buildList, err := o.Clients.ShipwrightClientSet.ShipwrightV1alpha1().Builds(o.Clients.Namespace).List(o.Context, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("an error occurred while attempting to get a list of existing builds: %v", err)
	}

	if !o.NoHeader {
		fmt.Fprintln(writer, columnNames)
	}

	for _, b := range buildList.Items {
		fmt.Fprintf(writer, columnTemplate, b.Name, b.Spec.Output.Image, b.Status.Message)
	}

	writer.Flush()

	return nil
}
