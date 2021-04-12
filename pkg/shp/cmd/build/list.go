package build

import (
	"fmt"
	"text/tabwriter"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/resource"
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
func (c *ListCommand) Complete(params *params.Params, args []string) error {
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

	var buildList buildv1alpha1.BuildList
	br := resource.GetBuildResource(params)

	if err := br.List(c.cmd.Context(), &buildList); err != nil {
		return err
	}

	if !c.noHeader {
		fmt.Fprintln(writer, columnNames)
	}

	for _, b := range buildList.Items {
		fmt.Fprintf(writer, columnTemplate, b.Name, b.Spec.Output.Image, b.Status.Message)
	}

	writer.Flush()

	return nil
}
