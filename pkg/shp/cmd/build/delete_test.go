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

func Test_BuildDeleteRequiredFlags(t *testing.T) {

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
		o := &BuildDeleteOptions{}
		t.Run(tt.name, func(t *testing.T) {
			var err error
			cmd := newBuildDeleteCmd(context.Background(), &genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}, &types.ClientSets{}, o)

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

func Test_BuildDeleteComplete(t *testing.T) {

	tests := []struct {
		name string
		args []string
		want map[string]string
	}{
		{
			name: "build name",
			args: []string{"my-build"},
			want: map[string]string{
				"BuildName": "my-build",
			},
		},
		{
			name: "delete runs",
			args: []string{"my-build", "--delete-runs"},
			want: map[string]string{
				"BuildName":  "my-build",
				"DeleteRuns": "true",
			},
		},
	}
	for _, tt := range tests {
		o := &BuildDeleteOptions{}
		t.Run(tt.name, func(t *testing.T) {
			var err error
			cmd := newBuildDeleteCmd(context.Background(), &genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}, &types.ClientSets{}, o)

			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Errorf("unexpected error occurred parsing flags: %#v", err)
			}

			if err := o.Complete(tt.args); err != nil {
				t.Errorf("unexpected error occurred executing Complete function: %#v", err)
			}

			var j []byte
			if j, err = json.Marshal(o); err != nil {
				t.Fatalf("error occurred marshalling Build object into json byte array: %#v", err)
			}

			for k, v := range tt.want {
				val := gjson.Get(string(j), k)
				if v != val.String() {
					t.Errorf("expected value %q at path %q in Build object, but found %q instead", v, k, val.String())
				}

			}

		})
	}
}
