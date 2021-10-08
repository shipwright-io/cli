package build

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/tidwall/gjson"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd/types"
	testflags "github.com/shipwright-io/cli/test/flags"
)

// TODO: Fix broken test
/*func TestStartBuildRunFollowLog(t *testing.T) {
	tests := []struct {
		name       string
		phase      corev1.PodPhase
		logText    string
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
		param := params.NewParamsForTest(kclientset, shpclientset, nil, metav1.NamespaceDefault)
		o := &BuildRunOptions{}
		ioStreams, _, out, _ := genericclioptions.NewTestIOStreams()
		cmd := newBuildRunCmd(context.Background(), &ioStreams, param, o)
		cmd.ExecuteC()
		switch {
		case test.cancelled:
			br.Spec.State = buildv1alpha1.BuildRunStateCancel
		case test.brDeleted:
			br.DeletionTimestamp = &metav1.Time{}
		case test.podDeleted:
			pod.DeletionTimestamp = &metav1.Time{}
		}

		o.Complete([]string{name})
		go func() {
			err := o.Run()
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
		o.WatchLock.Lock()
		err := wait.PollImmediate(1*time.Second, 3*time.Second, func() (done bool, err error) {
			// check any of the vars on RunCommand that are used in onEvent and make sure they are set;
			// we are verifying the initialization done in Run() on RunCommand is complete
			if o.PodWatcher != nil && o.Streams != nil && o.Clients.ShipwrightClientSet != nil {
				o.WatchLock.Unlock()
				return true, nil
			}
			return false, nil
		})
		if err != nil {
			o.WatchLock.Unlock()
			t.Errorf("Run initialization did not complete in time")
		}

		// mimic watch events, bypassing k8s fake client watch hoopla whose plug points are not always useful;
		pod.Status.Phase = test.phase
		o.onEvent(pod)
		if !strings.Contains(out.String(), test.logText) {
			t.Errorf("test %s: unexpected output: %s", test.name, out.String())
		}

	}
}*/

func Test_BuildRunRequiredFlags(t *testing.T) {

	tests := []struct {
		name        string
		args        []string
		completeErr string
		executeErr  string
	}{
		{
			name:        "build name",
			args:        []string{},
			completeErr: ``,
			executeErr:  `accepts 1 arg(s), received 0`,
		},
		{
			name:        "output image",
			args:        []string{"my-build"},
			completeErr: ``,
			executeErr:  `required flag(s) "output-image" not set`,
		},
	}
	for _, tt := range tests {
		o := &BuildRunOptions{}
		t.Run(tt.name, func(t *testing.T) {
			var err error
			cmd := newBuildRunCmd(context.Background(), &genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}, &types.ClientSets{}, o)

			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Errorf("unexpected error occurred parsing flags: %#v", err)
			}

			err = o.Complete(tt.args)
			if result := testflags.CheckError(err, "Complete", tt.completeErr); len(result) != 0 {
				t.Error(result)
			}

			cmd.SetArgs(tt.args)
			_, err = cmd.ExecuteC()
			if result := testflags.CheckError(err, "Execute", tt.executeErr); len(result) != 0 {
				t.Error(result)
			}

		})
	}
}

func Test_BuildRunComplete(t *testing.T) {

	tests := []struct {
		name        string
		args        []string
		wantOptions map[string]string
		wantObject  map[string]string
	}{

		{
			name: "defaults",
			args: []string{"my-build"},
			wantObject: map[string]string{
				"spec.serviceAccount.generate": "",
				"spec.timeout":                 "0s",
			},
		},
		{
			name: "build ref flags",
			args: []string{
				"my-build",
				"--buildref-name=my-name",
				"--buildref-apiversion=my-version",
			},
			wantObject: map[string]string{
				"spec.buildRef.name":       "my-name",
				"spec.buildRef.apiVersion": "my-version",
			},
		},
		{
			name: "service account flags",
			args: []string{
				"--sa-name=my-sa-name",
				"--sa-generate",
			},
			wantObject: map[string]string{
				"spec.serviceAccount.name":     "my-sa-name",
				"spec.serviceAccount.generate": "true",
			},
		},
		{
			name: "output flags",
			args: []string{
				"--output-image=my-image",
				"--output-credentials-secret=my-input-secret",
			},
			wantObject: map[string]string{
				"spec.output.image":            "my-image",
				"spec.output.credentials.name": "my-input-secret",
			},
		},
		{
			name: "timeout flags",
			args: []string{
				"--timeout=10m0s",
			},
			wantObject: map[string]string{
				"spec.timeout": "10m0s",
			},
		},
		{
			name: "follow",
			args: []string{"--follow"},
			wantOptions: map[string]string{
				"FollowLogs": "true",
			},
		},
	}
	for _, tt := range tests {
		o := &BuildRunOptions{}
		t.Run(tt.name, func(t *testing.T) {
			var err error
			cmd := newBuildRunCmd(context.Background(), &genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}, &types.ClientSets{}, o)

			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Errorf("unexpected error occurred parsing flags: %#v", err)
			}

			if err := o.Complete(tt.args); err != nil {
				t.Errorf("unexpected error occurred executing Complete function: %#v", err)
			}

			if len(tt.wantOptions) != 0 {
				var j []byte
				if j, err = json.Marshal(o); err != nil {
					t.Fatalf("error occurred marshalling Build object into json byte array: %#v", err)
				}

				for k, v := range tt.wantOptions {
					val := gjson.Get(string(j), k)
					if v != val.String() {
						t.Errorf("expected value %q at path %q in Options, but found %q instead", v, k, val.String())
					}

				}
			}

			if len(tt.wantObject) != 0 {
				var j []byte
				if j, err = json.Marshal(o.BuildRun); err != nil {
					t.Fatalf("error occurred marshalling Build object into json byte array: %#v", err)
				}

				for k, v := range tt.wantObject {
					val := gjson.Get(string(j), k)
					if v != val.String() {
						t.Errorf("expected value %q at path %q in Object, but found %q instead", v, k, val.String())
					}

				}
			}

		})
	}
}
