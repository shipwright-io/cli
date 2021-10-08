package buildrun

import (
	"context"
	"errors"
	"fmt"
	"strings"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/shipwright-io/cli/pkg/shp/cmd/types"
	"github.com/shipwright-io/cli/pkg/shp/util"
)

var (
	// Long description for the "buildrun logs" command
	buildRunLogsLongDescription = templates.LongDesc(`
		Displays the logs for a BuildRun
	`)

	// Examples for using the "buildrun logs" command
	buildRunLogsExamples = templates.Examples(`
		$ shp buildrun logs my-buildrun
	`)
)

// BuildRunLogsOptions stores data passed to the command via command line flags
type BuildRunLogsOptions struct {
	types.SharedOptions

	BuildRunName string

	NoHeader bool
}

// newBuildRunLogsCmd creates the "buildrun logs" command
func newBuildRunLogsCmd(ctx context.Context, ioStreams *genericclioptions.IOStreams, clients *types.ClientSets, o *BuildRunLogsOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "logs <name>",
		Short:   "See BuildRun log output",
		Long:    buildRunLogsLongDescription,
		Example: buildRunLogsExamples,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(args))
			cmdutil.CheckErr(o.Run())
		},
	}

	return cmd
}

// NewBuildRunLogsCmd is a wrapper for newBuildRunLogsCmd
func NewBuildRunLogsCmd(ctx context.Context, ioStreams *genericclioptions.IOStreams, clients *types.ClientSets) *cobra.Command {
	o := &BuildRunLogsOptions{
		SharedOptions: types.SharedOptions{
			Clients: clients,
			Context: ctx,
			Streams: ioStreams,
		},
	}

	return newBuildRunLogsCmd(ctx, ioStreams, clients, o)
}

// Complete processes any data that is needed before Run executes
func (o *BuildRunLogsOptions) Complete(args []string) error {
	// Guard against index out of bound errors
	if len(args) == 0 {
		return errors.New("argument list is empty")
	}

	o.BuildRunName = args[0]

	return nil
}

// Run executes the command logic
func (o *BuildRunLogsOptions) Run() error {
	pods, err := o.Clients.KubernetesClientSet.CoreV1().Pods(o.Clients.Namespace).List(o.Context, v1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", buildv1alpha1.LabelBuildRun, o.BuildRunName),
	})
	if err != nil {
		return err
	}

	if len(pods.Items) == 0 {
		return fmt.Errorf("no builder pod found for BuildRun %q", o.BuildRunName)
	}

	fmt.Fprintf(o.Streams.Out, "Obtaining logs for BuildRun %q\n\n", o.BuildRunName)

	var b strings.Builder
	pod := pods.Items[0]
	for _, container := range pod.Spec.Containers {
		logs, err := util.GetPodLogs(o.Context, o.Clients.KubernetesClientSet, pod, container.Name)
		if err != nil {
			return err
		}

		fmt.Fprintf(&b, "*** Pod %q, container %q: ***\n\n", pod.Name, container.Name)
		fmt.Fprintln(&b, logs)
	}

	fmt.Fprintln(o.Streams.Out, b.String())

	return nil
}
