package pod

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/onsi/gomega"
	o "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	typedv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type fakeContainer struct {
	name string
	logs []string
}

func newFakeContainer(name string, logs ...string) fakeContainer {
	return fakeContainer{
		name: name,
		logs: logs,
	}
}

func fakePodLog(name string, containers ...fakeContainer) fakeLog {
	return fakeLog{
		podName:    name,
		containers: containers,
	}
}

type fakeLog struct {
	podName    string
	containers []fakeContainer
}

func fakeLogs(logs ...fakeLog) []fakeLog {
	ret := []fakeLog{}
	ret = append(ret, logs...)
	return ret
}

type fakePodStream struct {
	logs    []fakeLog
	pods    typedv1.PodInterface
	podName string
	opts    *corev1.PodLogOptions
}

func (ps *fakePodStream) stream() (io.ReadCloser, error) {
	for _, fl := range ps.logs {
		if fl.podName != ps.podName {
			continue
		}

		log := ""
		for _, c := range fl.containers {
			log = log + strings.Join(c.logs, "\n")
		}
		return ioutil.NopCloser(strings.NewReader(log)), nil

	}
	return nil, fmt.Errorf("failed to stream container logs")
}

func fakeStreamer(l []fakeLog) newStreamFunc {
	return func(p typedv1.PodInterface, name string, o *corev1.PodLogOptions) streamFunction {
		return &fakePodStream{
			logs:    l,
			pods:    p,
			podName: name,
			opts:    o,
		}
	}
}

func setupFakeClientsAndObjects() (kubernetes.Interface, string, string, string) {
	podName := "pod-for-tr"
	containerName := "container"
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: metav1.NamespaceDefault,
			Name:      podName,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name: containerName,
			}},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}
	trName := "tr-for-br"
	scheme := runtime.NewScheme()
	corev1.SchemeBuilder.AddToScheme(scheme)
	kclientset := fake.NewSimpleClientset(pod)

	return kclientset, pod.Name, trName, pod.Spec.Containers[0].Name
}

func setupFakeTaskRunPodLogs(podName, containerName string) (newStreamFunc, string) {
	logData := "test logs"
	fc := newFakeContainer(containerName, logData)
	fpl1 := fakePodLog(podName, fc)
	fls := fakeLogs(fpl1) //, fpl2)
	fs := fakeStreamer(fls)

	return fs, logData
}

func Test_Tail(t *testing.T) {
	ioStreams, _, outB, _ := genericclioptions.NewTestIOStreams()

	kube, podName, taskRunName, containerName := setupFakeClientsAndObjects()

	streamer, logData := setupFakeTaskRunPodLogs(podName, containerName)

	innerTail(podName, taskRunName, metav1.NamespaceDefault, kube, &ioStreams, streamer)
	g := gomega.NewGomegaWithT(t)
	g.Expect(outB.String()).To(o.ContainSubstring(logData))
}
