package buildrun

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	"github.com/spf13/cobra"

	"github.com/shipwright-io/cli/pkg/shp/cmd/follower"
	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/util"
)

// LogsCommand contains data input from user for logs sub-command
type LogsCommand struct {
	cmd *cobra.Command

	name string

	follow   bool
	follower *follower.Follower
}

func logsCmd() runner.SubCommand {
	cmd := &cobra.Command{
		Use:   "logs <name>",
		Short: "See BuildRun log output",
		Args:  cobra.ExactArgs(1),
	}
	logCommand := &LogsCommand{
		cmd: cmd,
	}
	cmd.Flags().BoolVarP(&logCommand.follow, "follow", "F", logCommand.follow, "Follow the log of a buildrun until it completes or fails.")
	return logCommand
}

// Cmd returns cobra command object
func (c *LogsCommand) Cmd() *cobra.Command {
	return c.cmd
}

// Complete fills in data provided by user
func (c *LogsCommand) Complete(params *params.Params, ioStreams *genericclioptions.IOStreams, args []string) error {
	c.name = args[0]
	if !c.follow {
		return nil
	}

	br := types.NamespacedName{
		Namespace: params.Namespace(),
		Name:      c.name,
	}
	var err error
	c.follower, err = params.NewFollower(c.Cmd().Context(), br, ioStreams)
	return err
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
		LabelSelector: fmt.Sprintf("%v=%v", buildv1beta1.LabelBuildRun, c.name),
	}

	// first see if pod is already done; if so, even if we have follow == true, just do the normal path;
	// we don't employ a pod watch here since the buildrun may already be complete before 'shp buildrun logs -F'
	// is invoked.
	justGetLogs := false
	var pods *corev1.PodList
	err = wait.PollUntilContextTimeout(c.cmd.Context(), 1*time.Second, 10*time.Second, true, func(ctx context.Context) (done bool, err error) {
		if pods, err = clientset.CoreV1().Pods(params.Namespace()).List(ctx, lo); err != nil {
			fmt.Fprintf(ioStreams.ErrOut, "error listing Pods for BuildRun %q: %s\n", c.name, err.Error())
			return false, nil
		}
		if len(pods.Items) == 0 {
			fmt.Fprintf(ioStreams.ErrOut, "no builder pod found for BuildRun %q\n", c.name)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return err
	}
	pod := pods.Items[0]
	phase := pod.Status.Phase
	if phase == corev1.PodFailed || phase == corev1.PodSucceeded {
		justGetLogs = true
	}

	if !c.follow || justGetLogs {
		fmt.Fprintf(ioStreams.Out, "Obtaining logs for BuildRun %q\n\n", c.name)

		var b strings.Builder
		for _, container := range append(pod.Spec.InitContainers, pod.Spec.Containers...) {
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
	_, err = c.follower.Start(lo)
	return err
}
