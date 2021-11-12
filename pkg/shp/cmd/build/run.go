package build

import (
	"errors"
	"fmt"
	"sync"
	"time"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	buildclientset "github.com/shipwright-io/build/pkg/client/clientset/versioned"

	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/flags"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/reactor"
	"github.com/shipwright-io/cli/pkg/shp/tail"

	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// RunCommand represents the `build run` sub-command, which creates a unique BuildRun instance to run
// the build process, informed via arguments.
type RunCommand struct {
	cmd *cobra.Command // cobra command instance

	ioStreams       *genericclioptions.IOStreams // io-streams instance
	pw              *reactor.PodWatcher          // pod-watcher instance
	logTail         *tail.Tail                   // follow container logs
	tailLogsStarted map[string]bool              // controls tail instance per container

	logLock sync.Mutex

	buildName    string
	buildRunName string
	namespace    string
	buildRunSpec *buildv1alpha1.BuildRunSpec // stores command-line flags
	shpClientset buildclientset.Interface
	follow       bool // flag to tail pod logs
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
func (r *RunCommand) Complete(params *params.Params, io *genericclioptions.IOStreams, args []string) error {
	switch len(args) {
	case 1:
		r.buildName = args[0]
	default:
		return errors.New("build name is not informed")
	}

	clientset, err := params.ClientSet()
	if err != nil {
		return err
	}
	r.logTail = tail.NewTail(r.Cmd().Context(), clientset)
	r.ioStreams = io
	r.namespace = params.Namespace()
	if r.follow {
		if r.shpClientset, err = params.ShipwrightClientSet(); err != nil {
			return err
		}

		kclientset, err := params.ClientSet()
		if err != nil {
			return err
		}
		to, err := params.RequestTimeout()
		if err != nil {
			return err
		}
		r.pw, err = reactor.NewPodWatcher(r.Cmd().Context(), to, kclientset, params.Namespace())
		if err != nil {
			return err
		}

		r.pw.WithOnPodModifiedFn(r.onEvent)
		r.pw.WithTimeoutPodFn(r.onTimeout)
		r.pw.WithNoPodEventsYetFn(r.onNoPodEventsYet)

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

// tailLogs start tailing logs for each container name in init-containers and containers, if not
// started already.
func (r *RunCommand) tailLogs(pod *corev1.Pod) {
	containers := append(pod.Spec.InitContainers, pod.Spec.Containers...)
	for _, container := range containers {
		if _, exists := r.tailLogsStarted[container.Name]; exists {
			continue
		}
		r.tailLogsStarted[container.Name] = true
		r.logTail.Start(pod.GetNamespace(), pod.GetName(), container.Name)
	}
}

// onNoPodEventsYet reacts to the pod watcher telling us it has not received any pod events for our build run
func (r *RunCommand) onNoPodEventsYet() {
	r.Log(fmt.Sprintf("BuildRun %q log following has not observed any pod events yet.", r.buildRunName))
	br, err := r.shpClientset.ShipwrightV1alpha1().BuildRuns(r.namespace).Get(r.cmd.Context(), r.buildRunName, metav1.GetOptions{})
	if err != nil {
		r.Log(fmt.Sprintf("error accessing BuildRun %q: %s", r.buildRunName, err.Error()))
		return
	}

	c := br.Status.GetCondition(buildv1alpha1.Succeeded)
	giveUp := false
	msg := ""
	switch {
	case c != nil && c.Status == corev1.ConditionTrue:
		giveUp = true
		msg = fmt.Sprintf("BuildRun '%s' has been marked as successful.\n", br.Name)
	case c != nil && c.Status == corev1.ConditionFalse:
		giveUp = true
		msg = fmt.Sprintf("BuildRun '%s' has been marked as failed.\n", br.Name)
	case br.IsCanceled():
		giveUp = true
		msg = fmt.Sprintf("BuildRun '%s' has been canceled.\n", br.Name)
	case br.DeletionTimestamp != nil:
		giveUp = true
		msg = fmt.Sprintf("BuildRun '%s' has been deleted.\n", br.Name)
	case !br.HasStarted():
		r.Log(fmt.Sprintf("BuildRun '%s' has been marked as failed.\n", br.Name))
	}
	if giveUp {
		r.Log(msg)
		r.Log(fmt.Sprintf("exiting 'ship build run --follow' for BuildRun %q", br.Name))
		r.stop()
	}

}

// onTimeout reacts to either the context or request timeout causing the pod watcher to exit
func (r *RunCommand) onTimeout(msg string) {
	r.Log(fmt.Sprintf("BuildRun %q log following has stopped because: %q\n", r.buildRunName, msg))
}

// onEvent reacts on pod state changes, to start and stop tailing container logs.
func (r *RunCommand) onEvent(pod *corev1.Pod) error {
	switch pod.Status.Phase {
	case corev1.PodRunning:
		// graceful time to wait for container start
		time.Sleep(3 * time.Second)
		// start tailing container logs
		r.tailLogs(pod)
	case corev1.PodFailed:
		msg := ""
		br, err := r.shpClientset.ShipwrightV1alpha1().BuildRuns(pod.Namespace).Get(r.cmd.Context(), r.buildRunName, metav1.GetOptions{})
		switch {
		case err == nil && br.IsCanceled():
			msg = fmt.Sprintf("BuildRun '%s' has been canceled.\n", br.Name)
		case err == nil && br.DeletionTimestamp != nil:
			msg = fmt.Sprintf("BuildRun '%s' has been deleted.\n", br.Name)
		case pod.DeletionTimestamp != nil:
			msg = fmt.Sprintf("Pod '%s' has been deleted.\n", pod.GetName())
		default:
			msg = fmt.Sprintf("Pod '%s' has failed!\n", pod.GetName())
			err = fmt.Errorf("build pod '%s' has failed", pod.GetName())
		}
		// see if because of deletion or cancelation
		r.Log(msg)
		r.stop()
		return err
	case corev1.PodSucceeded:
		r.Log(fmt.Sprintf("Pod '%s' has succeeded!\n", pod.GetName()))
		r.stop()
	default:
		r.Log(fmt.Sprintf("Pod '%s' is in state %q...\n", pod.GetName(), string(pod.Status.Phase)))
		// handle any issues with pulling images that may fail
		for _, c := range pod.Status.Conditions {
			if c.Type == corev1.PodInitialized || c.Type == corev1.ContainersReady {
				if c.Status == corev1.ConditionUnknown {
					return fmt.Errorf(c.Message)
				}
			}
		}
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
	// resource using GenerateName, which will provide a unique instance
	br := &buildv1alpha1.BuildRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", r.buildName),
		},
		Spec: *r.buildRunSpec,
	}
	flags.SanitizeBuildRunSpec(&br.Spec)

	clientset, err := params.ShipwrightClientSet()
	if err != nil {
		return err
	}
	br, err = clientset.ShipwrightV1alpha1().BuildRuns(r.namespace).Create(r.cmd.Context(), br, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	if !r.follow {
		fmt.Fprintf(ioStreams.Out, "BuildRun created %q for build %q\n", br.GetName(), r.buildName)
		return nil
	}

	r.buildRunName = br.Name

	// instantiating a pod watcher with a specific label-selector to find the indented pod where the
	// actual build started by this subcommand is being executed, including the randomized buildrun
	// name
	listOpts := metav1.ListOptions{LabelSelector: fmt.Sprintf(
		"build.shipwright.io/name=%s,buildrun.shipwright.io/name=%s",
		r.buildName,
		br.GetName(),
	)}
	_, err = r.pw.Start(listOpts)
	return err
}

// Log prints a message
func (r *RunCommand) Log(msg string) {
	// concurrent fmt.Fprintf(r.ioStream.Out...) calls need locking to avoid data races, as we 'write' to the stream
	r.logLock.Lock()
	defer r.logLock.Unlock()
	fmt.Fprint(r.ioStreams.Out, msg)
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
		logLock:         sync.Mutex{},
	}
	cmd.Flags().BoolVarP(&runCommand.follow, "follow", "F", runCommand.follow, "Start a build and watch its log until it completes or fails.")
	return runCommand
}
