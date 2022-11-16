package streamer

import (
	"io"
	"testing"

	"github.com/onsi/gomega"
	"github.com/shipwright-io/cli/test/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	o "github.com/onsi/gomega"
)

func Test_Streamer(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	podName := "pod"
	f := mock.NewFakeClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: metav1.NamespaceDefault,
			Name:      podName,
		},
	})

	restConfig := f.RESTConfig()
	clientset := f.Clientset()

	s := NewStreamer(restConfig, clientset)

	re := mock.NewFakeRemoteExecutor(nil)
	s.remoteExecutor = re

	targetPod := &Target{
		Namespace: metav1.NamespaceDefault,
		Pod:       podName,
		Container: "container",
		BaseDir:   "/",
	}

	var size int64 = 1000

	// streaming mocked standard input data, and asserting both command informed is expected, and
	// stdin is preserved
	stdin := "standard input"
	err := s.Stream(targetPod, func(w io.Writer) error {
		_, err := w.Write([]byte(stdin))
		return err
	}, size)
	g.Expect(err).To(o.BeNil())
	g.Expect(re.Command()).To(o.Equal([]string{"tar", "xfv", "-", "-C", "/"}))
	g.Expect(re.Stdin()).To(o.Equal(stdin))

	// calling out "done" command on target pod, and making sure the command informed is expected
	err = s.Done(targetPod)
	g.Expect(err).To(o.BeNil())
	g.Expect(re.Command()).To(o.Equal([]string{"waiter", "done"}))
}
