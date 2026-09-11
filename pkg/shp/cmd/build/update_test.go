package build // nolint:revive

import (
	"strings"
	"testing"
	"time"

	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	shpfake "github.com/shipwright-io/build/pkg/client/clientset/versioned/fake"
	"github.com/shipwright-io/cli/pkg/shp/flags"
	"github.com/shipwright-io/cli/pkg/shp/params"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	kclientsetfake "k8s.io/client-go/kubernetes/fake"
)

func TestBuildUpdate_Success(t *testing.T) {
	oldRevision := "v1.0"
	testBuild := &buildv1beta1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-build",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: buildv1beta1.BuildSpec{
			Source: &buildv1beta1.Source{
				Git: &buildv1beta1.Git{
					URL:      "https://github.com/shipwright-io/old-repo",
					Revision: &oldRevision,
				},
			},
			Output: buildv1beta1.Image{
				Image: "quay.io/myuser/old-app:v1.0",
			},
		},
	}

	shpClientset := shpfake.NewSimpleClientset(testBuild)
	k8sClientset := kclientsetfake.NewSimpleClientset()

	cmd := updateCmd()
	configFlags := genericclioptions.NewConfigFlags(true)
	timeout := 10 * time.Second
	p := params.NewParamsForTest(k8sClientset, shpClientset, nil, configFlags, metav1.NamespaceDefault, &timeout, &timeout)

	ioStreams, _, out, _ := genericclioptions.NewTestIOStreams()

	if err := cmd.Complete(p, &ioStreams, []string{"my-build"}); err != nil {
		t.Fatalf("unexpected error in Complete: %v", err)
	}

	// Set updated flag values
	if err := cmd.Cmd().Flags().Set(flags.SourceGitURLFlag, "https://github.com/shipwright-io/new-repo"); err != nil {
		t.Fatalf("failed to set source-git-url flag: %v", err)
	}
	if err := cmd.Cmd().Flags().Set(flags.SourceGitRevisionFlag, "v2.0"); err != nil {
		t.Fatalf("failed to set source-git-revision flag: %v", err)
	}
	if err := cmd.Cmd().Flags().Set(flags.OutputImageFlag, "quay.io/myuser/new-app:v2.0"); err != nil {
		t.Fatalf("failed to set output-image flag: %v", err)
	}

	if err := cmd.Run(p, &ioStreams); err != nil {
		t.Fatalf("unexpected error in Run: %v", err)
	}

	output := out.String()
	expectedMsg := `Updated build "my-build"`
	if !strings.Contains(output, expectedMsg) {
		t.Errorf("expected output to contain %q, but got:\n%s", expectedMsg, output)
	}

	// Verify object fields were updated in client
	updatedBuild, err := shpClientset.ShipwrightV1beta1().Builds(metav1.NamespaceDefault).Get(cmd.Cmd().Context(), "my-build", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get updated build: %v", err)
	}

	if updatedBuild.Spec.Source.Git.URL != "https://github.com/shipwright-io/new-repo" {
		t.Errorf("expected updated URL 'https://github.com/shipwright-io/new-repo', got %q", updatedBuild.Spec.Source.Git.URL)
	}

	if *updatedBuild.Spec.Source.Git.Revision != "v2.0" {
		t.Errorf("expected updated revision 'v2.0', got %q", *updatedBuild.Spec.Source.Git.Revision)
	}

	if updatedBuild.Spec.Output.Image != "quay.io/myuser/new-app:v2.0" {
		t.Errorf("expected updated output image 'quay.io/myuser/new-app:v2.0', got %q", updatedBuild.Spec.Output.Image)
	}
}

func TestBuildUpdate_NotFound(t *testing.T) {
	shpClientset := shpfake.NewSimpleClientset()
	k8sClientset := kclientsetfake.NewSimpleClientset()

	cmd := updateCmd()
	configFlags := genericclioptions.NewConfigFlags(true)
	timeout := 10 * time.Second
	p := params.NewParamsForTest(k8sClientset, shpClientset, nil, configFlags, metav1.NamespaceDefault, &timeout, &timeout)

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
