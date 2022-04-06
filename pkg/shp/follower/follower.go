package follower

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/reactor"
	"github.com/shipwright-io/cli/pkg/shp/tail"
	"github.com/shipwright-io/cli/pkg/shp/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

// PodLogsFollower uses PodWatcher to react upon pod state changes with the final objective of
// streaming all logs from all the containers in the pod.
type PodLogsFollower struct {
	ctx context.Context

	pw        *reactor.PodWatcher          // podwatcher instance
	clientset kubernetes.Interface         // kubernetes client
	ioStreams *genericclioptions.IOStreams // log output interface

	podIsRunning   bool            // pod is set as running
	logTail        *tail.Tail      // helper to tail container logs
	logTailStarted map[string]bool // containers with logTail started

	onlyOnce bool // retrieve the pod logs only once
}

var (
	// ErrPodDeleted a pod has been deleted, having DeletionTimestamp set
	ErrPodDeleted = errors.New("pod has been deleted")

	// ErrPodFailed a pod has failed, generic reason
	ErrPodFailed = errors.New("pod has failed")

	// ErrPodStatusUnknown a pod has status condition set to unknown
	ErrPodStatusUnknown = errors.New("pod status is unknown")
)

// log log entries on the standard output of the ioStreams instance.
func (p *PodLogsFollower) log(format string, a ...interface{}) {
	fmt.Fprintf(p.ioStreams.Out, format, a...)
}

// log log entries on the error output of the ioStreams instance.
func (p *PodLogsFollower) logError(format string, a ...interface{}) {
	fmt.Fprintf(p.ioStreams.ErrOut, format, a...)
}

// onPodRunning as soon as the pod gets into the running state it will start streaming logs from all
// containers, and mark it as running.
func (p *PodLogsFollower) onPodRunning(pod *corev1.Pod) {
	if p.onlyOnce {
		p.getPodLogs(pod)
		return
	}
	// when the pod is already running the log tailing must be already set up, thefore skipping it
	// when dealing with subsequent pod modifications
	if p.podIsRunning {
		return
	}

	p.log("Pod %q in %q state, starting up log tail\n", pod.GetName(), corev1.PodRunning)

	time.Sleep(3 * time.Second)
	containers := append(pod.Spec.InitContainers, pod.Spec.Containers...)
	for _, container := range containers {
		if _, exists := p.logTailStarted[container.Name]; exists {
			continue
		}
		p.logTailStarted[container.Name] = true
		p.logTail.Start(pod.GetNamespace(), pod.GetName(), container.Name)
	}
	p.podIsRunning = true
}

// getPodLogs retrieve the current snapshot of the pod logs, without following it.
func (p *PodLogsFollower) getPodLogs(pod *corev1.Pod) {
	var buffer strings.Builder
	containers := append(pod.Spec.InitContainers, pod.Spec.Containers...)
	for _, container := range containers {
		logs, err := util.GetPodLogs(p.ctx, p.clientset, *pod, container.Name)
		if err != nil {
			p.logError("Error trying to get logs from container %q: %q", container.Name, err)
			continue
		}
		fmt.Fprintf(&buffer, "*** Pod %q, container %q: ***\n\n", pod.Name, container.Name)
		fmt.Fprintln(&buffer, logs)
	}
	p.log(buffer.String())

	if p.onlyOnce {
		p.Stop()
	}
}

// onPodFailed as soon as the pod fails, it will determine why and prepare the error accordingly.
func (p *PodLogsFollower) onPodFailed(pod *corev1.Pod) error {
	if pod.DeletionTimestamp != nil {
		p.log("Pod %q has been deleted!\n", pod.GetName())
		return fmt.Errorf("%w: %s", ErrPodDeleted, pod.GetName())
	}
	p.getPodLogs(pod)
	p.log("Pod %q has failed!\n", pod.GetName())
	return fmt.Errorf("%w: %s", ErrPodFailed, pod.GetName())
}

// onPodSucceeded as soon as the pod succeeds it checks if it was running before, otherwise fetch all
// logs at once.
func (p *PodLogsFollower) onPodSucceeded(pod *corev1.Pod) {
	if p.podIsRunning {
		p.log("Pod %q has succeeded!\n", pod.GetName())
		return
	}
	p.getPodLogs(pod)
}

// inspectPodStatus checks the informed pod status, trying to identify if it's on unknown condition.
func (p *PodLogsFollower) inspectPodStatus(pod *corev1.Pod) error {
	p.log("Pod %q is in state %q...\n", pod.GetName(), pod.Status.Phase)

	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodInitialized || condition.Type == corev1.ContainersReady {
			if condition.Status == corev1.ConditionUnknown {
				return fmt.Errorf("%w: %s", ErrPodStatusUnknown, pod.GetName())
			}
		}
	}
	return nil
}

// OnEvent reacts to the informed pod depending on its execution phase.
func (p *PodLogsFollower) OnEvent(pod *corev1.Pod) error {
	switch pod.Status.Phase {
	case corev1.PodRunning:
		p.onPodRunning(pod)
	case corev1.PodFailed:
		p.Stop()
		return p.onPodFailed(pod)
	case corev1.PodSucceeded:
		p.Stop()
		p.onPodSucceeded(pod)
	default:
		return p.inspectPodStatus(pod)
	}
	return nil
}

// WatchEventTimeout fallback mechanism to capture pods.
func (p *PodLogsFollower) WatchEventTimeout(pod *corev1.Pod) (bool, error) {
	return false, p.OnEvent(pod)
}

// Start watching for events.
func (p *PodLogsFollower) Start(listOpts metav1.ListOptions) (*corev1.Pod, error) {
	return p.pw.Start(listOpts)
}

// Stop watching for events and following logs.
func (p *PodLogsFollower) Stop() {
	p.logTail.Stop()
	p.pw.Stop()
}

// SetOnlyOnce retrieve the logs only once, as soon as the pod is able to provide logs.
func (p *PodLogsFollower) SetOnlyOnce() {
	p.onlyOnce = true
}

// NewPodLogsFollower instantiate the PodLogsFollower by setting up the PodWatcher event callbacks.
func NewPodLogsFollower(
	ctx context.Context,
	pw *reactor.PodWatcher,
	clientset kubernetes.Interface,
	ioStreams *genericclioptions.IOStreams,
) *PodLogsFollower {
	p := &PodLogsFollower{
		ctx: ctx,

		pw:        pw,
		clientset: clientset,
		ioStreams: ioStreams,

		logTail:        tail.NewTail(ctx, clientset),
		logTailStarted: map[string]bool{},
	}
	// when starting a watch against stale pods, those are reported as "added" event, and possibly
	// won't be subject to modifications
	pw.WithOnPodAddedFn(p.OnEvent)
	// when watching over pods in execution, as the build workflow progresses it will issue
	// modification events
	pw.WithOnPodModifiedFn(p.OnEvent)
	// when watching over pods subject to deletion
	pw.WithOnPodDeletedFn(p.OnEvent)
	// when watch does not issue events after a certain period
	pw.WithWatchEventTimeoutFn(p.WatchEventTimeout)
	return p
}

// NewPodLogsFollowerFromParams instantiate the PodLogsFollower based on the Params instance.
func NewPodLogsFollowerFromParams(
	ctx context.Context,
	p params.Interface,
	pw *reactor.PodWatcher,
	ioStreams *genericclioptions.IOStreams,
) (*PodLogsFollower, error) {
	clientset, err := p.ClientSet()
	if err != nil {
		return nil, err
	}
	return NewPodLogsFollower(ctx, pw, clientset, ioStreams), nil
}
