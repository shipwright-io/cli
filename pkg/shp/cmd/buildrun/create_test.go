package buildrun

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/tidwall/gjson"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd/types"
	testflags "github.com/shipwright-io/cli/test/flags"
)

func Test_BuildRunCreateRequiredFlags(t *testing.T) {

	tests := []struct {
		name        string
		args        []string
		completeErr string
		executeErr  string
	}{
		{
			name:        "required flags no name",
			args:        []string{},
			completeErr: `argument list is empty`,
			executeErr:  `accepts 1 arg(s), received 0`,
		},
		{
			name:        "required flags",
			args:        []string{"my-build"},
			completeErr: "",
			executeErr:  `required flag(s) "output-image", "source-url" not set`,
		},
	}
	for _, tt := range tests {
		o := &BuildRunCreateOptions{}
		t.Run(tt.name, func(t *testing.T) {
			var err error
			cmd := newBuildRunCreateCmd(context.Background(), &genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}, &types.ClientSets{}, o)

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

func Test_BuildRunCreateComplete(t *testing.T) {

	tests := []struct {
		name string
		args []string
		want map[string]string
	}{
		{
			name: "buildrun name",
			args: []string{"my-buildrun"},
			want: map[string]string{
				"metadata.name": "my-buildrun",
			},
		},
		{
			name: "defaults",
			args: []string{"my-buildrun"},
			want: map[string]string{
				"spec.timeout": "0s",
			},
		},
		{
			name: "build flags",
			args: []string{
				"my-buildrun",
				"--buildref-name=my-build",
				"--buildref-apiversion=my-version",
			},
			want: map[string]string{
				"spec.buildRef.name":       "my-build",
				"spec.buildRef.apiVersion": "my-version",
			},
		},
		{
			name: "service account name flag",
			args: []string{
				"my-buildrun",
				"--sa-name=my-sa-name",
				"--sa-generate=true",
			},
			want: map[string]string{
				"spec.serviceAccount.name":     "my-sa-name",
				"spec.serviceAccount.generate": "",
			},
		},
		{
			name: "service account generate flag",
			args: []string{
				"my-buildrun",
				"--sa-generate=true",
			},
			want: map[string]string{
				"spec.serviceAccount.name":     "",
				"spec.serviceAccount.generate": "true",
			},
		},
		{
			name: "output flags",
			args: []string{
				"my-buildrun",
				"--output-image=my-image",
				"--output-credentials-secret=my-input-secret",
			},
			want: map[string]string{
				"spec.output.image":            "my-image",
				"spec.output.credentials.name": "my-input-secret",
			},
		},
		{
			name: "env flags",
			args: []string{
				"my-buildrun",
				"--env=VAR_1=var-1-value",
				"--env=VAR_2=var-2-value",
			},
			want: map[string]string{
				"spec.env.0.name":  "VAR_1",
				"spec.env.0.value": "var-1-value",
				"spec.env.1.name":  "VAR_2",
				"spec.env.1.value": "var-2-value",
			},
		},
		{
			name: "timeout flags",
			args: []string{
				"my-buildrun",
				"--timeout=10m0s",
			},
			want: map[string]string{
				"spec.timeout": "10m0s",
			},
		},
	}
	for _, tt := range tests {
		o := &BuildRunCreateOptions{}
		t.Run(tt.name, func(t *testing.T) {
			var err error
			cmd := newBuildRunCreateCmd(context.Background(), &genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}, &types.ClientSets{}, o)

			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Errorf("unexpected error occurred parsing flags: %#v", err)
			}

			if err := o.Complete(tt.args); err != nil {
				t.Errorf("unexpected error occurred executing Complete function: %#v", err)
			}

			var j []byte
			if j, err = json.Marshal(o.BuildRun); err != nil {
				t.Fatalf("error occurred marshalling BuildRun object into json byte array: %#v", err)
			}

			for k, v := range tt.want {
				val := gjson.Get(string(j), k)
				if v != val.String() {
					t.Errorf("expected value %q at path %q in BuildRun object, but found %q instead", v, k, val.String())
				}
			}
			spew.Dump(string(j))
		})
	}
}
