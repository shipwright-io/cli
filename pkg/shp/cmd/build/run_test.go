package build

import (
	"bytes"
	"strings"
	"testing"
	"time"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	shpfake "github.com/shipwright-io/build/pkg/client/clientset/versioned/fake"
	"github.com/shipwright-io/cli/pkg/shp/flags"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/reactor"
	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes/fake"
	fakekubetesting "k8s.io/client-go/testing"
)

func TestStartBuildRunFollowLog(t *testing.T) {
	tests := []struct {
		name       string
		phase      corev1.PodPhase
		logText    string
		to         string
		noPodYet   bool
		cancelled  bool
		brDeleted  bool
		podDeleted bool
	}{
		{
			name:    "succeeded",
			phase:   corev1.PodSucceeded,
			logText: "Pod \"testpod\" has succeeded!",
		},
		{
			name:    "pending",
			phase:   corev1.PodPending,
			logText: "Pod \"testpod\" is in state \"Pending\"",
		},
		{
			name:    "unknown",
			phase:   corev1.PodUnknown,
			logText: "Pod \"testpod\" is in state \"Unknown\"",
		},
		{
			name:      "failed-cancelled",
			phase:     corev1.PodFailed,
			cancelled: true,
			logText:   "BuildRun \"testpod\" has been canceled.",
		},
		{
			name:      "failed-br-deleted",
			phase:     corev1.PodFailed,
			brDeleted: true,
			logText:   "BuildRun \"testpod\" has been deleted.",
		},
		{
			name:       "failed-pod-deleted",
			phase:      corev1.PodFailed,
			podDeleted: true,
			logText:    "Pod \"testpod\" has been deleted.",
		},
		{
			name:    "failed-something-else",
			phase:   corev1.PodFailed,
			logText: "BuildRun \"testpod\" has failed.",
		},
		{
			name:  "running",
			phase: corev1.PodRunning,
			// we do not verify log text here; the k8s fake client stuff around watches, getting pods logs, and
			// what we do in this repo's tail logic is unreliable, and we've received guidance from some upstream
			// k8s folks to "be careful" with it; fortunately, what we do for tail and pod_watcher so far is within
			// the realm of reliable.
		},
		{
			name:    "timeout",
			to:      "1ms",
			logText: reactor.RequestTimeoutMessage,
		},
		{
			name:     "no pod yet",
			noPodYet: true,
			logText:  "has not observed any pod events yet",
		},
	}

	for _, test := range tests {
		name := "testpod"
		containerName := "container"
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: metav1.NamespaceDefault,
				Name:      name,
				Labels: map[string]string{
					buildv1alpha1.LabelBuild:    name,
					buildv1alpha1.LabelBuildRun: name,
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name: containerName,
				}},
			},
		}
		br := &buildv1alpha1.BuildRun{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: metav1.NamespaceDefault,
				Name:      name,
			},
		}
		shpclientset := shpfake.NewSimpleClientset()

		// need this reactor since the Run method uses the ObjectMeta.GenerateName k8s feature to generate the random
		// name for the BuildRun.  However, for our purposes with unit testing, we want to control the name of the BuildRun
		// to facilitate the list/selector via labels that is also employed by the Run method.
		createReactorFunc := func(action fakekubetesting.Action) (handled bool, ret kruntime.Object, err error) {
			return true, br, nil
		}
		shpclientset.PrependReactor("create", "buildruns", createReactorFunc)
		// need this reactor to handle returning our test BuildRun with whatever updates we make based on the various
		// test bools that result in spec.state or deletion timestamp updates
		getReactorFunc := func(action fakekubetesting.Action) (handled bool, ret kruntime.Object, err error) {
			return true, br, nil
		}
		shpclientset.PrependReactor("get", "buildruns", getReactorFunc)
		kclientset := fake.NewSimpleClientset()
		if !test.noPodYet {
			kclientset = fake.NewSimpleClientset(pod)
		}
		ccmd := &cobra.Command{}
		cmd := &RunCommand{
			cmd:          ccmd,
			buildRunSpec: flags.BuildRunSpecFromFlags(ccmd.Flags()),
			follow:       true,
		}

		// set up context
		cmd.Cmd().ExecuteC()
		pm := genericclioptions.NewConfigFlags(true)
		if len(test.to) > 0 {
			*pm.Timeout = test.to
		}
		failureDuration := 1 * time.Millisecond
		param := params.NewParamsForTest(kclientset, shpclientset, pm, metav1.NamespaceDefault, &failureDuration, &failureDuration)

		ioStreams, _, out, _ := genericclioptions.NewTestIOStreams()

		switch {
		case test.cancelled:
			br.Spec.State = buildv1alpha1.BuildRunRequestedStatePtr(buildv1alpha1.BuildRunStateCancel)
			br.Status.Conditions = []buildv1alpha1.Condition{
				{
					Type:   buildv1alpha1.Succeeded,
					Status: corev1.ConditionFalse,
				},
			}
		case test.brDeleted:
			br.DeletionTimestamp = &metav1.Time{}
			br.Status.Conditions = []buildv1alpha1.Condition{
				{
					Type:   buildv1alpha1.Succeeded,
					Status: corev1.ConditionFalse,
				},
			}
		case test.podDeleted:
			pod.DeletionTimestamp = &metav1.Time{}
			br.Status.Conditions = []buildv1alpha1.Condition{
				{
					Type:   buildv1alpha1.Succeeded,
					Status: corev1.ConditionFalse,
				},
			}
		case test.phase == corev1.PodRunning:
			pod.Status.ContainerStatuses = append(pod.Status.ContainerStatuses, corev1.ContainerStatus{
				State: corev1.ContainerState{
					Running: &corev1.ContainerStateRunning{StartedAt: metav1.Now()},
				},
			})
		}

		cmd.Complete(param, &ioStreams, []string{name})
		if len(test.to) > 0 {
			cmd.Run(param, &ioStreams)
			checkLog(test.name, test.logText, cmd, out, t)
			continue
		}

		go func() {
			err := cmd.Run(param, &ioStreams)
			if err != nil {
				t.Errorf("%s", err.Error())
			}
		}()

		// when employing the Run() method in a multi-threaded capacity, we must make sure
		// the underlying Follower/PodWatcher watches are sync'ed and ready for use before
		// we start populating the event queue
		ready := cmd.FollowerReady()
		if !ready {
			t.Errorf("%s follower no ready", test.name)
		}

		if !test.noPodYet {
			// mimic watch events, bypassing k8s fake client watch hoopla whose plug points are not always useful;
			pod.Status.Phase = test.phase
			cmd.follower.OnEvent(pod)
		} else {
			cmd.follower.OnNoPodEventsYet(nil)
		}
		checkLog(test.name, test.logText, cmd, out, t)
	}
}

func checkLog(name, text string, cmd *RunCommand, out *bytes.Buffer, t *testing.T) {
	// need to employ log lock since accessing same iostream out used by Run cmd
	cmd.follower.GetLogLock().Lock()
	defer cmd.follower.GetLogLock().Unlock()
	if !strings.Contains(out.String(), text) {
		t.Errorf("test %s: unexpected output: %s", name, out.String())
	}
}
