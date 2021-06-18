package build

import (
	"errors"
	"fmt"
	"time"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/flags"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/reactor"
	"github.com/shipwright-io/cli/pkg/shp/resource"
	"github.com/shipwright-io/cli/pkg/shp/tail"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

// RunCommand represents the `build run` sub-command, which creates a unique BuildRun instance to run
// the build process, informed via arguments.
type RunCommand struct {
	cmd *cobra.Command // cobra command instance

	clientset       kubernetes.Interface         // kubernetes client
	ioStreams       *genericclioptions.IOStreams // io-streams instance
	ns              string                       // namespace name
	pw              *reactor.PodWatcher          // pod-watcher instance
	logTail         *tail.Tail                   // follow container logs
	tailLogsStarted map[string]bool

	buildName    string                      // build name
	buildRunSpec *buildv1alpha1.BuildRunSpec // stores command-line flags
	Follow       bool                        // flag to tail pod logs
}

const buildRunLongDesc = `
Creates a unique BuildRun instance for the given Build, which starts the build
process orchestrated by the Shipwright build controller. For example:

	$ shp build run my-app
`

// Cmd returns cobra.Command object of the create sub-command.
func (r *RunCommand) Cmd() *cobra.Command {
	return r.cmd
}

// Complete picks the build resource name from arguments, and instantiate additional components.
func (r *RunCommand) Complete(params *params.Params, args []string) error {
	switch len(args) {
	case 1:
		r.buildName = args[0]
	default:
		return errors.New("Build name is not informed")
	}

	var err error
	if r.clientset, err = params.ClientSet(); err != nil {
		return err
	}
	ctx := r.Cmd().Context()

	r.logTail = tail.NewTail(ctx, r.clientset)
	r.ns = params.Namespace()

	// instantiating a pod watcher with a specific label-selector to find the indented pod where the
	// actual build started by this subcommand is being executed
	labelSelector := fmt.Sprintf("build.shipwright.io/name=%s", r.buildName)
	listOpts := metav1.ListOptions{LabelSelector: labelSelector}
	r.pw, err = reactor.NewPodWatcher(ctx, r.clientset, listOpts, r.ns)
	if err != nil {
		return err
	}

	// overwriting build-ref name to use what's on arguments
	return r.Cmd().Flags().Set(flags.BuildrefNameFlag, r.buildName)
}

// Validate the user must inform the build resource name.
func (r *RunCommand) Validate() error {
	if r.buildName == "" {
		return fmt.Errorf("name is not informed")
	}
	return nil
}

// tailLogs start tailing logs for each container name, if not started already.
func (r *RunCommand) tailLogs(pod *corev1.Pod) {
	for _, container := range pod.Spec.Containers {
		if _, exists := r.tailLogsStarted[container.Name]; exists {
			continue
		}
		r.tailLogsStarted[container.Name] = true
		r.logTail.Start(pod.GetNamespace(), pod.GetName(), container.Name)
	}
}

// onEvent reacts on pod state changes, to start and stop tailing container logs.
func (r *RunCommand) onEvent(pod *corev1.Pod) error {
	switch pod.Status.Phase {
	case corev1.PodPending:
		fmt.Fprintf(r.ioStreams.Out, "Pod '%s' is pending...\n", pod.GetName())
	case corev1.PodRunning:
		// graceful time to wait for container start
		time.Sleep(3 * time.Second)
		// start tailing container logs
		r.tailLogs(pod)
	case corev1.PodFailed:
		fmt.Fprintf(r.ioStreams.Out, "Pod '%s' has failed!\n", pod.GetName())
		r.stop()
		return fmt.Errorf("build pod '%s' has failed", pod.GetName())
	case corev1.PodSucceeded:
		fmt.Fprintf(r.ioStreams.Out, "Pod '%s' has succeeded!\n", pod.GetName())
		r.stop()
		// TODO: print out details of the container image that has just been built;
		return nil
	}
	return nil
}

// stop invoke stop on streaming components.
func (r *RunCommand) stop() {
	r.logTail.Stop()
	r.pw.Stop()
}

// Run creates a BuildRun resource based on Build's name informed on arguments.
func (r *RunCommand) Run(params *params.Params, ioStreams *genericclioptions.IOStreams) error {
	// resource using GenerateName, which will provice a unique instance
	br := &buildv1alpha1.BuildRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", r.buildName),
		},
		Spec: *r.buildRunSpec,
	}
	flags.SanitizeBuildRunSpec(&br.Spec)

	buildRunResource := resource.GetBuildRunResource(params)
	if err := buildRunResource.Create(r.cmd.Context(), "", br); err != nil {
		return err
	}

	if !r.Follow {
		fmt.Fprintf(ioStreams.Out, "BuildRun created %q for build %q\n", br.GetName(), r.buildName)
		return nil
	}

	r.ioStreams = ioStreams
	r.pw.WithOnPodModifiedFn(r.onEvent)
	return r.pw.Start()
}

// runCmd instantiate the "build run" sub-command using common BuildRun flags.
func runCmd() runner.SubCommand {
	cmd := &cobra.Command{
		Use:   "run <name>",
		Short: "Start a build specified by 'name'",
		Long:  buildRunLongDesc,
	}
	runCommand := &RunCommand{
		cmd:             cmd,
		buildRunSpec:    flags.BuildRunSpecFromFlags(cmd.Flags()),
		tailLogsStarted: make(map[string]bool),
	}
	cmd.Flags().BoolVarP(&runCommand.Follow, "follow", "F", runCommand.Follow, "Start a build and watch its log until it completes or fails.")
	return runCommand
}
