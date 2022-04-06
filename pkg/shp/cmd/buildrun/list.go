package buildrun

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"

	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/params"
)

// ListCommand contains data input from user for list sub-command
type ListCommand struct {
	cmd *cobra.Command

	noHeader bool
}

func listCmd() runner.SubCommand {
	listCmd := &ListCommand{
		cmd: &cobra.Command{
			Use:   "list [flags]",
			Short: "List Builds",
		},
	}

	listCmd.cmd.Flags().BoolVar(&listCmd.noHeader, "no-header", false, "Do not show columns header in list output")

	return listCmd
}

// Cmd returns cobra command object
func (c *ListCommand) Cmd() *cobra.Command {
	return c.cmd
}

// Complete fills in data provided by user
func (c *ListCommand) Complete(_ params.Interface, _ *genericclioptions.IOStreams, _ []string) error {
	return nil
}

// Validate validates data input by user
func (c *ListCommand) Validate() error {
	return nil
}

// Run executes list sub-command logic
func (c *ListCommand) Run(p params.Interface, io *genericclioptions.IOStreams) error {
	// TODO: Support multiple output formats here, not only tabwriter
	//       find out more in kubectl libraries and use them

	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 2, '\t', 0)
	columnNames := "NAME\tSTATUS\tAGE"
	columnTemplate := "%s\t%s\t%s\n"

	clientset, err := p.ShipwrightClientSet()
	if err != nil {
		return err
	}

	var brs *buildv1alpha1.BuildRunList
	if brs, err = clientset.ShipwrightV1alpha1().BuildRuns(p.Namespace()).List(c.cmd.Context(), metav1.ListOptions{}); err != nil {
		return err
	}

	if !c.noHeader {
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
		age := duration.ShortHumanDuration(time.Since((br.Status.StartTime).Time))

		fmt.Fprintf(writer, columnTemplate, name, status, age)
	}

	writer.Flush()

	return nil
}
