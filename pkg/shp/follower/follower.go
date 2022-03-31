package follower

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	buildclientset "github.com/shipwright-io/build/pkg/client/clientset/versioned"
	"github.com/shipwright-io/cli/pkg/shp/reactor"
	"github.com/shipwright-io/cli/pkg/shp/tail"
	"github.com/shipwright-io/cli/pkg/shp/util"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

// Follower encapsulate the function of tailing the logs for Pods derived from BuildRuns
type Follower struct {
	ctx            context.Context              // global context instance
	buildRun       types.NamespacedName         // qualified object name
	ioStreams      *genericclioptions.IOStreams // io-streams instance
	pw             *reactor.PodWatcher          // pod-watcher instance
	clientset      kubernetes.Interface         // kubernetes api-client
	buildClientset buildclientset.Interface     // shipwright api-client

	logTail         *tail.Tail      // follow container logs
	tailLogsStarted map[string]bool // controls tail instance per container

	logLock             sync.Mutex // avoiding race condition to print logs
	enteredRunningState bool       // target pod is running
}

// NewFollower returns a Follower instance.
func NewFollower(
	ctx context.Context,
	buildRun types.NamespacedName,
	ioStreams *genericclioptions.IOStreams,
	pw *reactor.PodWatcher,
	clientset kubernetes.Interface,
	buildClientset buildclientset.Interface,
) *Follower {
	f := &Follower{
		ctx:            ctx,
		buildRun:       buildRun,
		ioStreams:      ioStreams,
		pw:             pw,
		clientset:      clientset,
		buildClientset: buildClientset,

		logTail:         tail.NewTail(ctx, clientset),
		logLock:         sync.Mutex{},
		tailLogsStarted: map[string]bool{},
	}

	f.pw.WithOnPodModifiedFn(f.OnEvent)
	f.pw.WithTimeoutPodFn(f.OnTimeout)
	f.pw.WithNoPodEventsYetFn(f.OnNoPodEventsYet)

	return f
}

// GetLogLock returns the mutex used for coordinating access to log buffers.
func (f *Follower) GetLogLock() *sync.Mutex {
	return &f.logLock
}

// Log prints a message
func (f *Follower) Log(msg string) {
	// concurrent fmt.Fprintf(r.ioStream.Out...) calls need locking to avoid data races, as we 'write' to the stream
	f.logLock.Lock()
	defer f.logLock.Unlock()
	fmt.Fprint(f.ioStreams.Out, msg)
}

// tailLogs start tailing logs for each container name in init-containers and containers, if not
// started already.
func (f *Follower) tailLogs(pod *corev1.Pod) {
	containers := append(pod.Spec.InitContainers, pod.Spec.Containers...)
	for _, container := range containers {
		if _, exists := f.tailLogsStarted[container.Name]; exists {
			continue
		}
		f.tailLogsStarted[container.Name] = true
		f.logTail.Start(pod.GetNamespace(), pod.GetName(), container.Name)
	}
}

// Stop stop log tail instance.
func (f *Follower) Stop() {
	f.logTail.Stop()
	f.pw.Stop()
}

// OnEvent reacts on pod state changes, to start and stop tailing container logs.
func (f *Follower) OnEvent(pod *corev1.Pod) error {
	switch pod.Status.Phase {
	case corev1.PodRunning:
		if !f.enteredRunningState {
			f.Log(fmt.Sprintf("Pod %q in %q state, starting up log tail", pod.GetName(), corev1.PodRunning))
			f.enteredRunningState = true
			// graceful time to wait for container start
			time.Sleep(3 * time.Second)
			// start tailing container logs
			f.tailLogs(pod)
		}
	case corev1.PodFailed:
		msg := ""
		var br *buildv1alpha1.BuildRun
		err := wait.PollImmediate(1*time.Second, 15*time.Second, func() (done bool, err error) {
			brClient := f.buildClientset.ShipwrightV1alpha1().BuildRuns(pod.Namespace)
			br, err = brClient.Get(f.ctx, f.buildRun.Name, metav1.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					return true, nil
				}
				f.Log(fmt.Sprintf("error getting buildrun %q for pod %q: %s\n", f.buildRun.Name, pod.GetName(), err.Error()))
				return false, nil
			}
			if br.IsDone() {
				return true, nil
			}
			return false, nil
		})
		if err != nil {
			f.Log(fmt.Sprintf("gave up trying to get a buildrun %q in a terminal state for pod %q, proceeding with pod failure processing", f.buildRun.Name, pod.GetName()))
		}
		switch {
		case br == nil:
			msg = fmt.Sprintf("BuildRun %q has been deleted.\n", br.Name)
		case err == nil && br.IsCanceled():
			msg = fmt.Sprintf("BuildRun %q has been canceled.\n", br.Name)
		case (err == nil && br.DeletionTimestamp != nil) || (err != nil && kerrors.IsNotFound(err)):
			msg = fmt.Sprintf("BuildRun %q has been deleted.\n", br.Name)
		case pod.DeletionTimestamp != nil:
			msg = fmt.Sprintf("Pod %q has been deleted.\n", pod.GetName())
		default:
			msg = fmt.Sprintf("Pod %q has failed!\n", pod.GetName())
			podBytes, err2 := json.MarshalIndent(pod, "", "    ")
			if err2 == nil {
				msg = fmt.Sprintf("Pod %q has failed!\nPod JSON:\n%s\n", pod.GetName(), string(podBytes))
			}
			err = fmt.Errorf("build pod %q has failed", pod.GetName())
		}
		// see if because of deletion or cancelation
		f.Log(msg)
		f.Stop()
		return err
	case corev1.PodSucceeded:
		// encountered scenarios where the build run quickly enough that the pod effectively skips the running state,
		// or the events come in reverse order, and we never enter the tail
		if !f.enteredRunningState {
			f.Log(fmt.Sprintf("succeeded event for pod %q arrived before or in place of running event so dumping logs now", pod.GetName()))
			var b strings.Builder
			for _, c := range pod.Spec.Containers {
				logs, err := util.GetPodLogs(f.ctx, f.clientset, *pod, c.Name)
				if err != nil {
					f.Log(fmt.Sprintf("could not get logs for container %q: %s", c.Name, err.Error()))
					continue
				}
				fmt.Fprintf(&b, "*** Pod %q, container %q: ***\n\n", pod.Name, c.Name)
				fmt.Fprintln(&b, logs)
			}
			f.Log(b.String())
		}
		f.Log(fmt.Sprintf("Pod %q has succeeded!\n", pod.GetName()))
		f.Stop()
	default:
		f.Log(fmt.Sprintf("Pod %q is in state %q...\n", pod.GetName(), string(pod.Status.Phase)))
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

// OnTimeout reacts to either the context or request timeout causing the pod watcher to exit
func (f *Follower) OnTimeout(msg string) {
	f.Log(fmt.Sprintf("BuildRun %q log following has stopped because: %q\n", f.buildRun.Name, msg))
}

// OnNoPodEventsYet reacts to the pod watcher telling us it has not received any pod events for our build run
func (f *Follower) OnNoPodEventsYet(podList *corev1.PodList) {
	f.Log(fmt.Sprintf("BuildRun %q log following has not observed any pod events yet.\n", f.buildRun.Name))
	if podList != nil && len(podList.Items) > 0 {
		f.Log(fmt.Sprintf("BuildRun %q's Pod completed before the log following's watch was established.\n", f.buildRun.Name))
		f.OnEvent(&podList.Items[0])
		return
	}
	brClient := f.buildClientset.ShipwrightV1alpha1().BuildRuns(f.buildRun.Namespace)
	br, err := brClient.Get(f.ctx, f.buildRun.Name, metav1.GetOptions{})
	if err != nil {
		f.Log(fmt.Sprintf("error accessing BuildRun %q: %s", f.buildRun.Name, err.Error()))
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
		f.Log(fmt.Sprintf("BuildRun '%s' has been marked as failed.\n", br.Name))
	}
	if giveUp {
		f.Log(msg)
		f.Log(fmt.Sprintf("exiting 'shp build run --follow' for BuildRun %q", br.Name))
		f.Stop()
	}
}

// Start initiates the log following for the referenced BuildRun's Pod
func (f *Follower) Start(lo metav1.ListOptions) (*corev1.Pod, error) {
	return f.pw.Start(lo)
}
