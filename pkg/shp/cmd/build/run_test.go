package build

import (
	"runtime"
	"strings"
	"sync"
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
	"k8s.io/apimachinery/pkg/util/wait"
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
		cancelled  bool
		brDeleted  bool
		podDeleted bool
	}{
		{
			name:    "succeeded",
			phase:   corev1.PodSucceeded,
			logText: "Pod 'testpod' has succeeded!",
		},
		{
			name:    "pending",
			phase:   corev1.PodPending,
			logText: "Pod 'testpod' is in state \"Pending\"",
		},
		{
			name:    "unknown",
			phase:   corev1.PodUnknown,
			logText: "Pod 'testpod' is in state \"Unknown\"",
		},
		{
			name:      "failed-cancelled",
			phase:     corev1.PodFailed,
			cancelled: true,
			logText:   "BuildRun 'testpod' has been canceled.",
		},
		{
			name:      "failed-br-deleted",
			phase:     corev1.PodFailed,
			brDeleted: true,
			logText:   "BuildRun 'testpod' has been deleted.",
		},
		{
			name:       "failed-pod-deleted",
			phase:      corev1.PodFailed,
			podDeleted: true,
			logText:    "Pod 'testpod' has been deleted.",
		},
		{
			name:    "failed-something-else",
			phase:   corev1.PodFailed,
			logText: "Pod 'testpod' has failed!",
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
			to:      "1s",
			logText: reactor.RequestTimeoutMessage,
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
		kclientset := fake.NewSimpleClientset(pod)
		ccmd := &cobra.Command{}
		cmd := &RunCommand{
			cmd:             ccmd,
			buildRunName:    name,
			buildRunSpec:    flags.BuildRunSpecFromFlags(ccmd.Flags()),
			follow:          true,
			shpClientset:    shpclientset,
			tailLogsStarted: make(map[string]bool),
			watchLock:       sync.Mutex{},
		}

		// set up context
		cmd.Cmd().ExecuteC()
		pm := genericclioptions.NewConfigFlags(true)
		if len(test.to) > 0 {
			pm.Timeout = &test.to
		}
		param := params.NewParamsForTest(kclientset, shpclientset, pm, metav1.NamespaceDefault)

		ioStreams, _, out, _ := genericclioptions.NewTestIOStreams()

		switch {
		case test.cancelled:
			br.Spec.State = buildv1alpha1.BuildRunStateCancel
		case test.brDeleted:
			br.DeletionTimestamp = &metav1.Time{}
		case test.podDeleted:
			pod.DeletionTimestamp = &metav1.Time{}
		}

		cmd.Complete(param, []string{name})
		if len(test.to) > 0 {
			cmd.Run(param, &ioStreams)
			if !strings.Contains(out.String(), test.logText) {
				t.Errorf("test %s: unexpected output: %s", test.name, out.String())
			}
			continue
		}
		go func() {
			err := cmd.Run(param, &ioStreams)
			if err != nil {
				t.Errorf("%s", err.Error())
			}

		}()

		// yield the processor, so the initialization in Run can occur; afterward, the watchLock should allow
		// coordination between Run and onEvent
		runtime.Gosched()

		// even with our release of the context above with Gosched(), repeated runs in CI have surfaced occasional timing issues between
		// cmd.Run() finishing initialization and cmd.onEvent trying to used struct variables, resulting in panics; so we employ the lock here
		// to insure the required initializations have run; this is still better than a generic "sleep log enough for
		// the init to occur.
		cmd.watchLock.Lock()
		err := wait.PollImmediate(1*time.Second, 10*time.Second, func() (done bool, err error) {
			// check any of the vars on RunCommand that are used in onEvent and make sure they are set;
			// we are verifying the initialization done in Run() on RunCommand is complete
			if cmd.pw != nil && cmd.ioStreams != nil && cmd.shpClientset != nil {
				cmd.watchLock.Unlock()
				return true, nil
			}
			return false, nil
		})
		if err != nil {
			cmd.watchLock.Unlock()
			t.Errorf("test %s: Run initialization did not complete in time: pw %#v ioStreams %#v shpClientset %#v", test.name, cmd.pw, cmd.ioStreams, cmd.shpClientset)
		}

		// mimic watch events, bypassing k8s fake client watch hoopla whose plug points are not always useful;
		pod.Status.Phase = test.phase
		cmd.onEvent(pod)
		if !strings.Contains(out.String(), test.logText) {
			t.Errorf("test %s: unexpected output: %s", test.name, out.String())
		}

	}
}
