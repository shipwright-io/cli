package buildrun

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/params"
)

// GetCommand contains user input for the `get` subcommand of BuildRun
type GetCommand struct {
	cmd    *cobra.Command
	name   string
	output string
}

func (c *GetCommand) Cmd() *cobra.Command {
	return c.cmd
}

func (c *GetCommand) Complete(_ *params.Params, _ *genericclioptions.IOStreams, args []string) error {
	c.name = args[0]
	return nil
}

func (c *GetCommand) Validate() error {
	if c.name == "" {
		return fmt.Errorf("name must be provided")
	}
	if c.output != "" && c.output != "json" && c.output != "yaml" {
		return fmt.Errorf("unsupported output format %q. Supported formats are: json, yaml", c.output)
	}
	return nil
}

func (c *GetCommand) Run(params *params.Params, ioStreams *genericclioptions.IOStreams) error {
	clientset, err := params.ShipwrightClientSet()
	if err != nil {
		return err
	}

	ns := params.Namespace()
	buildRun, err := clientset.ShipwrightV1beta1().BuildRuns(ns).Get(c.cmd.Context(), c.name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get buildrun %q in namespace %q: %w", c.name, ns, err)
	}

	switch c.output {
	case "json":
		data, err := json.MarshalIndent(buildRun, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(ioStreams.Out, string(data))
		return nil

	case "yaml":
		data, err := yaml.Marshal(buildRun)
		if err != nil {
			return err
		}
		fmt.Fprintln(ioStreams.Out, string(data))
		return nil

	case "":
		w := tabwriter.NewWriter(ioStreams.Out, 0, 8, 2, '\t', 0)
		fmt.Fprintf(w, "NAME:\t%s\n", buildRun.Name)
		fmt.Fprintf(w, "NAMESPACE:\t%s\n", buildRun.Namespace)
		if buildName := buildRun.Spec.BuildName(); buildName != "" {
			fmt.Fprintf(w, "BUILD NAME:\t%s\n", buildName)
		}

		status := "Unknown"
		if buildRun.Status.Conditions != nil {
			for _, condition := range buildRun.Status.Conditions {
				if condition.Type == buildv1beta1.Succeeded {
					if condition.Reason != "" {
						status = condition.Reason
					} else {
						status = string(condition.Status)
					}
					break
				}
			}
		}
		fmt.Fprintf(w, "STATUS:\t%s\n", status)

		if buildRun.Status.StartTime != nil {
			fmt.Fprintf(w, "START TIME:\t%s\n", buildRun.Status.StartTime.Time.Format(time.RFC3339))
		}
		if buildRun.Status.CompletionTime != nil {
			fmt.Fprintf(w, "COMPLETION TIME:\t%s\n", buildRun.Status.CompletionTime.Time.Format(time.RFC3339))
		}

		return w.Flush()

	default:
		return fmt.Errorf("unsupported output format %q. Supported formats are: json, yaml", c.output)
	}
}

func getCmd() runner.SubCommand {
	cmd := &cobra.Command{
		Use:   "get <name> [flags]",
		Short: "Get details of a BuildRun",
		Args:  cobra.ExactArgs(1),
	}

	c := &GetCommand{
		cmd: cmd,
	}

	cmd.Flags().StringVarP(&c.output, "output", "o", "", "Output format. Allowed values: json, yaml")
	return c
}
