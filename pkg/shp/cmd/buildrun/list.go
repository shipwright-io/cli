package buildrun

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"

	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/params"
)

// ListCommand contains data input from user for list sub-command
type ListCommand struct {
	cmd *cobra.Command

	noHeader bool
	wide     bool
}

func listCmd() runner.SubCommand {
	listCmd := &ListCommand{
		cmd: &cobra.Command{
			Use:   "list [flags]",
			Short: "List Builds",
		},
	}

	listCmd.cmd.Flags().BoolVar(&listCmd.noHeader, "no-header", false, "Do not show columns header in list output")
	listCmd.cmd.Flags().BoolVar(&listCmd.wide, "wide", false, "Display additional fields such as source, output-image, build-name, elapsed-time and source-origin in list output")

	return listCmd
}

// Cmd returns cobra command object
func (c *ListCommand) Cmd() *cobra.Command {
	return c.cmd
}

// Complete fills in data provided by user
func (c *ListCommand) Complete(_ *params.Params, _ *genericclioptions.IOStreams, _ []string) error {
	return nil
}

// Validate validates data input by user
func (c *ListCommand) Validate() error {
	return nil
}

// Run executes list sub-command logic
func (c *ListCommand) Run(params *params.Params, io *genericclioptions.IOStreams) error {
	// TODO: Support multiple output formats here, not only tabwriter
	//       find out more in kubectl libraries and use them

	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 2, '\t', 0)
	columnNames := "NAME\tSTATUS\tAGE"
	columnTemplate := "%s\t%s\t%s\n"

	clientset, err := params.ShipwrightClientSet()
	if err != nil {
		return err
	}

	k8sclient, err := params.ClientSet()
	if err != nil {
		return fmt.Errorf("failed to get k8s client: %w", err)
	}
	_, err = k8sclient.CoreV1().Namespaces().Get(c.cmd.Context(), params.Namespace(), metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			fmt.Fprintf(io.Out, "Namespace '%s' not found. Please ensure that the namespace exists and try again.\n", params.Namespace())
			return nil
		}
		return err
	}

	var brs *buildv1beta1.BuildRunList
	if brs, err = clientset.ShipwrightV1beta1().BuildRuns(params.Namespace()).List(c.cmd.Context(), metav1.ListOptions{}); err != nil {
		return err
	}
	if len(brs.Items) == 0 {
		fmt.Fprintf(io.Out, "No buildruns found in namespace '%s'. Please create a buildrun or verify the namespace.\n", params.Namespace())
		return nil
	}

	if !c.noHeader {
		if c.wide {
			columnNames = "NAME\tSTATUS\tAGE\tSOURCE\tOUTPUT-IMAGE\tBUILD-NAME\tELAPSED-TIME\tSOURCE-ORIGIN"
		}
		fmt.Fprintln(writer, columnNames)
	}

	if c.wide {
		columnTemplate = "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n"
	}

	for _, br := range brs.Items {
		name := br.Name
		buildName := br.Spec.BuildName()
		outputImage := br.Status.BuildSpec.Output.Image
		sourceOrigin := br.Status.BuildSpec.Source.Type
		source := "-"
		if sourceOrigin == "Git" {
			source = br.Status.BuildSpec.Source.Git.URL
			revision := br.Status.BuildSpec.Source.Git.Revision
			if revision != nil {
				source += "@" + *revision
			}
		}

		status := string(metav1.ConditionUnknown)
		for _, condition := range br.Status.Conditions {
			if condition.Type == buildv1beta1.Succeeded {
				status = condition.Reason
				break
			}
		}
		age := duration.ShortHumanDuration(time.Since((br.ObjectMeta.CreationTimestamp).Time))
		elapsedTime := age
		if br.Status.StartTime != nil && br.Status.CompletionTime != nil {
			duration := br.Status.CompletionTime.Time.Sub(br.Status.StartTime.Time)
			elapsedTime = duration.String()
		}
		if c.wide {
			fmt.Fprintf(writer, columnTemplate, name, status, age, source, outputImage, buildName, elapsedTime, sourceOrigin)
		} else {
			fmt.Fprintf(writer, columnTemplate, name, status, age)
		}
	}

	return writer.Flush()
}
