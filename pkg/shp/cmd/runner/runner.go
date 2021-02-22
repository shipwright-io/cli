package runner

import (
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/params"
)

// Runner execute the sub-command lifecycle, wrapper around sub-commands.
type Runner struct {
	p         *params.Params
	ioStreams genericclioptions.IOStreams // input, output and error io streams
	subCmd    SubCommand                  // sub-command instance
}

// Cmd is a wrapper around sub-command's Cobra, it wires up global flags and set a single RunE
// executor to self.
func (r *Runner) Cmd() *cobra.Command {
	cmd := r.subCmd.Cmd()
	cmd.RunE = r.RunE
	return cmd
}

// RunE cobra.Command's RunE implementation focusing on sub-commands lifecycle. To achieve it, a
// dynamic client and configured namespace are informed.
func (r *Runner) RunE(cmd *cobra.Command, args []string) error {
	if err := r.subCmd.Complete(r.p, args); err != nil {
		return err
	}
	if err := r.subCmd.Validate(); err != nil {
		return err
	}
	return r.subCmd.Run(r.p)
}

// NewRunner instantiate a Runner.
func NewRunner(params *params.Params, ioStreams genericclioptions.IOStreams, subCmd SubCommand) *Runner {
	return &Runner{p: params, ioStreams: ioStreams, subCmd: subCmd}
}
