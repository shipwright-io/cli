package buildrun

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/util"
)

// LogsCommand contains data input from user for logs sub-command
type LogsCommand struct {
	cmd *cobra.Command

	name string
}

func logsCmd() runner.SubCommand {
	return &LogsCommand{
		cmd: &cobra.Command{
			Use:   "logs <name>",
			Short: "See BuildRun log output",
			Args:  cobra.ExactArgs(1),
		},
	}
}

// Cmd returns cobra command object
func (c *LogsCommand) Cmd() *cobra.Command {
	return c.cmd
}

// Complete fills in data provided by user
func (c *LogsCommand) Complete(params *params.Params, args []string) error {
	c.name = args[0]

	return nil
}

// Validate validates data input by user
func (c *LogsCommand) Validate() error {
	return nil
}

// Run executes logs sub-command logic
func (c *LogsCommand) Run(params *params.Params, ioStreams *genericclioptions.IOStreams) error {
	clientset, err := params.ClientSet()
	if err != nil {
		return err
	}

	lo := v1.ListOptions{
		LabelSelector: fmt.Sprintf("%v=%v", buildv1alpha1.LabelBuildRun, c.name),
	}

	var pods *corev1.PodList
	if pods, err = clientset.CoreV1().Pods(params.Namespace()).List(c.cmd.Context(), lo); err != nil {
		return err
	}
	if len(pods.Items) == 0 {
		return fmt.Errorf("no builder pod found for BuildRun %q", c.name)
	}

	fmt.Fprintf(ioStreams.Out, "Obtaining logs for BuildRun %q\n\n", c.name)

	var b strings.Builder
	pod := pods.Items[0]
	for _, container := range pod.Spec.Containers {
		logs, err := util.GetPodLogs(c.cmd.Context(), clientset, pod, container.Name)
		if err != nil {
			return err
		}

		fmt.Fprintf(&b, "*** Pod %q, container %q: ***\n\n", pod.Name, container.Name)
		fmt.Fprintln(&b, logs)
	}

	fmt.Fprintln(ioStreams.Out, b.String())

	return nil
}
