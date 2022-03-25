package build

import (
	"errors"
	"fmt"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/cli/pkg/shp/cmd/buildrun"
	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/flags"
	"github.com/shipwright-io/cli/pkg/shp/follower"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/reactor"

	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// RunCommand represents the `build run` sub-command, which creates a unique BuildRun instance to run
// the build process, informed via arguments.
type RunCommand struct {
	cmd *cobra.Command // cobra command instance

	buildName string // build name, collected during complete

	buildRunSpec *buildv1alpha1.BuildRunSpec // stores command-line flags
	follow       bool                        // flag to tail pod logs

	podLogsFollower *follower.PodLogsFollower
}

const buildRunLongDesc = `
Creates a unique BuildRun instance for the given Build, which starts the build
process orchestrated by the Shipwright build controller. For example:

	$ shp build run my-app
`

// Cmd returns cobra.Command object of the create sub-command.
func (r *RunCommand) Cmd() *cobra.Command {
	return r.cmd
}

// Complete picks the build resource name from arguments, and instantiate additional components.
func (r *RunCommand) Complete(
	p params.Interface,
	ioStreams *genericclioptions.IOStreams,
	args []string,
) error {
	if len(args) != 1 {
		return errors.New("build name is not informed")
	}
	r.buildName = args[0]

	// overwriting build-ref name to use what's on arguments
	return r.Cmd().Flags().Set(flags.BuildrefNameFlag, r.buildName)
}

// Validate the user must inform the build resource name.
func (r *RunCommand) Validate() error {
	if r.buildName == "" {
		return fmt.Errorf("name is not informed")
	}
	return nil
}

// Run creates a BuildRun resource based on Build's name informed on arguments.
func (r *RunCommand) Run(p params.Interface, ioStreams *genericclioptions.IOStreams) error {
	// resource using GenerateName, which will provide a unique instance
	br := &buildv1alpha1.BuildRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", r.buildName),
		},
		Spec: *r.buildRunSpec,
	}
	flags.SanitizeBuildRunSpec(&br.Spec)

	ctx := r.cmd.Context()
	clientset, err := p.ShipwrightClientSet()
	if err != nil {
		return err
	}
	br, err = clientset.ShipwrightV1alpha1().
		BuildRuns(p.Namespace()).
		Create(ctx, br, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	fmt.Fprintf(ioStreams.Out, "BuildRun created %q for build %q\n", br.GetName(), r.buildName)
	if !r.follow {
		return nil
	}

	// during unit-testing the follower instance will be injected directly, which makes possible to
	// simulate the pod events without creating a race condition
	if r.podLogsFollower == nil {
		pw, err := reactor.NewPodWatcherFromParams(r.cmd.Context(), p)
		if err != nil {
			return err
		}
		r.podLogsFollower, err = follower.NewPodLogsFollowerFromParams(ctx, p, pw, ioStreams)
		if err != nil {
			return err
		}
	}

	// instantiating a pod watcher with a specific label-selector to find the indented pod where the
	// actual build started by this subcommand is being executed, including the randomized buildrun
	// name
	buildNameLabel := fmt.Sprintf("%s=%s", buildv1alpha1.LabelBuild, r.buildName)
	buildRunNameLabel := fmt.Sprintf("%s=%s", buildv1alpha1.LabelBuildRun, br.GetName())
	listOpts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s,%s", buildNameLabel, buildRunNameLabel),
	}

	if _, err = r.podLogsFollower.Start(listOpts); err != nil {
		buildRunName := types.NamespacedName{Namespace: br.GetNamespace(), Name: br.GetName()}
		_ = buildrun.InspectBuildRun(ctx, clientset, buildRunName, ioStreams)
	}
	return err
}

// runCmd instantiate the "build run" sub-command using common BuildRun flags.
func runCmd() runner.SubCommand {
	cmd := &cobra.Command{
		Use:   "run <name>",
		Short: "Start a build specified by 'name'",
		Long:  buildRunLongDesc,
	}
	runCommand := &RunCommand{
		cmd:          cmd,
		buildRunSpec: flags.BuildRunSpecFromFlags(cmd.Flags()),
	}
	flags.FollowFlag(cmd.Flags(), &runCommand.follow)
	return runCommand
}
