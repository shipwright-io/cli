package reactor

import (
	"context"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

const (
	ContextTimeoutMessage = "context deadline has been exceeded"
	RequestTimeoutMessage = "request timeout has expired"
)

// PodWatcher a simple function orchestrator based on watching a given pod and reacting upon the
// state modifications, should work as a helper to build business logic based on the build POD
// changes.
type PodWatcher struct {
	ctx         context.Context
	to          time.Duration
	stopCh      chan bool // stops the event loop execution
	stopLock    sync.Mutex
	stopped     bool
	eventTicker *time.Ticker
	clientset kubernetes.Interface
	listOpts  metav1.ListOptions
	ns string
	watcher     watch.Interface // client watch instance

	noPodEventsYetFn NoPodEventsYetFn
	toPodFn          TimeoutPodFn
	skipPodFn        SkipPodFn
	onPodAddedFn     OnPodEventFn
	onPodModifiedFn  OnPodEventFn
	onPodDeletedFn   OnPodEventFn
}

// SkipPodFn a given pod instance is informed and expects a boolean as return. When true is returned
// this container state processing is skipped completely.
type SkipPodFn func(pod *corev1.Pod) bool

// OnPodEventFn when a pod is modified this method handles the event.
type OnPodEventFn func(pod *corev1.Pod) error

// TimeoutPodFn when either the context or request timeout expires before the Pod finishes
type TimeoutPodFn func(msg string)

// NoPodEventsYetFn when the watch has not received the create event within a reasonable time
type NoPodEventsYetFn func()

// WithSkipPodFn sets the skip function instance.
func (p *PodWatcher) WithSkipPodFn(fn SkipPodFn) *PodWatcher {
	p.skipPodFn = fn
	return p
}

// WithOnPodAddedFn sets the function executed when a pod is added.
func (p *PodWatcher) WithOnPodAddedFn(fn OnPodEventFn) *PodWatcher {
	p.onPodAddedFn = fn
	return p
}

// WithOnPodModifiedFn sets the function executed when a pod is modified.
func (p *PodWatcher) WithOnPodModifiedFn(fn OnPodEventFn) *PodWatcher {
	p.onPodModifiedFn = fn
	return p
}

// WithOnPodDeletedFn sets the function executed when a pod is modified.
func (p *PodWatcher) WithOnPodDeletedFn(fn OnPodEventFn) *PodWatcher {
	p.onPodDeletedFn = fn
	return p
}

// WithTimeoutPodFn sets the function executed when the context or request timeout fires
func (p *PodWatcher) WithTimeoutPodFn(fn TimeoutPodFn) *PodWatcher {
	p.toPodFn = fn
	return p
}

// WithNoPodEventsYetFn sets the function executed when the watcher decides it has waited long enough for the first event
func (p *PodWatcher) WithNoPodEventsYetFn(fn NoPodEventsYetFn) *PodWatcher {
	p.noPodEventsYetFn = fn
	return p
}

// handleEvent applies user informed functions against informed pod and event.
func (p *PodWatcher) handleEvent(pod *corev1.Pod, event watch.Event) error {
	//p.stopLock.Lock()
	//defer p.stopLock.Unlock()
	p.eventTicker.Stop()
	switch event.Type {
	case watch.Added:
		if p.onPodAddedFn != nil {
			if err := p.onPodAddedFn(pod); err != nil {
				return err
			}
		}
	case watch.Modified:
		if p.onPodModifiedFn != nil {
			if err := p.onPodModifiedFn(pod); err != nil {
				return err
			}
		}
	case watch.Deleted:
		if p.onPodDeletedFn != nil {
			if err := p.onPodDeletedFn(pod); err != nil {
				return err
			}
		}
	}
	return nil
}

// Start runs the event loop based on a watch instantiated against informed pod. In case of errors
// the loop is interrupted.
func (p *PodWatcher) Start(listOpts metav1.ListOptions) (*corev1.Pod, error) {
	w, err := p.clientset.CoreV1().Pods(p.ns).Watch(p.ctx, listOpts)
	if err != nil {
		return nil, err
	}
	p.watcher = w
	for {
		select {
		// handling the regular pod modification events, which should trigger calling event functions
		// accordinly
		case event := <-p.watcher.ResultChan():
			if event.Object == nil {
				continue
			}
			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				continue
			}

			if p.skipPodFn != nil && p.skipPodFn(pod) {
				continue
			}
			if err := p.handleEvent(pod, event); err != nil {
				return pod, err
			}
		// watching over global context, when done is informed on the context it needs to reflect on
		// the event loop as well.
		case <-p.ctx.Done():
			p.watcher.Stop()
			if p.toPodFn != nil {
				p.toPodFn(ContextTimeoutMessage)
			}
			return nil, nil

		// handle k8s --request-timeout setting, converted to time.Duration, that is passed down to PodWatcher;
		// if we have exceeded it, we exit
		case <-time.After(p.to):
			p.watcher.Stop()
			if p.toPodFn != nil {
				p.toPodFn(RequestTimeoutMessage)
			}
			return nil, nil

		// deal with case where a lack of any pod event means there is some sort of issue;
		// we let the called function decide whether to stop the watch
		// NOTE: a k8s event watch coupled with our pod watch proved problematic with unit tests; also, with
		// a lot of the relevant constants in github.com/k8s/k8s, which is a hassle to vendor in, prototypes
		// felt fragile
		case <-p.eventTicker.C:
			if p.noPodEventsYetFn != nil {
				p.noPodEventsYetFn()
			}

		// watching over stop channel to stop the event loop on demand.
		case <-p.stopCh:
			p.watcher.Stop()
			return nil, nil
		}
	}
}

// Stop closes the stop channel, and stops the execution loop.
func (p *PodWatcher) Stop() {
	// employ sync because of observed 'panic: close of closed channel' when running build run log following
	// along with canceling of builds
	p.stopLock.Lock()
	defer p.stopLock.Unlock()
	p.eventTicker.Stop()
	if !p.stopped {
		close(p.stopCh)
		p.stopped = true
	}
}

// NewPodWatcher instantiate PodWatcher event-loop.
func NewPodWatcher(
	ctx context.Context,
	timeout time.Duration,
	clientset kubernetes.Interface,
	ns string,
) (*PodWatcher, error) {
	//TODO don't think the have not received events yet ticker needs to be tunable, but leaving a TODO for now while we get feedback
	return &PodWatcher{ctx: ctx, to: timeout, ns: ns, clientset: clientset, eventTicker: time.NewTicker(1 * time.Second), stopCh: make(chan bool), stopLock: sync.Mutex{}}, nil
}
