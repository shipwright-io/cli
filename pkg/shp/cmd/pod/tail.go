package pod

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/informers"
	k8s "k8s.io/client-go/kubernetes"
	typedv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

// logChannel represents data to write on log channel
type logChannel struct {
	taskRun string
	step    string
	log     string
}

type podStreamer struct {
	name       string
	namespace  string
	kubeClient k8s.Interface
	streamer   newStreamFunc
}

func (p *podStreamer) wait() (*corev1.Pod, error) {
	// ensure pod exists before we actually check for it
	if _, err := p.get(); err != nil {
		return nil, err
	}

	stopC := make(chan struct{})
	eventC := make(chan interface{})
	defer close(eventC)
	defer close(stopC)

	p.watcher(stopC, eventC)

	var pod *corev1.Pod
	var err error
	for e := range eventC {
		pod, err = checkPodStatus(e)
		if pod != nil || err != nil {
			break
		}
	}

	return pod, err
}

func (p *podStreamer) get() (*corev1.Pod, error) {
	return p.kubeClient.CoreV1().Pods(p.namespace).Get(context.Background(), p.name, metav1.GetOptions{})
}

type containerStreamer struct {
	name        string
	NewStreamer newStreamFunc
	pod         *podStreamer
}

func (c *containerStreamer) status() error {
	pod, err := c.pod.get()
	if err != nil {
		return err
	}

	container := c.name
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name != container {
			continue
		}

		if cs.State.Terminated != nil && cs.State.Terminated.ExitCode == 1 {
			msg := ""

			if cs.State.Terminated.Reason != "" && cs.State.Terminated.Reason != "Error" {
				msg = msg + " : " + cs.State.Terminated.Reason
			}

			if cs.State.Terminated.Message != "" && cs.State.Terminated.Message != "Error" {
				msg = msg + " : " + cs.State.Terminated.Message
			}

			return fmt.Errorf("container %s has failed %s", container, msg)
		}
	}

	for _, cs := range pod.Status.InitContainerStatuses {
		if cs.Name != container {
			continue
		}

		if cs.State.Terminated != nil && cs.State.Terminated.ExitCode == 1 {
			return fmt.Errorf("container %s has failed: %s", container, cs.State.Terminated.Reason)
		}
	}

	return nil
}

type podLog struct {
	podName       string
	containerName string
	log           string
}
type podLogReader struct {
	containerName string
	pod           *podStreamer
	follow        bool
}

func (c *containerStreamer) logReader(follow bool) *podLogReader {
	return &podLogReader{c.name, c.pod, follow}
}

func (lr *podLogReader) read() (<-chan podLog, <-chan error, error) {
	pod := lr.pod
	opts := &corev1.PodLogOptions{
		Follow:    lr.follow,
		Container: lr.containerName,
	}

	stream, err := pod.Stream(opts)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting logs for pod %s(%s) : %s", pod.name, lr.containerName, err)
	}

	logC := make(chan podLog)
	errC := make(chan error)

	go func() {
		defer close(logC)
		defer close(errC)
		defer stream.Close()

		r := bufio.NewReader(stream)
		for {
			line, _, err := r.ReadLine()

			if err != nil {
				if err != io.EOF {
					errC <- err
				}
				return
			}

			logC <- podLog{
				podName:       pod.name,
				containerName: lr.containerName,
				log:           string(line),
			}
		}
	}()

	return logC, errC, nil
}

func (p *podStreamer) containerStreamer(c string) *containerStreamer {
	return &containerStreamer{
		name:        c,
		pod:         p,
		NewStreamer: p.streamer,
	}
}

func (p *podStreamer) watcher(stopC <-chan struct{}, eventC chan<- interface{}) {
	factory := informers.NewSharedInformerFactoryWithOptions(
		p.kubeClient, time.Second*10,
		informers.WithNamespace(p.namespace))

	factory.Core().V1().Pods().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pod, _ := obj.(*corev1.Pod)
				if pod.Name != p.name {
					return
				}
				eventC <- obj
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				pod, _ := newObj.(*corev1.Pod)
				if pod.Name != p.name {
					return
				}
				eventC <- newObj
			},
			DeleteFunc: func(obj interface{}) {
				pod, _ := obj.(*corev1.Pod)
				if pod.Name != p.name {
					return
				}
				eventC <- obj
			},
		})

	factory.Start(stopC)
	factory.WaitForCacheSync(stopC)
}

func (p *podStreamer) Stream(opt *corev1.PodLogOptions) (io.ReadCloser, error) {
	pods := p.kubeClient.CoreV1().Pods(p.namespace)
	if pods == nil {
		return nil, fmt.Errorf("error getting pods")
	}

	return p.streamer(pods, p.name, opt).stream()
}

type step struct {
	name      string
	container string
	state     corev1.ContainerState
}

type podStream struct {
	name string
	pods typedv1.PodInterface
	opts *corev1.PodLogOptions
}

func (s *podStream) stream() (io.ReadCloser, error) {
	return s.pods.GetLogs(s.name, s.opts).Stream(context.Background())
}

type inputOutputStreams struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

type streamFunction interface {
	stream() (io.ReadCloser, error)
}

type newStreamFunc func(p typedv1.PodInterface, name string, o *corev1.PodLogOptions) streamFunction

func podOpts(name string) func(opts *metav1.ListOptions) {
	return func(opts *metav1.ListOptions) {
		opts.FieldSelector = fields.OneTermEqualSelector("metadata.podName", name).String()
	}
}

func checkPodStatus(obj interface{}) (*corev1.Pod, error) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return nil, fmt.Errorf("failed to cast to pod object")
	}

	if pod.DeletionTimestamp != nil {
		return pod, fmt.Errorf("failed to run the pod %s ", pod.Name)
	}

	if pod.Status.Phase == corev1.PodSucceeded ||
		pod.Status.Phase == corev1.PodRunning ||
		pod.Status.Phase == corev1.PodFailed {
		return pod, nil
	}

	// Handle any issues with pulling images that may fail
	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.PodInitialized || c.Type == corev1.ContainersReady {
			if c.Status == corev1.ConditionUnknown {
				return pod, fmt.Errorf(c.Message)
			}
		}
	}

	return nil, nil
}

func newPodStreamer(name, ns string, client k8s.Interface, streamer newStreamFunc) *podStreamer {
	return &podStreamer{
		name:       name,
		namespace:  ns,
		kubeClient: client,
		streamer:   streamer,
	}
}

func getSteps(pod *corev1.Pod) []*step {
	status := map[string]corev1.ContainerState{}
	for _, cs := range pod.Status.ContainerStatuses {
		status[cs.Name] = cs.State
	}

	steps := []*step{}
	for _, c := range pod.Spec.Containers {
		steps = append(steps, &step{
			name:      strings.TrimPrefix(c.Name, "step-"),
			container: c.Name,
			state:     status[c.Name],
		})
	}

	return steps
}

func getInitSteps(pod *corev1.Pod) []*step {
	status := map[string]corev1.ContainerState{}
	for _, ics := range pod.Status.InitContainerStatuses {
		status[ics.Name] = ics.State
	}

	steps := []*step{}
	for _, ic := range pod.Spec.InitContainers {
		steps = append(steps, &step{
			name:      strings.TrimPrefix(ic.Name, "step-"),
			container: ic.Name,
			state:     status[ic.Name],
		})
	}

	return steps
}

func readStepsLogs(logC chan<- logChannel, errC chan<- error, steps []*step, pod *podStreamer, taskRunName string) {
	for _, step := range steps {
		container := pod.containerStreamer(step.container)
		containerLogC, containerLogErrC, err := container.logReader(true).read()
		if err != nil {
			errC <- fmt.Errorf("error in getting logs for step %s: %s", step.name, err)
			continue
		}

		for containerLogC != nil || containerLogErrC != nil {
			select {
			case l, ok := <-containerLogC:
				if !ok {
					containerLogC = nil
					logC <- logChannel{taskRun: taskRunName, step: step.name, log: "EOFLOG"}
					continue
				}
				logC <- logChannel{taskRun: taskRunName, step: step.name, log: l.log}

			case e, ok := <-containerLogErrC:
				if !ok {
					containerLogErrC = nil
					continue
				}

				errC <- fmt.Errorf("failed to get logs for %s: %s", step.name, e)
			}
		}

		if err := container.status(); err != nil {
			errC <- err
			return
		}
	}
}

func (l *tail) readPodLogs(podName, taskRunName, namespace string, kube k8s.Interface) (<-chan logChannel, <-chan error) {
	logC := make(chan logChannel)
	errC := make(chan error)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		wg.Done()

		// wait for all goroutines to close before closing errC channel
		wg.Wait()
		close(errC)
	}()

	wg.Add(1)
	go func() {
		defer func() {
			close(logC)
			wg.Done()
		}()

		p := newPodStreamer(podName, namespace, kube, l.streamer)
		p.streamer(kube.CoreV1().Pods(namespace), podName, &corev1.PodLogOptions{}).stream()
		var pod *corev1.Pod
		var err error

		pod, err = p.wait()
		if err != nil {
			errC <- fmt.Errorf("pod %s failed: %s", podName, err.Error())
		}

		steps := []*step{}
		if pod != nil {
			steps = getInitSteps(pod)
			steps = append(steps, getSteps(pod)...)
		}
		readStepsLogs(logC, errC, steps, p, taskRunName)

	}()

	return logC, errC
}

func write(s inputOutputStreams, logC <-chan logChannel, errC <-chan error) {
	//NOTE: if we want color based formatting like with tkn, there is more code to pull in from tkn
	// specifically the github.com/tektoncd/cli/pkg/formatted package

	for logC != nil || errC != nil {
		select {
		case l, ok := <-logC:
			if !ok {
				logC = nil
				continue
			}

			if l.log == "EOFLOG" {
				fmt.Fprintf(s.Out, "\n")
				continue
			}

			fmt.Fprintf(s.Out, "[%s : %s] ", l.taskRun, l.step)

			fmt.Fprintf(s.Out, "%s\n", l.log)
		case e, ok := <-errC:
			if !ok {
				errC = nil
				continue
			}
			fmt.Fprintf(s.Err, "%s\n", e)

		}
	}

}

type tail struct {
	streamer newStreamFunc
	kube     k8s.Interface
}

func newTail(kube k8s.Interface, streamer newStreamFunc) *tail {
	return &tail{
		streamer: streamer,
		kube:     kube,
	}
}

func innerTail(podName, taskRunName, namespace string, kube k8s.Interface, ioStreams *genericclioptions.IOStreams, streamer newStreamFunc) {
	tail := newTail(kube, streamer)
	logC, errC := tail.readPodLogs(podName, taskRunName, namespace, kube)
	s := inputOutputStreams{
		In:  ioStreams.In,
		Out: ioStreams.Out,
		Err: ioStreams.ErrOut,
	}
	write(s, logC, errC)

}

func Tail(podName, taskRunName, namespace string, kube k8s.Interface, ioStreams *genericclioptions.IOStreams) {
	s := func(p typedv1.PodInterface, name string, o *corev1.PodLogOptions) streamFunction {
		s := podStream{
			name: name,
			pods: kube.CoreV1().Pods(namespace),
			opts: o,
		}
		return &s
	}

	innerTail(podName, taskRunName, namespace, kube, ioStreams, s)
}
