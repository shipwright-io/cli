package buildrun

import (
	"strings"
	"testing"
	"time"

	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	shpfake "github.com/shipwright-io/build/pkg/client/clientset/versioned/fake"
	"github.com/shipwright-io/cli/pkg/shp/params"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	kclientsetfake "k8s.io/client-go/kubernetes/fake"
)

func TestBuildRunGet_DefaultTable(t *testing.T) {
	now := metav1.Now()
	testBuildRun := &buildv1beta1.BuildRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-buildrun",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: buildv1beta1.BuildRunSpec{
			Build: buildv1beta1.ReferencedBuild{
				Name: ptr("my-build"),
			},
		},
		Status: buildv1beta1.BuildRunStatus{
			StartTime:      &now,
			CompletionTime: &now,
			Conditions: []buildv1beta1.Condition{
				{
					Type:   buildv1beta1.Succeeded,
					Reason: "Succeeded",
				},
			},
		},
	}

	shpClientset := shpfake.NewSimpleClientset(testBuildRun)
	k8sClientset := kclientsetfake.NewSimpleClientset()

	cmd := getCmd()
	flags := genericclioptions.NewConfigFlags(true)
	timeout := 10 * time.Second
	p := params.NewParamsForTest(k8sClientset, shpClientset, nil, flags, metav1.NamespaceDefault, &timeout, &timeout)

	ioStreams, _, out, _ := genericclioptions.NewTestIOStreams()

	if err := cmd.Complete(p, &ioStreams, []string{"my-buildrun"}); err != nil {
		t.Fatalf("unexpected error in Complete: %v", err)
	}

	if err := cmd.Validate(); err != nil {
		t.Fatalf("unexpected error in Validate: %v", err)
	}

	if err := cmd.Run(p, &ioStreams); err != nil {
		t.Fatalf("unexpected error in Run: %v", err)
	}

	output := out.String()
	expectedStrings := []string{
		"NAME:",
		"my-buildrun",
		"NAMESPACE:",
		"default",
		"BUILD NAME:",
		"my-build",
		"STATUS:",
		"Succeeded",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, but got:\n%s", expected, output)
		}
	}
}

func TestBuildRunGet_JSON(t *testing.T) {
	testBuildRun := &buildv1beta1.BuildRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-buildrun",
			Namespace: metav1.NamespaceDefault,
		},
	}

	shpClientset := shpfake.NewSimpleClientset(testBuildRun)
	k8sClientset := kclientsetfake.NewSimpleClientset()

	cmd := getCmd()
	flags := genericclioptions.NewConfigFlags(true)
	timeout := 10 * time.Second
	p := params.NewParamsForTest(k8sClientset, shpClientset, nil, flags, metav1.NamespaceDefault, &timeout, &timeout)

	ioStreams, _, out, _ := genericclioptions.NewTestIOStreams()

	if err := cmd.Complete(p, &ioStreams, []string{"my-buildrun"}); err != nil {
		t.Fatalf("unexpected error in Complete: %v", err)
	}

	if err := cmd.Cmd().Flags().Set("output", "json"); err != nil {
		t.Fatalf("failed to set output flag: %v", err)
	}

	if err := cmd.Validate(); err != nil {
		t.Fatalf("unexpected error in Validate: %v", err)
	}

	if err := cmd.Run(p, &ioStreams); err != nil {
		t.Fatalf("unexpected error in Run: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, `"name": "my-buildrun"`) {
		t.Errorf("expected JSON output, but got:\n%s", output)
	}
}

func TestBuildRunGet_YAML(t *testing.T) {
	testBuildRun := &buildv1beta1.BuildRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-buildrun",
			Namespace: metav1.NamespaceDefault,
		},
	}

	shpClientset := shpfake.NewSimpleClientset(testBuildRun)
	k8sClientset := kclientsetfake.NewSimpleClientset()

	cmd := getCmd()
	flags := genericclioptions.NewConfigFlags(true)
	timeout := 10 * time.Second
	p := params.NewParamsForTest(k8sClientset, shpClientset, nil, flags, metav1.NamespaceDefault, &timeout, &timeout)

	ioStreams, _, out, _ := genericclioptions.NewTestIOStreams()

	if err := cmd.Complete(p, &ioStreams, []string{"my-buildrun"}); err != nil {
		t.Fatalf("unexpected error in Complete: %v", err)
	}

	if err := cmd.Cmd().Flags().Set("output", "yaml"); err != nil {
		t.Fatalf("failed to set output flag: %v", err)
	}

	if err := cmd.Validate(); err != nil {
		t.Fatalf("unexpected error in Validate: %v", err)
	}

	if err := cmd.Run(p, &ioStreams); err != nil {
		t.Fatalf("unexpected error in Run: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "name: my-buildrun") {
		t.Errorf("expected YAML output, but got:\n%s", output)
	}
}

func TestBuildRunGet_NotFound(t *testing.T) {
	shpClientset := shpfake.NewSimpleClientset()
	k8sClientset := kclientsetfake.NewSimpleClientset()

	cmd := getCmd()
	flags := genericclioptions.NewConfigFlags(true)
	timeout := 10 * time.Second
	p := params.NewParamsForTest(k8sClientset, shpClientset, nil, flags, metav1.NamespaceDefault, &timeout, &timeout)

	ioStreams, _, _, _ := genericclioptions.NewTestIOStreams()

	if err := cmd.Complete(p, &ioStreams, []string{"nonexistent-buildrun"}); err != nil {
		t.Fatalf("unexpected error in Complete: %v", err)
	}

	err := cmd.Run(p, &ioStreams)
	if err == nil {
		t.Fatalf("expected error for nonexistent buildrun, but got nil")
	}

	if !strings.Contains(err.Error(), "nonexistent-buildrun") {
		t.Errorf("expected error message to contain buildrun name, got %v", err)
	}
}

func TestBuildRunGet_InvalidOutput(t *testing.T) {
	testBuildRun := &buildv1beta1.BuildRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-buildrun",
			Namespace: metav1.NamespaceDefault,
		},
	}
	shpClientset := shpfake.NewSimpleClientset(testBuildRun)
	k8sClientset := kclientsetfake.NewSimpleClientset()

	cmd := getCmd()
	flags := genericclioptions.NewConfigFlags(true)
	timeout := 10 * time.Second
	p := params.NewParamsForTest(k8sClientset, shpClientset, nil, flags, metav1.NamespaceDefault, &timeout, &timeout)

	ioStreams, _, _, _ := genericclioptions.NewTestIOStreams()

	if err := cmd.Complete(p, &ioStreams, []string{"my-buildrun"}); err != nil {
		t.Fatalf("unexpected error in Complete: %v", err)
	}

	if err := cmd.Cmd().Flags().Set("output", "invalid"); err != nil {
		t.Fatalf("failed to set output flag: %v", err)
	}

	err := cmd.Validate()
	if err == nil {
		t.Fatalf("expected error for unsupported output format, but got nil")
	}

	expectedErr := `unsupported output format "invalid". Supported formats are: json, yaml`
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, but got %q", expectedErr, err.Error())
	}
}

func ptr(s string) *string {
	return &s
}
