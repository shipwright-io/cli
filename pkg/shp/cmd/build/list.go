package build // nolint:revive

import (
	"fmt"
	"text/tabwriter"

	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/spf13/cobra"

	k8serrors "k8s.io/apimachinery/pkg/api/errors" // Import the k8serrors package
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// ListCommand struct contains user input to the List subcommand of Build
type ListCommand struct {
	cmd *cobra.Command

	noHeader bool
}

func listCmd() runner.SubCommand {
	listCommand := &ListCommand{
		cmd: &cobra.Command{
			Use:   "list [flags]",
			Short: "List Builds",
		},
	}

	listCommand.cmd.Flags().BoolVar(&listCommand.noHeader, "no-header", false, "Do not show columns header in list output")

	return listCommand
}

// Cmd returns cobra command object of List subcommand
func (c *ListCommand) Cmd() *cobra.Command {
	return c.cmd
}

// Complete fills object with user input data
func (c *ListCommand) Complete(_ *params.Params, _ *genericclioptions.IOStreams, _ []string) error {
	return nil
}

// Validate checks user input data
func (c *ListCommand) Validate() error {
	return nil
}

// Run contains main logic of List subcommand of Build
func (c *ListCommand) Run(params *params.Params, io *genericclioptions.IOStreams) error {
	// TODO: Support multiple output formats here, not only tabwriter
	//       find out more in kubectl libraries and use them

	// Initialize tabwriter for command output
	writer := tabwriter.NewWriter(io.Out, 0, 8, 2, '\t', 0)
	columnNames := "NAME\tOUTPUT\tSTATUS"
	columnTemplate := "%s\t%s\t%s\n"

	var buildList *buildv1beta1.BuildList
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

	if buildList, err = clientset.ShipwrightV1beta1().Builds(params.Namespace()).List(c.cmd.Context(), metav1.ListOptions{}); err != nil {
		return err
	}
	if len(buildList.Items) == 0 {
		fmt.Fprintf(io.Out, "No builds found in namespace '%s'. Please create a build or verify the namespace.\n", params.Namespace())
		return nil
	}

	if !c.noHeader {
		fmt.Fprintln(writer, columnNames)
	}

	for _, b := range buildList.Items {
		message := ""
		if b.Status.Message != nil {
			message = *b.Status.Message
		}
		fmt.Fprintf(writer, columnTemplate, b.Name, b.Spec.Output.Image, message)
	}

	return writer.Flush()
}
