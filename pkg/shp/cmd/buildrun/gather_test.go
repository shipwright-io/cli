package buildrun

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	shpfake "github.com/shipwright-io/build/pkg/client/clientset/versioned/fake"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
)

func TestGatherTaskRun(t *testing.T) {
	name := "test-br-taskrun"
	namespace := "default"
	executorName := "test-tr"
	podName := "test-pod"

	buildRun := &buildv1beta1.BuildRun{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Status: buildv1beta1.BuildRunStatus{
			Executor: &buildv1beta1.BuildExecutor{Kind: "TaskRun", Name: executorName},
		},
	}

	taskRun := &unstructured.Unstructured{
		Object: map[string]interface{} {
			"apiVersion": "tekton.dev/v1",
			"kind": "TaskRun",
			"metadata": map[string]interface{}{"name": executorName, "namespace": namespace},
			"status": map[string]interface{}{"podName": podName},
		},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
			Namespace: namespace,
			Labels: map[string]string{"tekton.dev/taskRun": executorName},
		},
		Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "step-build"}}},
	}

	shpClient := shpfake.NewSimpleClientset(buildRun)
	kubeClient := fake.NewSimpleClientset(pod)
	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme.Scheme, taskRun)

	tmpDir, _ := os.MkdirTemp("", "gather-tr-*")
	defer os.RemoveAll(tmpDir)

	p := params.NewParamsForTest(kubeClient, shpClient, dynamicClient, nil, namespace, nil, nil)
	ioStreams, _, _, _ := genericclioptions.NewTestIOStreams()
	c := &cobra.Command{}
	c.SetContext(context.Background())
	cmd := &GatherCommand{cmd: c, name: name, outputDir: tmpDir}

	if err := cmd.Run(p, &ioStreams); err != nil {
		t.Fatalf("Gather.Run failed: %v", err)
	}

	expectedDir := filepath.Join(tmpDir, fmt.Sprintf("buildrun-%s-gather", name))
	checkFiles(t, expectedDir, []string{"buildrun.yaml", "taskrun.yaml", "pod.yaml", "logs/step-build.log"})
}

func TestGatherPipelineRun(t *testing.T) {
	name := "test-br-pipelinerun"
	namespace := "default"
	pipelineRunName := "test-pr"
	taskRunName := "test-pr-tr"
	podName := "test-pr-pod"

	buildRun := &buildv1beta1.BuildRun{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Status: buildv1beta1.BuildRunStatus{
			Executor: &buildv1beta1.BuildExecutor{Kind: "PipelineRun", Name: pipelineRunName},
		},
	}

	pipelineRun := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "tekton.dev/v1",
			"kind": "PipelineRun",
			"metadata": map[string]interface{}{"name": pipelineRunName, "namespace": namespace},
		},
	}

	taskRun := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "tekton.dev/v1",
			"kind": "TaskRun",
			"metadata": map[string]interface{}{
				"name": taskRunName,
				"namespace": namespace,
				"labels": map[string]interface{}{"tekton.dev/pipelineRun": pipelineRunName},
			},
			"status": map[string]interface{}{"podName": podName},
		},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
			Namespace: namespace,
			Labels: map[string]string{"tekton.dev/taskRun": taskRunName},
		},
		Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "step-build"}}},
	}

	shpClient := shpfake.NewSimpleClientset(buildRun)
	kubeClient := fake.NewSimpleClientset(pod)

	// Seed dynamic client with both PipelineRun and TaskRun
	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme.Scheme, pipelineRun, taskRun)

	tmpDir, _ := os.MkdirTemp("", "gather-pr-*")
	defer os.RemoveAll(tmpDir)

	p := params.NewParamsForTest(kubeClient, shpClient, dynamicClient, nil, namespace, nil, nil)
	ioStreams, _, _, _ := genericclioptions.NewTestIOStreams()
	
	c := &cobra.Command{}
	c.SetContext(context.Background())
	cmd := &GatherCommand{cmd: c, name: name, outputDir: tmpDir}


	if err := cmd.Run(p, &ioStreams); err != nil {
		t.Fatalf("Gather.Run failed: %v", err)
	}

	expectedDir := filepath.Join(tmpDir, fmt.Sprintf("buildrun-%s-gather", name))


	// PipelineRun gather creates a slightly different directory structure
	checkFiles(t, expectedDir, []string{
		"buildrun.yaml",
		"pipelinerun.yaml",
		filepath.Join("taskruns", taskRunName +".yaml"),
		filepath.Join("pods", podName+".yaml"),
		filepath.Join("logs", taskRunName, "step-build.log"),
	})
}

func TestGatherArchive(t *testing.T) {
	name := "test-br-archive"
	namespace := "default"
	executorName := "test-br"

	buildRun := &buildv1beta1.BuildRun{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Status: buildv1beta1.BuildRunStatus{
			Executor: &buildv1beta1.BuildExecutor{Kind: "TaskRun", Name: executorName},
		},
	}

	taskRun := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "tekton.dev/v1",
			"kind":       "TaskRun",
			"metadata":   map[string]interface{}{"name": executorName, "namespace": namespace},
			"status":     map[string]interface{}{"podName": "some-pod"},
		},
	}

	shpClient := shpfake.NewSimpleClientset(buildRun)
	kubeClient := fake.NewSimpleClientset() // Pod not strictly needed for archive logic test
	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme.Scheme, taskRun)

	tmpDir, _ := os.MkdirTemp("", "gather-archive-*")
	defer os.RemoveAll(tmpDir)

	p := params.NewParamsForTest(kubeClient, shpClient, dynamicClient, nil, namespace, nil, nil)
	ioStreams, _, _, _ := genericclioptions.NewTestIOStreams()

	c := &cobra.Command{}
	c.SetContext(context.Background())

	// Enable the archive flag
	cmd := &GatherCommand{
		cmd:       c,
		name:      name,
		outputDir: tmpDir,
		archive:   true, 
	}

	if err := cmd.Run(p, &ioStreams); err != nil {
		t.Fatalf("Gather.Run failed: %v", err)
	}

	expectedPrefix := filepath.Join(tmpDir, fmt.Sprintf("buildrun-%s-gather", name))
	archivePath := expectedPrefix + ".tar.gz"

	// Verify the archive file exists
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Errorf("expected archive file %q was not created", archivePath)
	}

	// Verify the temporary directory was cleaned up (deleted)
	if _, err := os.Stat(expectedPrefix); err == nil {
		t.Errorf("expected directory %q to be deleted after archiving, but it still exists", expectedPrefix)
	}
}

func checkFiles(t *testing.T, baseDir string, files []string) {
	for _, f := range files {
		path := filepath.Join(baseDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %q not found in gathered directory", path)
		}
	}
}
