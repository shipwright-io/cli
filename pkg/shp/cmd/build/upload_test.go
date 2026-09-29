package build // nolint:revive

import (
	"context"
	"strings"
	"testing"
	"time"

	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	shpfake "github.com/shipwright-io/build/pkg/client/clientset/versioned/fake"
	"github.com/shipwright-io/cli/pkg/shp/flags"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/reactor"
	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes/fake"
	fakekubetesting "k8s.io/client-go/testing"
)

// newUploadCommandForTest builds an UploadCommand wired with a fresh cobra command and the standard
// BuildRun spec flags, mimicking uploadCmd().
func newUploadCommandForTest() *UploadCommand {
	ccmd := &cobra.Command{}
	u := &UploadCommand{
		cmd:          ccmd,
		buildRunSpec: flags.BuildRunSpecFromFlags(ccmd.Flags()),
	}
	flags.FollowFlag(ccmd.Flags(), &u.follow)
	buildRunNameFlag(ccmd.Flags(), &u.buildRunName)
	return u
}

func TestUploadValidateBuildRunNameConflicts(t *testing.T) {
	t.Run("buildrun-name with a spec flag is rejected", func(t *testing.T) {
		u := newUploadCommandForTest()
		if err := u.cmd.Flags().Set(flags.BuildrunNameFlag, "existing-br"); err != nil {
			t.Fatal(err)
		}
		if err := u.cmd.Flags().Set(flags.ServiceAccountNameFlag, "builder"); err != nil {
			t.Fatal(err)
		}

		err := u.checkBuildRunNameConflicts()
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
		if !strings.Contains(err.Error(), flags.ServiceAccountNameFlag) {
			t.Errorf("error should mention the conflicting flag, got: %s", err.Error())
		}
	})

	t.Run("buildrun-name alone is accepted", func(t *testing.T) {
		u := newUploadCommandForTest()
		if err := u.cmd.Flags().Set(flags.BuildrunNameFlag, "existing-br"); err != nil {
			t.Fatal(err)
		}
		// buildref-name is set programmatically from the positional argument
		if err := u.cmd.Flags().Set(flags.BuildrefNameFlag, "my-build"); err != nil {
			t.Fatal(err)
		}

		if err := u.checkBuildRunNameConflicts(); err != nil {
			t.Errorf("expected no error, got: %s", err.Error())
		}
	})

	t.Run("no buildrun-name allows spec flags", func(t *testing.T) {
		u := newUploadCommandForTest()
		if err := u.cmd.Flags().Set(flags.ServiceAccountNameFlag, "builder"); err != nil {
			t.Fatal(err)
		}

		if err := u.checkBuildRunNameConflicts(); err != nil {
			t.Errorf("expected no error, got: %s", err.Error())
		}
	})
}

func TestUploadResolveBuildRunFetchesExisting(t *testing.T) {
	localSource := &buildv1beta1.BuildRunSource{
		Type:  buildv1beta1.LocalType,
		Local: &buildv1beta1.Local{Name: "local-copy"},
	}

	tests := []struct {
		name      string
		buildName string
		source    *buildv1beta1.BuildRunSource
		wantErr   string
	}{
		{name: "matching Build and Local source", buildName: "my-build", source: localSource},
		{name: "another Build is rejected", buildName: "other-build", source: localSource, wantErr: "does not reference Build 'my-build'"},
		{name: "a source that is not Local is rejected", buildName: "my-build", source: &buildv1beta1.BuildRunSource{}, wantErr: "must have a source of type"},
		{name: "no source is rejected", buildName: "my-build", wantErr: "must have a source of type"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			existing := &buildv1beta1.BuildRun{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: metav1.NamespaceDefault,
					Name:      "existing-br",
				},
				Spec: buildv1beta1.BuildRunSpec{
					Build:  buildv1beta1.ReferencedBuild{Name: &tt.buildName},
					Source: tt.source,
				},
			}
			shpclientset := shpfake.NewSimpleClientset(existing)
			kclientset := fake.NewSimpleClientset()
			param := params.NewParamsForTest(kclientset, shpclientset, nil, genericclioptions.NewConfigFlags(true), metav1.NamespaceDefault, nil, nil)

			ioStreams, _, _, _ := genericclioptions.NewTestIOStreams()

			u := newUploadCommandForTest()
			u.ioStreams = &ioStreams
			u.buildRefName = "my-build"
			u.buildRunName = "existing-br"

			br, err := u.resolveBuildRun(param)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected an error containing %q, got: %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %s", err.Error())
				}
				if br.Name != "existing-br" {
					t.Errorf("expected to fetch existing-br, got: %s", br.Name)
				}
			}

			// no BuildRun should have been created
			list, err := shpclientset.ShipwrightV1beta1().BuildRuns(metav1.NamespaceDefault).List(u.cmd.Context(), metav1.ListOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if len(list.Items) != 1 {
				t.Errorf("expected exactly 1 BuildRun (the pre-existing one), got: %d", len(list.Items))
			}
		})
	}
}

func TestUploadValidateBuildRunNameAllowsGlobalFlags(t *testing.T) {
	u := newUploadCommandForTest()
	// the global flags, such as --namespace, are inherited from the root command
	root := &cobra.Command{}
	genericclioptions.NewConfigFlags(true).AddFlags(root.PersistentFlags())
	root.AddCommand(u.cmd)

	if err := u.cmd.ParseFlags([]string{"--namespace=other", "--buildrun-name=existing-br"}); err != nil {
		t.Fatal(err)
	}

	if err := u.checkBuildRunNameConflicts(); err != nil {
		t.Errorf("expected no error, got: %s", err.Error())
	}
}

func TestUploadValidateBuildRunNameRejectsSourceBundle(t *testing.T) {
	u := newUploadCommandForTest()
	u.sourceDir = t.TempDir()
	u.sourceBundleImage = "registry.example.com/source-bundle:latest"
	if err := u.cmd.Flags().Set(flags.BuildrunNameFlag, "existing-br"); err != nil {
		t.Fatal(err)
	}

	err := u.Validate()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "source bundle") {
		t.Errorf("error should mention the source bundle, got: %s", err.Error())
	}
}

func TestUploadRunHandlesPodThatExistsBeforeTheWatch(t *testing.T) {
	buildName := "my-build"
	existing := &buildv1beta1.BuildRun{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: metav1.NamespaceDefault,
			Name:      "existing-br",
		},
		Spec: buildv1beta1.BuildRunSpec{
			Build: buildv1beta1.ReferencedBuild{Name: &buildName},
			Source: &buildv1beta1.BuildRunSource{
				Type:  buildv1beta1.LocalType,
				Local: &buildv1beta1.Local{Name: "local-copy"},
			},
		},
	}
	shpclientset := shpfake.NewSimpleClientset(existing)
	kclientset := fake.NewSimpleClientset()

	// the pod of an existing BuildRun can be past the point the upload waits for when the watch
	// starts, so the only event the command receives for it is the first one, of type Added
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: metav1.NamespaceDefault, Name: "existing-br-pod"},
		Status:     corev1.PodStatus{Phase: corev1.PodSucceeded},
	}
	watcher := watch.NewFakeWithChanSize(1, false)
	watcher.Add(pod)
	kclientset.PrependWatchReactor("pods", fakekubetesting.DefaultWatchReactor(watcher, nil))

	param := params.NewParamsForTest(kclientset, shpclientset, nil, genericclioptions.NewConfigFlags(true), metav1.NamespaceDefault, nil, nil)
	ioStreams, _, _, _ := genericclioptions.NewTestIOStreams()

	u := newUploadCommandForTest()
	u.cmd.SetContext(context.Background())
	u.ioStreams = &ioStreams
	u.buildRefName = buildName
	u.buildRunName = "existing-br"
	if err := u.cmd.Flags().Set(flags.BuildrefNameFlag, buildName); err != nil {
		t.Fatal(err)
	}

	pw, err := reactor.NewPodWatcher(context.Background(), time.Minute, kclientset, metav1.NamespaceDefault)
	if err != nil {
		t.Fatal(err)
	}
	u.pw = pw

	done := make(chan error, 1)
	go func() { done <- u.Run(param, &ioStreams) }()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
	case <-time.After(10 * time.Second):
		pw.Stop()
		t.Fatal("upload did not react to the pod's first event and kept waiting")
	}
}
