package buildrun

import (
	"github.com/otaviof/shp/pkg/shp/util"
	"github.com/spf13/cobra"
)

// NewCreateBuildRun represents the "create build-run" sub-command.
func NewCreateBuildRun() *BuildRun {
	cmd := &cobra.Command{
		Use:          "create build-run [name]",
		Short:        "create a new BuildRun resource",
		SilenceUsage: true,
	}
	return newBuildRun(cmd, util.Create)
}

// NewRunBuild represents the "run build" sub-command.
func NewRunBuild() *BuildRun {
	cmd := &cobra.Command{
		Use:          "run build [name]",
		Short:        "run the build process",
		SilenceUsage: true,
	}
	return newBuildRun(cmd, util.Create)
}
