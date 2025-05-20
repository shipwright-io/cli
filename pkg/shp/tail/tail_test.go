package tail

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	o "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_Tail(t *testing.T) {
	g := o.NewWithT(t)

	name := "pod"
	containerName := "container"
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: metav1.NamespaceDefault,
			Name:      name,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name: containerName,
			}},
		},
	}
	clientset := fake.NewSimpleClientset(pod)

	logTail := NewTail(context.TODO(), clientset)

	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()
	logTail.SetStdout(stdoutWriter)
	logTail.SetStderr(stderrWriter)

	logTail.Start(metav1.NamespaceDefault, "pod", "container")

	// graceful waiting for possible output written, both stdout and stderr
	time.Sleep(10 * time.Second)
	logTail.Stop()

	g.Expect(stdoutWriter.Close()).To(o.Succeed())
	g.Expect(stderrWriter.Close()).To(o.Succeed())

	// reading out stdout and stderr output by copying the iWriter contents to an intermediary
	// buffer and checking how many bytes were subject to copying.
	var buf bytes.Buffer

	stdoutNumBytes, err := io.Copy(&buf, stdoutReader)
	g.Expect(err).To(o.BeNil())
	g.Expect(stdoutNumBytes).To(o.Equal(int64(0)))

	stderrNumBytes, err := io.Copy(&buf, stderrReader)
	g.Expect(err).To(o.BeNil())
	g.Expect(stderrNumBytes).To(o.Equal(int64(0)))
}
