package buildrun

import (
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"

	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/follower"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/reactor"
)

// LogsCommand contains data input from user for logs sub-command
type LogsCommand struct {
	cmd *cobra.Command

	follow       bool   // follow flag, follows the pod logs
	buildRunName string // buildrun name, added during complete

	podLogsFollower *follower.PodLogsFollower
}

func logsCmd() runner.SubCommand {
	logCommand := &LogsCommand{cmd: &cobra.Command{
		Use:   "logs <name>",
		Short: "See BuildRun log output",
		Args:  cobra.ExactArgs(1),
	}}
	logCommand.cmd.Flags().BoolVarP(
		&logCommand.follow,
		"follow",
		"F",
		logCommand.follow,
		"Follow the log of a buildrun until it completes or fails.",
	)
	return logCommand
}

// Cmd returns cobra command object
func (c *LogsCommand) Cmd() *cobra.Command {
	return c.cmd
}

// Complete fills in data provided by user
func (c *LogsCommand) Complete(p params.Interface, ioStreams *genericclioptions.IOStreams, args []string) error {
	if len(args) != 1 {
		return errors.New("buildrun name is not informed")
	}
	c.buildRunName = args[0]
	return nil
}

// Validate validates data input by user
func (c *LogsCommand) Validate() error {
	return nil
}

// Run stream the pod logs, either following the pod subsequent modifications, or only once.
func (c *LogsCommand) Run(p params.Interface, ioStreams *genericclioptions.IOStreams) error {
	clientset, err := p.ShipwrightClientSet()
	if err != nil {
		return nil
	}

	ctx := c.cmd.Context()

	// checking if the target BuildRun instance exists, otherwise works as a short circuit to print
	// out error and avoid trying to retrieve/follow logs
	name := types.NamespacedName{Namespace: p.Namespace(), Name: c.buildRunName}
	_, err = clientset.ShipwrightV1alpha1().
		BuildRuns(name.Namespace).
		Get(ctx, name.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if c.podLogsFollower == nil {
		pw, err := reactor.NewPodWatcherFromParams(ctx, p)
		if err != nil {
			return err
		}
		c.podLogsFollower, err = follower.NewPodLogsFollowerFromParams(ctx, p, pw, ioStreams)
		if err != nil {
			return err
		}
	}

	if !c.follow {
		c.podLogsFollower.SetOnlyOnce()
	}

	listOpts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%v=%v", buildv1alpha1.LabelBuildRun, c.buildRunName),
	}
	if _, err = c.podLogsFollower.Start(listOpts); err != nil {
		_ = InspectBuildRun(ctx, clientset, name, ioStreams)
	}
	return err
}
