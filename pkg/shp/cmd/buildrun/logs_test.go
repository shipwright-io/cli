package buildrun

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/tidwall/gjson"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd/types"
	testflags "github.com/shipwright-io/cli/test/flags"
)

func Test_BuildRunLogsRequiredFlags(t *testing.T) {

	tests := []struct {
		name        string
		args        []string
		completeErr string
		executeErr  string
	}{
		{
			name:        "required flags",
			args:        []string{},
			completeErr: `argument list is empty`,
			executeErr:  `accepts 1 arg(s), received 0`,
		},
	}
	for _, tt := range tests {
		o := &BuildRunLogsOptions{}
		t.Run(tt.name, func(t *testing.T) {
			var err error
			cmd := newBuildRunLogsCmd(context.Background(), &genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}, &types.ClientSets{}, o)

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

func Test_BuildRunLogsComplete(t *testing.T) {

	tests := []struct {
		name string
		args []string
		want map[string]string
	}{
		{
			name: "buildrun name",
			args: []string{"my-buildrun"},
			want: map[string]string{
				"BuildRunName": "my-buildrun",
			},
		},
	}
	for _, tt := range tests {
		o := &BuildRunLogsOptions{}
		t.Run(tt.name, func(t *testing.T) {
			var err error
			cmd := newBuildRunLogsCmd(context.Background(), &genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}, &types.ClientSets{}, o)

			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Errorf("unexpected error occurred parsing flags: %#v", err)
			}

			if err := o.Complete(tt.args); err != nil {
				t.Errorf("unexpected error occurred executing Complete function: %#v", err)
			}

			var j []byte
			if j, err = json.Marshal(o); err != nil {
				t.Fatalf("error occurred marshalling BuildRun object into json byte array: %#v", err)
			}

			for k, v := range tt.want {
				val := gjson.Get(string(j), k)
				if v != val.String() {
					t.Errorf("expected value %q at path %q in BuildRun object, but found %q instead", v, k, val.String())
				}

			}

		})
	}
}

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

	ioStreams, _, out, _ := genericclioptions.NewTestIOStreams()

	cmd := NewBuildRunLogsCmd(context.Background(), &ioStreams, &types.ClientSets{})

	cmd.SetArgs([]string{name})
	_, err := cmd.ExecuteC()
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	if !strings.Contains(out.String(), "fake logs") {
		t.Fatalf("unexpected output: %s", out.String())
	}

	t.Logf("%s", out.String())

}
