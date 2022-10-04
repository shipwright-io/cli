package build

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/build/pkg/reconciler/buildrun/resources/sources"

	"github.com/shipwright-io/cli/pkg/shp/bundle"
	"github.com/shipwright-io/cli/pkg/shp/cmd/follower"
	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/flags"
	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/reactor"
	"github.com/shipwright-io/cli/pkg/shp/streamer"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// UploadCommand represents the "build upload" subcommand, implements runner.SubCommand interface.
type UploadCommand struct {
	cmd          *cobra.Command              // cobra command instance
	buildRunSpec *buildv1alpha1.BuildRunSpec // command-line flags stored directly on the BuildRun
	follow       bool                        // flag to tail pod logs

	buildRefName string // build name
	sourceDir    string // local directory to be streamed

	dataStreamer    *streamer.Streamer // tar streamer instance
	streamingIsDone bool               // marks the streaming is completed

	sourceBundleImage string // image to be used as the source bundle

	pw       *reactor.PodWatcher // pod-watcher instance
	follower *follower.Follower  // follower instance
}

const (
	buildRunUploadLongDesc = `
Creates a new BuildRun instance and instructs the Build Controller to use data from a local directory
to be used for the Build. Two options are supported: streaming and bundling. With these, you can
employ Shipwright Builds from a local repository clone.

When streaming is used, the Build Controller waits for the data being streamed to the build pod,
instead of executing "git clone". The upload skips the ".git" directory completely, and it follows
the ".gitignore" directives, when the file is found at the root of the directory uploaded.

In case a source bundle image is defined, the bundling feature is used, which will bundle the local
source code into a bundle container and upload it to the specified container registry. Instead of
executing using Git in the source step, it will use the container registry to obtain the source code.

	$ shp buildrun upload <build-name>
	$ shp buildrun upload <build-name> /path/to/repository
`

	// targetBaseDir directory where data will be uploaded.
	targetBaseDir = "/workspace/source"
	// buildNameAnnotation label to identify the Build name.
	buildNameAnnotation = "build.shipwright.io/name"
	// buildRunNameAnnotation label to identify the BuildRun name.
	buildRunNameAnnotation = "buildrun.shipwright.io/name"
)

// Cmd exposes the Cobra command instance.
func (u *UploadCommand) Cmd() *cobra.Command {
	return u.cmd
}

// extractArgs inspect the command-line arguments to extract the name and source directory path.
func (u *UploadCommand) extractArgs(args []string) error {
	switch len(args) {
	case 1:
		u.buildRefName = args[0]
		u.sourceDir = "."
	case 2:
		u.buildRefName = args[0]
		u.sourceDir = args[1]
	default:
		return fmt.Errorf("wrong amount of arguments, expected one or two")
	}

	if u.sourceDir == "." {
		var err error
		if u.sourceDir, err = os.Getwd(); err != nil {
			return err
		}
	}
	// making sure the final path is absolute and clean up
	u.sourceDir = path.Clean(u.sourceDir)

	// overwriting build-ref name to use what's on arguments
	return u.Cmd().Flags().Set(flags.BuildrefNameFlag, u.buildRefName)
}

// Complete instantiate the dependencies for the log following and the data streaming.
func (u *UploadCommand) Complete(p *params.Params, _ *genericclioptions.IOStreams, args []string) error {
	// extracting the command-line arguments to store the build-name and the path to the directory
	// to be uploaded, in subsequent steps
	if err := u.extractArgs(args); err != nil {
		return err
	}

	restConfig, err := p.RESTConfig()
	if err != nil {
		return err
	}
	clientset, err := p.ClientSet()
	if err != nil {
		return err
	}
	shpClientSet, err := p.ShipwrightClientSet()
	if err != nil {
		return err
	}

	// check that the cluster actually contains a build with this name
	build, err := shpClientSet.ShipwrightV1alpha1().Builds(p.Namespace()).Get(u.cmd.Context(), u.buildRefName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// detect upload method, if build has bundle container image set, it
	// is assumed that the source bundle upload via registry is used
	if build.Spec.Source.BundleContainer != nil && build.Spec.Source.BundleContainer.Image != "" {
		u.sourceBundleImage = build.Spec.Source.BundleContainer.Image

	} else {
		u.dataStreamer = streamer.NewStreamer(restConfig, clientset)
	}

	u.pw, err = p.NewPodWatcher(u.Cmd().Context())
	return err
}

// Validate the current subcommand state, make sure the directory to be uploaded exists.
func (u *UploadCommand) Validate() error {
	stat, err := os.Stat(u.sourceDir)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return fmt.Errorf("informed path is not a directory: '%s'", u.sourceDir)
	}
	return nil
}

// createBuildRun creates the BuildRun instance to receive the data upload afterwards, it returns the
// BuildRun name just created and error.
func (u *UploadCommand) createBuildRun(p *params.Params) (*buildv1alpha1.BuildRun, error) {
	var br *buildv1alpha1.BuildRun
	switch {
	// Use bundle feature for source upload and build
	case u.sourceBundleImage != "":
		br = &buildv1alpha1.BuildRun{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: fmt.Sprintf("%s-", u.buildRefName),
			},
			Spec: *u.buildRunSpec,
		}

	// Use local copy streaming feature for source upload and build
	default:
		u.buildRunSpec.Sources = []buildv1alpha1.BuildSource{{
			Name: "local-copy",
			Type: buildv1alpha1.LocalCopy,
		}}
		br = &buildv1alpha1.BuildRun{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: fmt.Sprintf("%s-", u.buildRefName),
			},
			Spec: *u.buildRunSpec,
		}
	}

	flags.SanitizeBuildRunSpec(&br.Spec)

	ns := p.Namespace()
	log.Printf("Creating a BuildRun for '%s/%s' Build...", ns, u.buildRefName)
	clientset, err := p.ShipwrightClientSet()
	if err != nil {
		return nil, err
	}
	br, err = clientset.ShipwrightV1alpha1().
		BuildRuns(ns).
		Create(u.cmd.Context(), br, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	log.Printf("BuildRun '%s' created!", br.GetName())
	return br, nil
}

// performDataStreaming execute the data transfer process end-to-end.
func (u *UploadCommand) performDataStreaming(target *streamer.Target) error {
	if u.streamingIsDone {
		return nil
	}

	fmt.Fprintf(os.Stdout, "Streaming %q to the Build POD %q ...\n", u.sourceDir, target.Pod)
	// creates an in-memory tarball with source directory data, and ready to start data streaming
	tarball, err := streamer.NewTar(u.sourceDir)
	if err != nil {
		return err
	}

	tarsize, err := streamer.GetTarSize(u.sourceDir)
	if err != nil {
		return err
	}

	// start writing the data using the tarball format, and streaming it via STDIN, which is
	// redirected to the correct container
	if err = u.dataStreamer.Stream(target, tarball.Create, tarsize.Size); err != nil {
		return err
	}

	// graceful waiting for the container finish writing the streamed data, and right after, calling
	// done on the container, so the rest of the build process can continue and use streamed data
	time.Sleep(5 * time.Second)
	if err = u.dataStreamer.Done(target); err != nil {
		return err
	}

	u.streamingIsDone = true
	return nil
}

// stop following logs and watch over pod.
func (u *UploadCommand) stop() {
	if u.follower != nil {
		u.follower.Stop()
	}
	u.pw.Stop()
}

// onPodModifiedEventStreaming is invoked everytime the pod running the actual build process changes, thus it
// can react upon the state changes in order to orchestrate the data upload.
func (u *UploadCommand) onPodModifiedEventStreaming(pod *corev1.Pod) error {
	switch pod.Status.Phase {
	case corev1.PodRunning:
		return u.performDataStreaming(&streamer.Target{
			Namespace: pod.GetNamespace(),
			Pod:       pod.GetName(),
			Container: fmt.Sprintf("step-%s", sources.WaiterContainerName),
			BaseDir:   targetBaseDir,
		})
	case corev1.PodFailed:
		u.stop()
		return fmt.Errorf("build pod '%s' has failed", pod.GetName())
	case corev1.PodSucceeded:
		u.stop()
	}
	return nil
}

func (u *UploadCommand) onPodModifiedEventBundling(pod *corev1.Pod) error {
	switch pod.Status.Phase {
	case corev1.PodFailed:
		u.stop()
		return fmt.Errorf("build pod '%s' has failed", pod.GetName())

	case corev1.PodSucceeded:
		u.stop()
	}

	return nil
}

// Run executes the primary business logic of this subcommand, by starting to watch over the build
// pod status and react accordingly.
func (u *UploadCommand) Run(p *params.Params, ioStreams *genericclioptions.IOStreams) error {
	// creating a BuildRun with settings for the local source upload
	br, err := u.createBuildRun(p)
	if err != nil {
		return err
	}

	if u.follow {
		// when follow flag is enabled, instantiating the "follower" to live tail logs
		if u.follower, err = p.NewFollower(u.Cmd().Context(), types.NamespacedName{Namespace: br.Namespace, Name: br.Name}, ioStreams); err != nil {
			return err
		}
	}

	switch {
	// Using bundling to upload local source code
	case u.sourceBundleImage != "":
		_, err = bundle.Push(u.cmd.Context(), ioStreams, u.sourceDir, u.sourceBundleImage)
		if err != nil {
			return err
		}

		u.pw.WithOnPodModifiedFn(u.onPodModifiedEventBundling)

	// Using streaming to upload local source code
	default:
		// registering the routine that will react upon build pod state changes
		u.pw.WithOnPodModifiedFn(u.onPodModifiedEventStreaming)
	}

	// preparing a label-selector with annotations that can pinpoint the exact pod created for the
	// BuildRun we've just issued
	labelSelector := fmt.Sprintf(
		"%s=%s,%s=%s",
		buildNameAnnotation, u.buildRunSpec.BuildRef.Name,
		buildRunNameAnnotation, br.Name,
	)
	listOpts := metav1.ListOptions{LabelSelector: labelSelector}

	// starting the event reactor with the ListOptions instance to find the desired pod, as the pod
	// status changes, different routines are issued
	_, err = u.pw.Start(listOpts)
	return err
}

// uploadCmd instantiate the "upload" subcommand by creating the cobra command and its flags.
func uploadCmd() runner.SubCommand {
	cmd := &cobra.Command{
		Use:          "upload <build-name> [path/to/source|.]",
		Short:        "Run a Build with local data",
		Long:         buildRunUploadLongDesc,
		SilenceUsage: true,
	}
	u := &UploadCommand{
		cmd:          cmd,
		buildRunSpec: flags.BuildRunSpecFromFlags(cmd.Flags()),
		follow:       false,
	}
	flags.FollowFlag(cmd.Flags(), &u.follow)
	return u
}
