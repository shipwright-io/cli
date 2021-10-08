package build

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/tidwall/gjson"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd/types"
	testflags "github.com/shipwright-io/cli/test/flags"
)

func Test_BuildCreateRequiredFlags(t *testing.T) {

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
		o := &BuildCreateOptions{}
		t.Run(tt.name, func(t *testing.T) {
			var err error
			cmd := newBuildCreateCmd(context.Background(), &genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}, &types.ClientSets{}, o)

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

func Test_BuildCreateComplete(t *testing.T) {

	tests := []struct {
		name string
		args []string
		want map[string]string
	}{
		{
			name: "build name",
			args: []string{"my-build"},
			want: map[string]string{
				"spec.strategy.apiVersion": buildv1alpha1.SchemeGroupVersion.Version,
				"spec.strategy.kind":       string(clusterBuildStrategyKind),
				"spec.strategy.name":       "buildpacks-v3",
				"spec.timeout":             "0s",
			},
		},
		{
			name: "defaults",
			args: []string{"my-build"},
			want: map[string]string{
				"spec.strategy.apiVersion": buildv1alpha1.SchemeGroupVersion.Version,
				"spec.strategy.kind":       string(clusterBuildStrategyKind),
				"spec.strategy.name":       "buildpacks-v3",
				"spec.timeout":             "0s",
			},
		},
		{
			name: "source flags",
			args: []string{
				"--source-url=https://github.com/example/repo.git",
				"--source-revision=my-revision",
				"--source-context-dir=my-context-dir",
				"--source-credentials-secret=my-secret",
			},
			want: map[string]string{
				"spec.source.url":              "https://github.com/example/repo.git",
				"spec.source.revision":         "my-revision",
				"spec.source.contextDir":       "my-context-dir",
				"spec.source.credentials.name": "my-secret",
			},
		},
		{
			name: "strategy flags",
			args: []string{
				"--strategy-apiversion=my-apiversion",
				fmt.Sprintf("--strategy-kind=%s", string(buildv1alpha1.NamespacedBuildStrategyKind)),
				"--strategy-name=my-name",
			},
			want: map[string]string{
				"spec.strategy.apiVersion": "my-apiversion",
				"spec.strategy.kind":       string(buildv1alpha1.NamespacedBuildStrategyKind),
				"spec.strategy.name":       "my-name",
			},
		},
		{
			name: "docker flags",
			args: []string{
				"--dockerfile=my-dockerfile",
			},
			want: map[string]string{
				"spec.dockerfile": "my-dockerfile",
			},
		},
		{
			name: "builder flags",
			args: []string{
				"--input-image=my-image",
				"--input-credentials-secret=my-input-secret",
			},
			want: map[string]string{
				"spec.builder.image":            "my-image",
				"spec.builder.credentials.name": "my-input-secret",
			},
		},
		{
			name: "output flags",
			args: []string{
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
				"--timeout=10m0s",
			},
			want: map[string]string{
				"spec.timeout": "10m0s",
			},
		},
	}
	for _, tt := range tests {
		o := &BuildCreateOptions{}
		t.Run(tt.name, func(t *testing.T) {
			var err error
			cmd := newBuildCreateCmd(context.Background(), &genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}, &types.ClientSets{}, o)

			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Errorf("unexpected error occurred parsing flags: %#v", err)
			}

			if err := o.Complete(tt.args); err != nil {
				t.Errorf("unexpected error occurred executing Complete function: %#v", err)
			}

			var j []byte
			if j, err = json.Marshal(o.Build); err != nil {
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
