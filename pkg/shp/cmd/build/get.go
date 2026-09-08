package build // nolint:revive

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/params"
)

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
	return nil
}

func (c *GetCommand) Run(params *params.Params, ioStreams *genericclioptions.IOStreams) error {
	clientset, err := params.ShipwrightClientSet()
	if err != nil {
		return err
	}

	ns := params.Namespace()
	build, err := clientset.ShipwrightV1beta1().Builds(ns).Get(c.cmd.Context(), c.name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			fmt.Fprintf(ioStreams.Out, "Build '%s' not found in namespace '%s'.\n", c.name, ns)
			return nil
		}
		return err
	}

	switch c.output {
	case "json":
		data, err := json.MarshalIndent(build, "", " ");
		if err != nil {
			return err
		}
		fmt.Fprintln(ioStreams.Out, string(data))
		return nil
	case "yaml":
		data, err := yaml.Marshal(build);
		if err != nil {
			return err;
		}
		fmt.Fprintln(ioStreams.Out, string(data))
		return nil
	case "":
		w := tabwriter.NewWriter(ioStreams.Out, 0, 8, 2, '\t', 0);
		fmt.Fprintf(w, "NAME:\t%s\n", build.Name)
		fmt.Fprintf(w, "NAMESPACE:\t%s\n", build.Namespace)
		if build.Spec.Source.Git != nil {
			fmt.Fprintf(w, "SOURCE URL:\t%s\n", build.Spec.Source.Git.URL)
			if build.Spec.Source.Git.Revision != nil {
				fmt.Fprintf(w, "REVISION:\t%s\n", *build.Spec.Source.Git.Revision)
			}
		}
		if build.Spec.Strategy.Name != "" {
			kind := ""
			if build.Spec.Strategy.Kind != nil {
				kind = string(*build.Spec.Strategy.Kind)
			}
			if kind != "" {
				fmt.Fprintf(w, "STRATEGY:\t%s (%s)\n", build.Spec.Strategy.Name, kind)
			} else {
				fmt.Fprintf(w, "STRATEGY:\t%s\n", build.Spec.Strategy.Name)
			}
		}
		if build.Spec.Output.Image != "" {
			fmt.Fprintf(w, "OUTPUT IMAGE:\t%s\n", build.Spec.Output.Image)
		}
		if build.Status.Registered != nil {
			fmt.Fprintf(w, "REGISTERED:\t%s\n", *build.Status.Registered)
		}

		return w.Flush()
	default:
		return fmt.Errorf("unsupported output format %q. Supported formats are: json, yaml", c.output)
	}
}

func getCmd() runner.SubCommand {
	cmd := &cobra.Command{
		Use:   "get <name> [flags]",
		Short: "Get details of a build",
		Args:  cobra.ExactArgs(1),
	}

	c := &GetCommand{
		cmd: cmd,
	}

	cmd.Flags().StringVarP(&c.output, "output", "o", "", "Output format. Allowed values: json, yaml")
	return c
}