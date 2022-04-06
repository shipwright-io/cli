package follower

import (
	"context"
	"testing"
	"time"

	"github.com/onsi/gomega"
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/reactor"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type podTestCase string

const (
	// pod entered in the running state
	podRunning podTestCase = "running"

	// pod failed without a specific reason
	podFailed podTestCase = "failed"

	// pod failed because it has been deleted
	podDeleted podTestCase = "deleted"

	// pod status shows it's on unknown condition
	podStatusUnknown podTestCase = "statusUnknown"

	// pod is on succeeded status, but has never set as running
	podSucceededBeforeRunning podTestCase = "succeededBeforeRunning"

	// pod is on succeeded status
	podSucceeded podTestCase = "succeeded"
)

func preparePodForTestCase(testCase podTestCase) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: metav1.NamespaceDefault,
			Name:      "pod",
			Labels: map[string]string{
				buildv1alpha1.LabelBuild:    "build",
				buildv1alpha1.LabelBuildRun: "build",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name: "container",
			}},
		},
	}

	switch testCase {
	case podRunning:
		pod.Status.Phase = corev1.PodRunning
	case podFailed:
		pod.Status.Phase = corev1.PodFailed
	case podDeleted:
		now := metav1.Now()
		pod.DeletionTimestamp = &now
		pod.Status.Phase = corev1.PodFailed
	case podStatusUnknown:
		pod.Status.Phase = corev1.PodUnknown
		pod.Status.Conditions = []corev1.PodCondition{{
			Type:   corev1.ContainersReady,
			Status: corev1.ConditionUnknown,
		}}
	case podSucceeded, podSucceededBeforeRunning:
		pod.Status.Phase = corev1.PodSucceeded
	}

	return pod
}

func TestNewPodFollower(t *testing.T) {
	g := gomega.NewWithT(t)

	tests := []struct {
		testCase podTestCase
		logText  string
		wantErr  error
	}{{
		testCase: podRunning,
		logText:  "Pod \"pod\" in \"Running\" state, starting up log tail",
		wantErr:  nil,
	}, {
		testCase: podFailed,
		logText:  "Pod \"pod\" has failed!",
		wantErr:  ErrPodFailed,
	}, {
		testCase: podDeleted,
		logText:  "Pod \"pod\" has been deleted!",
		wantErr:  ErrPodDeleted,
	}, {
		testCase: podStatusUnknown,
		logText:  "Pod \"pod\" is in state \"Unknown\"...",
		wantErr:  ErrPodStatusUnknown,
	}, {
		testCase: podSucceededBeforeRunning,
		logText:  "*** Pod \"pod\", container \"container\": ***\n\nfake logs\n",
		wantErr:  nil,
	}, {
		testCase: podSucceeded,
		logText:  "Pod \"pod\" has succeeded!",
		wantErr:  nil,
	}}

	for _, tt := range tests {
		t.Run(string(tt.testCase), func(t *testing.T) {
			// preparing the pod for the test case, add and remove attributes to represent the test
			// case expected pod definition
			pod := preparePodForTestCase(tt.testCase)

			// setting up the requirements for the testing, putting together a PodWatcher instance
			// with the pod modified for the current test-case
			ctx := context.Background()

			fakeParams := params.NewFakeParams(20*time.Second, []runtime.Object{pod}, nil)
			pw, err := reactor.NewPodWatcherFromParams(ctx, fakeParams)
			g.Expect(err).To(gomega.BeNil())

			ioStreams, _, out, _ := genericclioptions.NewTestIOStreams()
			podFollower, err := NewPodLogsFollowerFromParams(ctx, fakeParams, pw, &ioStreams)
			g.Expect(err).To(gomega.BeNil())

			// setting up the PodFollower instance as previously running, in contrast with test-case
			// podSucceededBeforeRunning must have this flag false
			if tt.testCase == podSucceeded {
				podFollower.podIsRunning = true
			}

			// dispatching events directly on the PodWatcher instance, distributing events between
			// pod modified and deleted
			if tt.testCase == podDeleted {
				err = podFollower.pw.HandleEvent(pod, watch.Event{Type: watch.Deleted})
			} else {
				err = podFollower.pw.HandleEvent(pod, watch.Event{Type: watch.Modified})
			}

			output := out.String()
			t.Logf("output=%q", output)

			if tt.wantErr != nil {
				g.Expect(err).To(gomega.MatchError(tt.wantErr))
			}
			if tt.testCase == podRunning {
				g.Expect(podFollower.podIsRunning).To(gomega.BeTrue())
			}
			g.Expect(output).To(gomega.ContainSubstring(tt.logText))
		})
	}
}
