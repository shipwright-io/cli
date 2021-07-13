package buildrun

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/spf13/cobra"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/cli/pkg/shp/bundle"
	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/flags"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/reactor"
	"github.com/shipwright-io/cli/pkg/shp/tail"
)

// CreateCommand reprents the build's create subcommand.
type CreateCommand struct {
	cmd *cobra.Command

	name         string
	buildRunSpec *buildv1alpha1.BuildRunSpec
	buildRunOpts *flags.BuildRunOpts
}

const buildRunCreateLongDesc = `
Creates a new BuildRun instance using the given name, and requires --buildref-name to
find the Build object. Example:

	$ shp buildrun create my-app-build --buildref-name="..."
`

// Cmd returns cobra.Command object of the create sub-command.
func (c *CreateCommand) Cmd() *cobra.Command {
	return c.cmd
}

// Complete checks if the arguments is informing the BuildRun name.
func (c *CreateCommand) Complete(params *params.Params, args []string) error {
	switch len(args) {
	case 0:
		c.name = ""

	case 1:
		c.name = args[0]

	default:
		return fmt.Errorf("wrong amount of arguments, expected only one (specific name) or none (generated name)")
	}

	return nil
}

// Validate makes sure a name is informed.
func (c *CreateCommand) Validate() error {
	if c.buildRunSpec.BuildRef == nil || c.buildRunSpec.BuildRef.Name == "" {
		return fmt.Errorf("build name is not informed")
	}

	return nil
}

// Run executes the creation of BuildRun object.
func (c *CreateCommand) Run(params *params.Params, ioStreams *genericclioptions.IOStreams) error {
	return CreateBuildRun(
		c.cmd.Context(),
		c.name,
		params,
		ioStreams,
		c.buildRunSpec,
		c.buildRunOpts,
	)
}

// createCmd instantiate a new CreateCommand, by wiring it as a cobra.Command and registering the
// flags and marking flags required.
func createCmd() runner.SubCommand {
	cmd := &cobra.Command{
		Use:   "create <name> [flags]",
		Short: "Creates a BuildRun instance.",
		Long:  buildRunCreateLongDesc,
	}

	createCommand := &CreateCommand{
		cmd:          cmd,
		buildRunSpec: flags.BuildRunSpecFromFlags(cmd.Flags()),
		buildRunOpts: flags.BuildRunOptsFromFlags(cmd.Flags()),
	}

	if err := cmd.MarkFlagRequired(flags.BuildrefNameFlag); err != nil {
		panic(err)
	}

	return createCommand
}

// CreateBuildRun creates a BuildRun based on the given name and spec using the
// provided end-user options/settings, for example printing the pod container
// logs of the build.
func CreateBuildRun(
	ctx context.Context,
	buildRunName string,
	params *params.Params,
	ioStreams *genericclioptions.IOStreams,
	buildRunSpec *buildv1alpha1.BuildRunSpec,
	buildRunOpts *flags.BuildRunOpts) error {

	// The prune-bundle, and follow flags implicitly enable the wait flag
	if buildRunOpts.PruneBundle || buildRunOpts.Follow {
		buildRunOpts.Wait = true
	}

	shpClient, err := params.ShipwrightClientSet()
	if err != nil {
		return err
	}

	// Make sure the referenced build of the buildrun does exist
	build, err := shpClient.ShipwrightV1alpha1().Builds(params.Namespace()).Get(ctx, buildRunSpec.BuildRef.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get referenced build %q: %w", buildRunSpec.BuildRef.Name, err)
	}

	// Setup the buildrun to be created
	buildRun := &buildv1alpha1.BuildRun{Spec: *buildRunSpec}
	flags.SanitizeBuildRunSpec(&buildRun.Spec)

	// Enable usage of generated buildrun names in case there is none given
	if buildRunName == "" {
		buildRun.GenerateName = buildRunSpec.BuildRef.Name
	}

	// Local source code mode:
	// Make sure to bundle the configured local source directory into an image
	// bundle and push it to the provided target registry.
	var digest name.Digest
	if build.Spec.Source.BundleContainer.Image != "" {
		if digest, err = bundle.Push(ctx, ioStreams, buildRunOpts.LocalSourceDir, build.Spec.Source.BundleContainer.Image); err != nil {
			return err
		}
	}

	// Initiate the actual create command in the cluster
	buildRun, err = shpClient.ShipwrightV1alpha1().BuildRuns(params.Namespace()).Create(ctx, buildRun, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create BuildRun: %v", err)
	}

	buildRunName = buildRun.Name
	fmt.Fprintf(ioStreams.Out, "BuildRun created %q for Build %q\n", buildRun.Name, buildRun.Spec.BuildRef.Name)
	if !buildRunOpts.Wait {
		return nil
	}

	kubeClient, err := params.ClientSet()
	if err != nil {
		return err
	}

	podWatcher, err := reactor.NewPodWatcher(ctx, kubeClient, podByLabelSelector(build, buildRun), params.Namespace())
	if err != nil {
		return err
	}

	tailLogsStarted := make(map[string]bool)
	logTail := tail.NewTail(ctx, kubeClient)
	podPrevState := ""

	podWatcher.WithOnPodModifiedFn(func(pod *v1.Pod) error {
		if podPrevState != string(pod.Status.Phase) {
			fmt.Fprintf(ioStreams.Out, "Pod %q is in phase %q ...\n", pod.GetName(), string(pod.Status.Phase))
			podPrevState = string(pod.Status.Phase)
		}

		switch pod.Status.Phase {
		case v1.PodRunning:
			if buildRunOpts.Follow {
				// graceful time to wait for container start
				time.Sleep(3 * time.Second)

				// start tailing container logs
				for _, container := range append(pod.Spec.InitContainers, pod.Spec.Containers...) {
					if _, exists := tailLogsStarted[container.Name]; exists {
						continue
					}

					tailLogsStarted[container.Name] = true
					logTail.Start(
						pod.GetNamespace(),
						pod.GetName(),
						container.Name,
					)
				}
			}

		case v1.PodFailed:
			br, err := shpClient.ShipwrightV1alpha1().BuildRuns(params.Namespace()).Get(ctx, buildRunName, metav1.GetOptions{})
			switch {
			case err == nil && br.IsCanceled():
				fmt.Fprintf(ioStreams.Out, "BuildRun %q has been canceled.\n", br.Name)

			case err == nil && br.DeletionTimestamp != nil:
				fmt.Fprintf(ioStreams.Out, "BuildRun %q has been deleted.\n", br.Name)

			case pod.DeletionTimestamp != nil:
				fmt.Fprintf(ioStreams.Out, "Pod %q has been deleted.\n", pod.GetName())

			default:
				fmt.Fprintf(ioStreams.Out, "Pod %q has failed!\n", pod.GetName())
				err = fmt.Errorf("build pod %q has failed", pod.GetName())
			}

			logTail.Stop()
			podWatcher.Stop()
			return err

		case v1.PodSucceeded:
			logTail.Stop()
			podWatcher.Stop()

		default:
			// handle any issues with pulling images that may fail
			for _, c := range pod.Status.Conditions {
				if c.Type == v1.PodInitialized || c.Type == v1.ContainersReady {
					if c.Status == v1.ConditionUnknown {
						return fmt.Errorf(c.Message)
					}
				}
			}
		}

		return nil
	})

	_, err = podWatcher.Start()
	if err != nil {
		return err
	}

	// If configured, delete the source code bundle image once buildrun is done
	if buildRunOpts.PruneBundle {
		fmt.Fprintf(ioStreams.Out, "Pruning source code bundle %q\n", digest.String())
		if err := bundle.Prune(ctx, ioStreams, digest.String()); err != nil {
			return err
		}
	}

	return nil
}

func podByLabelSelector(build *buildv1alpha1.Build, buildRun *buildv1alpha1.BuildRun) metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: fmt.Sprintf(
			"build.shipwright.io/name=%s,buildrun.shipwright.io/name=%s",
			build.Name,
			buildRun.Name,
		),
	}
}
