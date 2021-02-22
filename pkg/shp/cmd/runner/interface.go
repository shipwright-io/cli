package runner

import (
	"github.com/spf13/cobra"

	"github.com/shipwright-io/cli/pkg/shp/params"
)

// SubCommand defines the methods for a sub-command wrapped with Runner.
type SubCommand interface {
	// Cmd shares the cobra.Command instance.
	Cmd() *cobra.Command
	// Complete aggregate data needed for the sub-command primary logic.
	Complete(params *params.Params, args []string) error
	// Validate perform validation against the context collected.
	Validate() error
	// Run execute the primary sub-command logic.
	Run(params *params.Params) error
}
