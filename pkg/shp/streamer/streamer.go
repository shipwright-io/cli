package streamer

import (
	"io"
	"os"
	"sync"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/cmd/exec"
	"k8s.io/kubectl/pkg/util/interrupt"
)

// Streamer represents the actor that streams data onto a POD, running on Kubernetes. It does so via
// `kubectl exec` to run `tar` on target container, and it redirects stdin to upload the local data.
type Streamer struct {
	restConfig     *rest.Config         // rest API client configuration
	clientset      kubernetes.Interface // kubernetes client
	remoteExecutor exec.RemoteExecutor  // overwritten during testing
}

// WriterFn exposes the writer interface, receives the data to be streamed.
type WriterFn func(w io.Writer) error

// tarCmd base tar command to be executed on the POD, a target directory should be appended.
var tarCmd = []string{"tar", "xfv", "-", "-C"}

// doneCmd command to notify the container the data streaming is done, thus the container build
// process can continue.
var doneCmd = []string{"waiter", "done"}

// execute the informed exec command, by invokeing Validate and Run methods.
func (s *Streamer) execute(opts *exec.ExecOptions) error {
	if err := opts.Validate(); err != nil {
		return err
	}
	return opts.Run()
}

// Stream the data onto the informed target, and it uses the BaseDir as the path to store the data on
// the running POD. The writerFn is employed to expose the writer interface to callers.
func (s *Streamer) Stream(target *Target, writerFn WriterFn) error {
	var wg sync.WaitGroup
	wg.Add(1)

	// using a IO pipe redirect to collect all data written to the writter interface and stream it to
	// the reader end. Additionally creating a error channel to receive either error or nil, at the
	// end of writerFn execution
	reader, writer := io.Pipe()
	errCh := make(chan error, 1)
	defer close(errCh)

	go func() {
		defer writer.Close()
		errCh <- writerFn(writer)
		wg.Done()
	}()

	// defines the target pod using namespace and pod name, and wires up the local stdin with the
	// pipe reader interface, therefore all data written on the writer interface will be redirected
	// to the pod
	streamOpts := exec.StreamOptions{
		Namespace:     target.Namespace,
		PodName:       target.Pod,
		ContainerName: target.Container,
		Stdin:         true,
		IOStreams: genericclioptions.IOStreams{
			In:     reader,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
	}
	// creates the equivalent of "kubectl exec" structure, plus the stdin redirect, and then runs the
	// predefined ar command on the pod to receive the data stream
	execOpts := &exec.ExecOptions{
		StreamOptions: streamOpts,
		Config:        s.restConfig,
		PodClient:     s.clientset.CoreV1(),
		Command:       append(tarCmd, target.BaseDir),
		Executor:      s.remoteExecutor,
	}
	if err := s.execute(execOpts); err != nil {
		return err
	}

	// blocking the execution, waiting for writerFn to return either error or nil
	wg.Wait()
	return <-errCh
}

// Done uses "kubectl exec" to run an command on target container, notifying the upload is done.
func (s *Streamer) Done(target *Target) error {
	streamOpts := exec.StreamOptions{
		Namespace:       target.Namespace,
		PodName:         target.Pod,
		ContainerName:   target.Container,
		InterruptParent: &interrupt.Handler{},
		IOStreams: genericclioptions.IOStreams{
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
	}
	execOpts := &exec.ExecOptions{
		StreamOptions: streamOpts,
		Config:        s.restConfig,
		PodClient:     s.clientset.CoreV1(),
		Command:       doneCmd,
		Executor:      s.remoteExecutor,
	}
	return s.execute(execOpts)
}

// NewStreamer instantiate Streamer.
func NewStreamer(restConfig *rest.Config, clientset kubernetes.Interface) *Streamer {
	return &Streamer{
		restConfig:     restConfig,
		clientset:      clientset,
		remoteExecutor: &exec.DefaultRemoteExecutor{},
	}
}
