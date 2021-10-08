package build

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/shipwright-io/cli/pkg/shp/cmd/types"
	"github.com/tidwall/gjson"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func Test_BuildListComplete(t *testing.T) {

	tests := []struct {
		name string
		args []string
		want map[string]string
	}{
		{
			name: "no header",
			args: []string{"--no-header"},
			want: map[string]string{
				"NoHeader": "true",
			},
		},
	}
	for _, tt := range tests {
		o := &BuildListOptions{}
		t.Run(tt.name, func(t *testing.T) {
			var err error
			cmd := newBuildListCmd(context.Background(), &genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}, &types.ClientSets{}, o)

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
