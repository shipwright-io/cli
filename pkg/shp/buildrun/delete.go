package buildrun

import (
	"github.com/otaviof/shp/pkg/shp/util"
	"github.com/spf13/cobra"
)

// NewDeleteBuildRun represents the "delete build-run" sub-command.
func NewDeleteBuildRun() *BuildRun {
	cmd := &cobra.Command{
		Use:          "delete build-run [name]",
		Short:        "delete a BuildRun resource",
		SilenceUsage: true,
	}
	return newBuildRun(cmd, util.Delete)
}
