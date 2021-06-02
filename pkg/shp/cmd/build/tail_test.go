package build

import (
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/onsi/gomega"
	o "github.com/onsi/gomega"
	tknopt "github.com/tektoncd/cli/pkg/options"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"

	tkncli "github.com/tektoncd/cli/pkg/cli"
	tknpodlogfake "github.com/tektoncd/cli/pkg/pods/fake"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	tknfake "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/fake"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	dfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func setupFakeClientsAndObjects() (tkncli.Params, string, string, string) {
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
	tr := &v1beta1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: metav1.NamespaceDefault,
			Name:      trName,
		},
		Status: v1beta1.TaskRunStatus{
			TaskRunStatusFields: v1beta1.TaskRunStatusFields{
				PodName: podName,
			},
		},
	}
	brName := "br"
	br := &buildv1alpha1.BuildRun{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: metav1.NamespaceDefault,
			Name:      brName,
		},
		Status: buildv1alpha1.BuildRunStatus{
			LatestTaskRunRef: &trName,
		},
	}
	scheme := runtime.NewScheme()
	buildv1alpha1.SchemeBuilder.AddToScheme(scheme)
	corev1.SchemeBuilder.AddToScheme(scheme)
	v1beta1.AddToScheme(scheme)
	dclientset := dfake.NewSimpleDynamicClient(scheme, pod, tr, br)
	dclientset.Resource(v1beta1.SchemeGroupVersion.WithResource("taskruns")).Namespace(metav1.NamespaceDefault)
	tclientset := tknfake.NewSimpleClientset(tr)
	kclientset := fake.NewSimpleClientset(pod)
	testTKNParams := &TestTKNParams{fakeClients: &tkncli.Clients{
		Dynamic: dclientset,
		Tekton:  tclientset,
		Kube:    kclientset,
	}}
	testTKNParams.SetNamespace(metav1.NamespaceDefault)
	apil := &metav1.APIResourceList{
		TypeMeta:     metav1.TypeMeta{},
		GroupVersion: v1beta1.SchemeGroupVersion.Identifier(),
		APIResources: nil,
	}
	apil.APIResources = append(apil.APIResources, metav1.APIResource{
		Name:         "taskruns",
		SingularName: "taskrun",
		Namespaced:   true,
		Group:        v1beta1.SchemeGroupVersion.Group,
		Version:      v1beta1.SchemeGroupVersion.Version,
		Kind:         "TaskRun",
	})
	tclientset.Resources = append(tclientset.Resources, apil)

	return testTKNParams, pod.Name, tr.Name, pod.Spec.Containers[0].Name
}

func setupFakeTaskRunPodLogs(testTKNParams tkncli.Params, ioStreams *genericclioptions.IOStreams, podName, taskRunName, containerName string) (*tknopt.LogOptions, string) {
	logData := "test logs"
	// the closest plug / override point we have the the tkn taskrun log following is the Streamer function;
	// tkn has a fake version for this that they use in log testing
	fakePodLog := tknpodlogfake.Log{
		PodName: podName,
		Containers: []tknpodlogfake.Container{
			{
				Name: containerName,
				Logs: []string{logData},
			},
		},
	}
	fakePodsLogs := tknpodlogfake.Logs(fakePodLog)
	logOpts := getTKNLogOpts(testTKNParams, ioStreams, taskRunName)
	// since we are doing live log streaming, this mimics tkn unit tests and aborting the unit
	// test quickly, but after enough time to pull the log data
	logOpts.ActivityTimeout = 250 * time.Millisecond
	logOpts.Streamer = tknpodlogfake.Streamer(fakePodsLogs)

	return logOpts, logData
}

func Test_Tail(t *testing.T) {
	ioStreams, _, outB, _ := genericclioptions.NewTestIOStreams()

	testTKNParams, podName, taskRunName, containerName := setupFakeClientsAndObjects()

	logOpts, logData := setupFakeTaskRunPodLogs(testTKNParams, &ioStreams, podName, taskRunName, containerName)

	err := Tail(logOpts)
	g := gomega.NewGomegaWithT(t)
	g.Expect(err).To(o.BeNil())
	g.Expect(outB.String()).To(o.ContainSubstring(logData))
}

type TestTKNParams struct {
	fakeClients *tkncli.Clients
	namespace   string
}

func (t *TestTKNParams) SetKubeConfigPath(path string) {

}

func (t *TestTKNParams) SetKubeContext(ctx string) {

}

func (t *TestTKNParams) Clients() (*tkncli.Clients, error) {
	return t.fakeClients, nil
}

func (t *TestTKNParams) KubeClient() (kubernetes.Interface, error) {
	return nil, nil
}

func (t *TestTKNParams) SetNamespace(ns string) {
	t.namespace = ns
}

func (t *TestTKNParams) Namespace() string {
	return t.namespace
}

func (t *TestTKNParams) SetNoColour(bool) {

}

func (t *TestTKNParams) Time() clockwork.Clock {
	return clockwork.NewFakeClock()
}
