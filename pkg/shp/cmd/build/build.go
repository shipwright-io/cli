package build

import (
	"os"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/resource"
)

var (
	buildResource *resource.Resource
)

// Command returns Build subcommand of Shipwright CLI
// for interaction with shipwright builds
func Command(p *params.Params) *cobra.Command {
	buildResource = resource.NewShpResource(
		p,
		buildv1alpha1.SchemeGroupVersion,
		"Build",
		"builds",
	)

	command := &cobra.Command{
		Use:     "build",
		Aliases: []string{"bd"},
		Short:   "Manage Builds",
		Annotations: map[string]string{
			"commandType": "main",
		},
	}

	streams := genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	// TODO: add support for `update` and `get` commands
	command.AddCommand(
		runner.NewRunner(p, streams, createCmd()).Cmd(),
		runner.NewRunner(p, streams, listCmd()).Cmd(),
		runner.NewRunner(p, streams, deleteCmd()).Cmd(),
	)
	return command
}
