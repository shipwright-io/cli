package buildrun

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/shipwright-io/cli/pkg/shp/cmd/types"
)

var (
	// Long description for the "buildrun list" command
	buildRunListLongDescription = templates.LongDesc(`
		Lists existing BuildRuns
	`)

	// Examples for using the "buildrun list" command
	buildRunListExamples = templates.Examples(`
		$ shp buildrun list
	`)
)

// BuildRunListOptions stores data passed to the command via command line flags
type BuildRunListOptions struct {
	types.SharedOptions

	Name string

	NoHeader bool
}

// newBuildRunListCmd creates the "buildrun list" command
func newBuildRunListCmd(ctx context.Context, ioStreams *genericclioptions.IOStreams, clients *types.ClientSets, o *BuildRunListOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list [flags]",
		Short:   "List existing Builds",
		Long:    buildRunListLongDescription,
		Example: buildRunListExamples,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(args))
			cmdutil.CheckErr(o.Run())
		},
	}

	cmd.Flags().BoolVar(&o.NoHeader, "no-header", false, "Do not show columns header in list output")

	return cmd
}

// NewBuildRunListCmd is a wrapper for newBuildRunListCmd
func NewBuildRunListCmd(ctx context.Context, ioStreams *genericclioptions.IOStreams, clients *types.ClientSets) *cobra.Command {
	o := &BuildRunListOptions{
		SharedOptions: types.SharedOptions{
			Clients: clients,
			Context: ctx,
			Streams: ioStreams,
		},
	}

	return newBuildRunListCmd(ctx, ioStreams, clients, o)
}

// Complete processes any data that is needed before Run executes
func (o *BuildRunListOptions) Complete(args []string) error {
	return nil
}

// Run executes the command logic
func (o *BuildRunListOptions) Run() error {
	// TODO: Support multiple output formats here, not only tabwriter
	//       find out more in kubectl libraries and use them

	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 2, '\t', 0)
	columnNames := "NAME\tSTATUS"
	columnTemplate := "%s\t%s\n"

	brs, err := o.Clients.ShipwrightClientSet.ShipwrightV1alpha1().BuildRuns(o.Clients.Namespace).List(o.Context, metav1.ListOptions{})
	if err != nil {
		return err
	}

	if !o.NoHeader {
		fmt.Fprintln(writer, columnNames)
	}

	for _, br := range brs.Items {
		name := br.Name
		status := string(metav1.ConditionUnknown)
		for _, condition := range br.Status.Conditions {
			if condition.Type == buildv1alpha1.Succeeded {
				status = condition.Reason
				break
			}
		}

		fmt.Fprintf(writer, columnTemplate, name, status)
	}

	writer.Flush()

	return nil
}
