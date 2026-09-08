package build // nolint:revive

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

func TestBuildGet_DefaultTable(t *testing.T) {
	revision := "main"
	strategyKind := buildv1beta1.ClusterBuildStrategyKind
	testBuild := &buildv1beta1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-build",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: buildv1beta1.BuildSpec{
			Source: &buildv1beta1.Source{
				Git: &buildv1beta1.Git{
					URL:      "https://github.com/shipwright-io/sample-go",
					Revision: &revision,
				},
			},
			Strategy: buildv1beta1.Strategy{
				Name: "buildpacks-v3",
				Kind: &strategyKind,
			},
			Output: buildv1beta1.Image{
				Image: "quay.io/myuser/my-app:latest",
			},
		},
	}

	shpClientset := shpfake.NewSimpleClientset(testBuild);
	k8sClientset := kclientsetfake.NewSimpleClientset();

	cmd := getCmd()
	flags := genericclioptions.NewConfigFlags(true);
	timeout := 10 * time.Second;
	p := params.NewParamsForTest(k8sClientset, shpClientset, nil, flags, metav1.NamespaceDefault, &timeout, &timeout);
	
	ioStreams, _, out, _ := genericclioptions.NewTestIOStreams()

	if err := cmd.Complete(p, &ioStreams, []string{"my-build"}); err != nil {
		t.Fatalf("unexpected error in Complete: %v", err)
	}

	if err := cmd.Run(p, &ioStreams); err != nil {
		t.Fatalf("unexpected error in Run: %v", err)
	}

	output := out.String();
	expectedStrings := []string{
		"NAME:",
		"my-build",
		"NAMESPACE:",
		"default",
		"SOURCE URL:",
		"https://github.com/shipwright-io/sample-go",
		"REVISION:",
		"main",
		"STRATEGY:",
		"buildpacks-v3 (ClusterBuildStrategy)",
		"OUTPUT IMAGE:",
		"quay.io/myuser/my-app:latest",
	}

	for _, expected  := range expectedStrings {
		if !strings.Contains(output, expected){
			t.Errorf("expected output to contain %q, but got:\n%s", expected, output)
		}
	}
}

func TestBuildGet_JSON(t *testing.T) {
	testBuild := &buildv1beta1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-build",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: buildv1beta1.BuildSpec{
			Output: buildv1beta1.Image{
				Image: "quay.io/myuser/my-app:latest",
			},
		},
	}

	shpClientset := shpfake.NewSimpleClientset(testBuild)
	k8sClientset := kclientsetfake.NewSimpleClientset()

	cmd := getCmd()
	flags := genericclioptions.NewConfigFlags(true)
	timeout := 10 * time.Second
	p := params.NewParamsForTest(k8sClientset, shpClientset, nil, flags, metav1.NamespaceDefault, &timeout, &timeout)

	ioStreams, _, out, _ := genericclioptions.NewTestIOStreams()

	if err := cmd.Complete(p, &ioStreams, []string{"my-build"}); err != nil {
		t.Fatalf("unexpected error in Complete: %v", err)
	}

	// Set output flag to "json"
	if err := cmd.Cmd().Flags().Set("output", "json"); err != nil {
		t.Fatalf("failed to set output flag: %v", err)
	}

	if err := cmd.Run(p, &ioStreams); err != nil {
		t.Fatalf("unexpected error in Run: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, `"name": "my-build"`) || !strings.Contains(output, `"image": "quay.io/myuser/my-app:latest"`) {
		t.Errorf("expected JSON output containing build details, but got:\n%s", output)
	}
}

func TestBuildGet_YAML(t *testing.T) {
	testBuild := &buildv1beta1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-build",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: buildv1beta1.BuildSpec{
			Output: buildv1beta1.Image{
				Image: "quay.io/myuser/my-app:latest",
			},
		},
	}

	shpClientset := shpfake.NewSimpleClientset(testBuild)
	k8sClientset := kclientsetfake.NewSimpleClientset()

	cmd := getCmd()
	flags := genericclioptions.NewConfigFlags(true)
	timeout := 10 * time.Second
	p := params.NewParamsForTest(k8sClientset, shpClientset, nil, flags, metav1.NamespaceDefault, &timeout, &timeout)

	ioStreams, _, out, _ := genericclioptions.NewTestIOStreams()

	if err := cmd.Complete(p, &ioStreams, []string{"my-build"}); err != nil {
		t.Fatalf("unexpected error in Complete: %v", err)
	}

	// Set output flag to "yaml"
	if err := cmd.Cmd().Flags().Set("output", "yaml"); err != nil {
		t.Fatalf("failed to set output flag: %v", err)
	}

	if err := cmd.Run(p, &ioStreams); err != nil {
		t.Fatalf("unexpected error in Run: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "name: my-build") || !strings.Contains(output, "image: quay.io/myuser/my-app:latest") {
		t.Errorf("expected YAML output containing build details, but got:\n%s", output)
	}
}

func TestBuildGet_NotFound(t *testing.T) {
	shpClientset := shpfake.NewSimpleClientset();
	k8sClientSet := kclientsetfake.NewSimpleClientset();

	cmd := getCmd()
	flags := genericclioptions.NewConfigFlags(true)
	timeout := 10 * time.Second;
	p := params.NewParamsForTest(k8sClientSet, shpClientset, nil, flags, metav1.NamespaceDefault, &timeout, &timeout);
	
	ioStreams, _, out, _ := genericclioptions.NewTestIOStreams()

	if err := cmd.Complete(p, &ioStreams, []string{"nonexistent-build"}); err != nil {
		t.Fatalf("unexpected error in Complete: %v", err)
	}
	if err := cmd.Run(p, &ioStreams); err != nil {
		t.Fatalf("unexpected error in Run: %v", err)
	}
	output := out.String()
	expectedMsg := "Build 'nonexistent-build' not found in namespace 'default'."
	if !strings.Contains(output, expectedMsg) {
		t.Errorf("expected output to contain %q, but got:\n%s", expectedMsg, output)
	}
}

func TestBuildGet_InvalidOutput(t *testing.T) {
	testBuild := &buildv1beta1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-build",
			Namespace: metav1.NamespaceDefault,
		},
	}
	shpClientset := shpfake.NewSimpleClientset(testBuild)
	k8sClientset := kclientsetfake.NewSimpleClientset()
	cmd := getCmd()
	flags := genericclioptions.NewConfigFlags(true)
	timeout := 10 * time.Second
	p := params.NewParamsForTest(k8sClientset, shpClientset, nil, flags, metav1.NamespaceDefault, &timeout, &timeout)
	ioStreams, _, _, _ := genericclioptions.NewTestIOStreams()
	if err := cmd.Complete(p, &ioStreams, []string{"my-build"}); err != nil {
		t.Fatalf("unexpected error in Complete: %v", err)
	}
	// Set invalid output flag
	if err := cmd.Cmd().Flags().Set("output", "invalid"); err != nil {
		t.Fatalf("failed to set output flag: %v", err)
	}
	err := cmd.Run(p, &ioStreams)
	if err == nil {
		t.Fatalf("expected error for unsupported output format, but got nil")
	}
	expectedErr := `unsupported output format "invalid". Supported formats are: json, yaml`
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, but got %q", expectedErr, err.Error())
	}
}