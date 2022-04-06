package reactor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/shipwright-io/cli/pkg/shp/params"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// PodWatcher a simple function orchestrator based on watching a given pod and reacting upon the
// state modifications, should work as a helper to build business logic based on the build POD
// changes.
type PodWatcher struct {
	ctx context.Context

	clientset         kubernetes.Interface // kubernetes client instance
	ns                string               // target namespace
	requestTimeout    time.Duration        // kubernetes client request timeout
	watchEventTimeout time.Duration        // how long wait for the watch event

	stopCh   chan bool  // stops the event loop execution
	stopLock sync.Mutex // stopping muttex
	stopped  bool       // pod watcher stopped

	skipPodFns           []SkipPodFn
	onPodAddedFns        []OnPodEventFn
	onPodModifiedFns     []OnPodEventFn
	onPodDeletedFns      []OnPodEventFn
	watchEventTimeoutFns []WatchEventTimeoutFn
	timeoutPodFns        []TimeoutPodFn
}

// SkipPodFn a given pod instance is informed and expects a boolean as return. When true is returned
// this container state processing is skipped completely.
type SkipPodFn func(pod *corev1.Pod) bool

// OnPodEventFn when a pod is modified this method handles the event.
type OnPodEventFn func(pod *corev1.Pod) error

// WatchEventTimeoutFn when there is no watch event.
type WatchEventTimeoutFn func(pod *corev1.Pod) (bool, error)

// TimeoutPodFn when either the context or request timeout expires before the Pod finishes.
type TimeoutPodFn func()

var ErrTimeout = errors.New("reached timeout without pods")

// WithSkipPodFn sets the skip function instance.
func (p *PodWatcher) WithSkipPodFn(fn SkipPodFn) *PodWatcher {
	p.skipPodFns = append(p.skipPodFns, fn)
	return p
}

// WithOnPodAddedFn sets the function executed when a pod is added.
func (p *PodWatcher) WithOnPodAddedFn(fn OnPodEventFn) *PodWatcher {
	p.onPodAddedFns = append(p.onPodAddedFns, fn)
	return p
}

// WithOnPodModifiedFn sets the function executed when a pod is modified.
func (p *PodWatcher) WithOnPodModifiedFn(fn OnPodEventFn) *PodWatcher {
	p.onPodModifiedFns = append(p.onPodModifiedFns, fn)
	return p
}

// WithOnPodDeletedFn sets the function executed when a pod is modified.
func (p *PodWatcher) WithOnPodDeletedFn(fn OnPodEventFn) *PodWatcher {
	p.onPodDeletedFns = append(p.onPodDeletedFns, fn)
	return p
}

// WithTimeoutPodFn sets the function executed when the context or request timeout fires
func (p *PodWatcher) WithTimeoutPodFn(fn TimeoutPodFn) *PodWatcher {
	p.timeoutPodFns = append(p.timeoutPodFns, fn)
	return p
}

func (p *PodWatcher) WithWatchEventTimeoutFn(fn WatchEventTimeoutFn) *PodWatcher {
	p.watchEventTimeoutFns = append(p.watchEventTimeoutFns, fn)
	return p
}

// HandleEvent applies user informed functions against informed pod and event.
func (p *PodWatcher) HandleEvent(pod *corev1.Pod, event watch.Event) error {
	switch event.Type {
	case watch.Added:
		for _, fn := range p.onPodAddedFns {
			if err := fn(pod); err != nil {
				return err
			}
		}
	case watch.Modified:
		for _, fn := range p.onPodModifiedFns {
			if err := fn(pod); err != nil {
				return err
			}
		}
	case watch.Deleted:
		for _, fn := range p.onPodDeletedFns {
			if err := fn(pod); err != nil {
				return err
			}
		}
	}
	return nil
}

// Start runs the event loop based on a watch instantiated against informed pod. In case of errors
// the loop is interrupted.
func (p *PodWatcher) Start(listOpts metav1.ListOptions) (*corev1.Pod, error) {
	watcher, err := p.clientset.CoreV1().Pods(p.ns).Watch(p.ctx, listOpts)
	if err != nil {
		return nil, err
	}

	watchEventTimer := time.NewTicker(p.watchEventTimeout)
	requestTimeoutTimer := time.NewTicker(p.requestTimeout)

	for {
		select {
		// handling the regular pod events, will trigger calling event functions accordinly,
		// depending on the event type
		case event := <-watcher.ResultChan():
			if event.Object == nil {
				continue
			}
			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				continue
			}

			// stopping to watch for event timeout, since a event has already arrived here, meaning a
			// pod was found using ListOptions informed, and the same for the request timeout
			watchEventTimer.Stop()
			requestTimeoutTimer.Stop()

			if len(p.skipPodFns) > 0 {
				skip := false
				for _, fn := range p.skipPodFns {
					if fn(pod) {
						skip = true
						break
					}
				}
				if skip {
					continue
				}
			}

			if err := p.HandleEvent(pod, event); err != nil {
				return pod, err
			}

		// when the global context is interrupted, it will be dealt in the same way than timeouts
		case <-p.ctx.Done():
			watcher.Stop()
			for _, fn := range p.timeoutPodFns {
				fn()
			}
			return nil, nil

		// when there is no event coming through the watch, it will trigger the watchEventTimeout
		// functions, in case of error or returning `true`, it interrupts this select
		case <-watchEventTimer.C:
			if len(p.watchEventTimeoutFns) == 0 {
				continue
			}

			podList, err := p.clientset.CoreV1().Pods(p.ns).List(p.ctx, listOpts)
			if err != nil {
				return nil, err
			}
			for i := range podList.Items {
				for _, fn := range p.watchEventTimeoutFns {
					stop, err := fn(&podList.Items[i])
					if err != nil {
						return nil, err
					}
					if stop {
						watcher.Stop()
						return nil, nil
					}
				}
			}

		// when the Kubernetes client request has reach timeout
		case <-requestTimeoutTimer.C:
			watcher.Stop()
			for _, fn := range p.timeoutPodFns {
				fn()
			}
			return nil, fmt.Errorf("%w: after %.1fs", ErrTimeout, p.requestTimeout.Seconds())

		// watching over stop channel to stop the event loop on demand
		case <-p.stopCh:
			watcher.Stop()
			return nil, nil
		}
	}
}

// Stop closes the stop channel, and stops the execution loop.
func (p *PodWatcher) Stop() {
	// employ sync because of observed 'panic: close of closed channel' when running build run log
	// following along with canceling of builds
	p.stopLock.Lock()
	defer p.stopLock.Unlock()

	if !p.stopped {
		close(p.stopCh)
		p.stopped = true
	}
}

// NewPodWatcher instantiate PodWatcher event-loop.
func NewPodWatcher(
	ctx context.Context,
	clientset kubernetes.Interface,
	ns string,
	watchEventTimeout time.Duration,
	requestTimeout time.Duration,
) *PodWatcher {
	return &PodWatcher{
		ctx: ctx,

		clientset:         clientset,
		ns:                ns,
		requestTimeout:    requestTimeout,
		watchEventTimeout: watchEventTimeout,

		stopCh:   make(chan bool),
		stopLock: sync.Mutex{},
	}
}

// NewPodWatcherFromParams instantiate the PodWatcher based on the informed Params.
func NewPodWatcherFromParams(ctx context.Context, p params.Interface) (*PodWatcher, error) {
	clientset, err := p.ClientSet()
	if err != nil {
		return nil, err
	}
	watchEventTimeout := 10 * time.Second
	requestTimeout, err := p.RequestTimeout()
	if err != nil {
		return nil, err
	}

	return NewPodWatcher(ctx, clientset, p.Namespace(), watchEventTimeout, requestTimeout), nil
}
