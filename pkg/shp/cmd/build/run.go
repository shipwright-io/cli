package build

import (
	"context"
	"fmt"
	"sync"
	"time"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/shipwright-io/cli/pkg/shp/cmd/types"
	"github.com/shipwright-io/cli/pkg/shp/reactor"
	"github.com/shipwright-io/cli/pkg/shp/tail"
)

var (
	// Long description for the "build run" command
	buildRunLongDescription = templates.LongDesc(`
		Run a Build
	`)

	// Examples for using the "build run" command
	buildRunExamples = templates.Examples(`
		$ shp build run my-build
	`)
)

// BuildRunOptions stores data passed to the command via command line flags
type BuildRunOptions struct {
	types.SharedOptions

	PodWatcher *reactor.PodWatcher

	BuildRun *buildv1alpha1.BuildRun

	BuildName    string
	BuildRunName string
	BuildRunSpec *buildv1alpha1.BuildRunSpec

	BuildRefName       string
	BuildRefAPIVersion string

	ServiceAccountName     string
	ServiceAccountGenerate bool

	OutputImage                 string
	OutputCredentialsSecretName string

	Timeout metav1.Duration

	FollowLogs bool

	LogTail          *tail.Tail
	LogTailStartedOn map[string]bool

	WatchLock sync.Mutex
}

// newBuildRunCmd creates the "build run" command
func newBuildRunCmd(ctx context.Context, ioStreams *genericclioptions.IOStreams, clients *types.ClientSets, o *BuildRunOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run <name>",
		Short:   "Run a Build",
		Long:    buildRunLongDescription,
		Example: buildRunExamples,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(args))
			cmdutil.CheckErr(o.Run())
		},
	}

	cmd.Flags().StringVar(&o.BuildRefName, "buildref-name", "", "name of build resource to reference")
	cmd.Flags().StringVar(&o.BuildRefAPIVersion, "buildref-apiversion", "", "API version of build resource to reference")

	cmd.Flags().StringVar(&o.ServiceAccountName, "sa-name", "", "Kubernetes service-account name")
	cmd.Flags().BoolVar(&o.ServiceAccountGenerate, "sa-generate", false, "generate a Kubernetes service-account for the build")

	cmd.Flags().StringVar(&o.OutputImage, "output-image", "", "The location to push the output image to.")
	cmd.MarkFlagRequired("output-image")
	cmd.Flags().StringVar(&o.OutputCredentialsSecretName, "output-credentials-secret", "", "The name of the Secret that contains credentials for the repository to push the built Image to.")

	cmd.Flags().DurationVar(&o.Timeout.Duration, "timeout", time.Duration(0), "How long to let the build run before timing out.")

	cmd.Flags().BoolVarP(&o.FollowLogs, "follow", "F", o.FollowLogs, "Watch logs until the BuildRun completes or fails.")

	return cmd
}

// NewBuildRunCmd is a wrapper for newBuildRunCmd
func NewBuildRunCmd(ctx context.Context, ioStreams *genericclioptions.IOStreams, clients *types.ClientSets) *cobra.Command {
	o := &BuildRunOptions{
		SharedOptions: types.SharedOptions{
			Clients: clients,
			Context: ctx,
			Streams: ioStreams,
		},
	}

	return newBuildRunCmd(ctx, ioStreams, clients, o)
}

// Complete proceses any data that is needed before Run executes
func (o *BuildRunOptions) Complete(args []string) error {

	// resource using GenerateName, which will provice a unique instance
	o.BuildRun = &buildv1alpha1.BuildRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", o.BuildName),
		},
		Spec: buildv1alpha1.BuildRunSpec{
			BuildRef: &buildv1alpha1.BuildRef{
				Name:       o.BuildRefName,
				APIVersion: o.BuildRefAPIVersion,
			},
			ServiceAccount: &buildv1alpha1.ServiceAccount{
				Name:     &o.ServiceAccountName,
				Generate: o.ServiceAccountGenerate,
			},
			Timeout: &o.Timeout,
			Output: &buildv1alpha1.Image{
				Credentials: &corev1.LocalObjectReference{},
				Image:       o.OutputImage,
			},
		},
	}

	if len(o.OutputCredentialsSecretName) != 0 {
		o.BuildRun.Spec.Output.Credentials = &corev1.LocalObjectReference{
			Name: o.OutputCredentialsSecretName,
		}
	}

	return nil
}

// Run executes the command logic
func (o *BuildRunOptions) Run() error {
	o.LogTailStartedOn = make(map[string]bool)

	o.LogTail = tail.NewTail(o.Context, o.Clients.KubernetesClientSet)

	br, err := o.Clients.ShipwrightClientSet.ShipwrightV1alpha1().BuildRuns(o.Clients.Namespace).Create(o.Context, o.BuildRun, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	if !o.FollowLogs {
		fmt.Fprintf(o.Streams.Out, "BuildRun %q created for build %q\n", br.GetName(), o.BuildName)
		return nil
	}

	o.BuildRunName = br.Name

	// instantiating a pod watcher with a specific label-selector to find the intended pod where the
	// actual build started by this subcommand is being executed, including the randomized buildrun
	// name
	listOpts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("build.shipwright.io/name=%s,buildrun.shipwright.io/name=%s", o.BuildName, br.GetName()),
	}

	if o.PodWatcher, err = reactor.NewPodWatcher(o.Context, o.Clients.KubernetesClientSet, listOpts, o.Clients.Namespace); err != nil {
		return err
	}

	o.PodWatcher.WithOnPodModifiedFn(o.onEvent)
	if _, err = o.PodWatcher.Start(); err != nil {
		return err
	}

	return nil
}

// tailLogs starts tailing logs for each container name in init-containers and containers
func (o *BuildRunOptions) tailLogs(pod *corev1.Pod) {
	containers := append(pod.Spec.InitContainers, pod.Spec.Containers...)
	for _, container := range containers {
		if _, exists := o.LogTailStartedOn[container.Name]; exists {
			continue
		}
		o.LogTailStartedOn[container.Name] = true
		o.LogTail.Start(pod.GetNamespace(), pod.GetName(), container.Name)
	}
}

// onEvent reacts to pod state changes, starting and stopping as needed
func (o *BuildRunOptions) onEvent(pod *corev1.Pod) error {
	// found more data races during unit testing with concurrent events coming in
	o.WatchLock.Lock()
	defer o.WatchLock.Unlock()
	switch pod.Status.Phase {
	case corev1.PodRunning:
		// graceful time to wait for container start
		time.Sleep(3 * time.Second)
		// start tailing container logs
		o.tailLogs(pod)
	case corev1.PodFailed:
		msg := ""
		br, err := o.Clients.ShipwrightClientSet.ShipwrightV1alpha1().BuildRuns(pod.Namespace).Get(o.Context, o.BuildRunName, metav1.GetOptions{})
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
		fmt.Fprint(o.Streams.Out, msg)
		o.stop()
		return err
	case corev1.PodSucceeded:
		fmt.Fprintf(o.Streams.Out, "Pod '%s' has succeeded!\n", pod.GetName())
		o.stop()
	default:
		fmt.Fprintf(o.Streams.Out, "Pod '%s' is in state %q...\n", pod.GetName(), string(pod.Status.Phase))
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

// stop invokes the stop command on streaming components.
func (o *BuildRunOptions) stop() {
	o.LogTail.Stop()
	o.PodWatcher.Stop()
}
