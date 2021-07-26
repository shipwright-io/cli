package buildrun

import (
	"strings"
	"testing"

	"github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/cli/pkg/shp/params"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/spf13/cobra"
)

func TestStreamBuildLogs(t *testing.T) {
	name := "test-obj"
	pod := &corev1.Pod{}
	pod.Name = name
	pod.Namespace = metav1.NamespaceDefault
	pod.Labels = map[string]string{
		v1alpha1.LabelBuildRun: name,
	}
	pod.Spec.Containers = []corev1.Container{
		{
			Name: name,
		},
	}

	cmd := LogsCommand{cmd: &cobra.Command{}}
	cmd.name = name
	// set up context
	cmd.Cmd().ExecuteC()

	clientset := fake.NewSimpleClientset(pod)
	ioStreams, _, out, _ := genericclioptions.NewTestIOStreams()
	param := params.NewParamsForTest(clientset, nil, nil, metav1.NamespaceDefault)
	err := cmd.Run(param, &ioStreams)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	if !strings.Contains(out.String(), "fake logs") {
		t.Fatalf("unexpected output: %s", out.String())
	}

	t.Logf("%s", out.String())

}
