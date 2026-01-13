package build // nolint:revive

import (
	"errors"
	"fmt"

	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	"github.com/shipwright-io/cli/pkg/shp/cmd/follower"
	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/flags"
	"github.com/shipwright-io/cli/pkg/shp/params"

	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// RunCommand represents the `build run` sub-command, which creates a unique BuildRun instance to run
// the build process, informed via arguments.
type RunCommand struct {
	cmd *cobra.Command // cobra command instance

	buildName     string
	namespace     string
	buildRunSpec  *buildv1beta1.BuildRunSpec // stores command-line flags
	follow        bool                       // flag to tail pod logs
	follower      *follower.Follower
	followerReady chan bool
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
func (r *RunCommand) Complete(params *params.Params, ioStreams *genericclioptions.IOStreams, args []string) error {
	switch len(args) {
	case 1:
		r.buildName = args[0]
	default:
		return errors.New("build name is not informed")
	}

	r.namespace = params.Namespace()

	if r.follow {
		var err error
		// provide empty build run name; will be set in Run()
		r.follower, err = params.NewFollower(r.cmd.Context(), types.NamespacedName{}, ioStreams)
		if err != nil {
			return err
		}
		r.followerReady = make(chan bool, 1)
	}
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

// FollowerReady blocks until the any log following connections are established in the Run call.
// Useful if you have code that calls Run on a separate thread and coordination is needed.
func (r *RunCommand) FollowerReady() bool {
	if !r.follow {
		return false
	}
	_, closed := <-r.followerReady
	return !closed
}

// Run creates a BuildRun resource based on Build's name informed on arguments.
func (r *RunCommand) Run(params *params.Params, ioStreams *genericclioptions.IOStreams) error {
	// resource using GenerateName, which will provide a unique instance
	br := &buildv1beta1.BuildRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", r.buildName),
		},
		Spec: *r.buildRunSpec,
	}
	flags.SanitizeBuildRunSpec(&br.Spec)

	ctx := r.cmd.Context()
	clientset, err := params.ShipwrightClientSet()
	if err != nil {
		return err
	}
	br, err = clientset.ShipwrightV1beta1().BuildRuns(r.namespace).Create(ctx, br, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	if !r.follow {
		fmt.Fprintf(ioStreams.Out, "BuildRun created %q for build %q\n", br.GetName(), r.buildName)
		return nil
	}

	buildRun := types.NamespacedName{Namespace: r.namespace, Name: br.GetName()}
	r.follower.SetBuildRunName(buildRun)

	// instantiating a pod watcher with a specific label-selector to find the indented pod where the
	// actual build started by this subcommand is being executed, including the randomized buildrun
	// name
	listOpts := metav1.ListOptions{LabelSelector: fmt.Sprintf(
		"build.shipwright.io/name=%s,buildrun.shipwright.io/name=%s",
		r.buildName,
		br.GetName(),
	)}
	err = r.follower.Connect(listOpts)
	if err != nil {
		return err
	}
	close(r.followerReady)
	_, err = r.follower.WaitForCompletion()
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
