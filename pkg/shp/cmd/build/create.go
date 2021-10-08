package build

import (
	"context"
	"errors"
	"fmt"
	"time"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/shipwright-io/cli/pkg/shp/cmd/types"
	"github.com/shipwright-io/cli/pkg/shp/util"
)

var (
	// Long description for the "build create" command
	buildCreateLongDescription = templates.LongDesc(`
		Creates a new Build
	`)

	// Examples for using the "build create" command
	buildCreateExamples = templates.Examples(`
		$ shp build create my-app --source-url=https://example.org/some/repo --output-image=some-image
	`)

	clusterBuildStrategyKind = buildv1alpha1.ClusterBuildStrategyKind
)

// BuildCreateOptions stores data passed to the command via command line flags
type BuildCreateOptions struct {
	types.SharedOptions

	Build *buildv1alpha1.Build

	Name string

	SourceURL                   string
	SourceRevision              string
	SourceContextDir            string
	SourceCredentialsSecretName string

	StrategyAPIVersion string
	StrategyKind       string
	StrategyName       string

	Dockerfile string

	BuilderImage                 string
	BuilderCredentialsSecretName string

	OutputImage                 string
	OutputCredentialsSecretName string

	Envs []string

	Timeout metav1.Duration
}

// newBuildCreateCmd creates the "build create" command
func newBuildCreateCmd(ctx context.Context, ioStreams *genericclioptions.IOStreams, clients *types.ClientSets, o *BuildCreateOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create [flags]",
		Short:   "Create a new Build",
		Long:    buildCreateLongDescription,
		Args:    cobra.ExactArgs(1),
		Example: buildCreateExamples,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(args))
			cmdutil.CheckErr(o.Run())
		},
	}

	cmd.Flags().StringVar(&o.SourceURL, "source-url", "", "The URL of the git repository to use as the source for the build")
	cmd.MarkFlagRequired("source-url")
	cmd.Flags().StringVar(&o.SourceRevision, "source-revision", "", "The version of the source to use.")
	cmd.Flags().StringVar(&o.SourceContextDir, "source-context-dir", "", "The directory within the git repository to use.")
	cmd.Flags().StringVar(&o.SourceCredentialsSecretName, "source-credentials-secret", "", "The name of the Secret that contains credentials to access the git repository.")

	cmd.Flags().StringVar(&o.StrategyAPIVersion, "strategy-apiversion", buildv1alpha1.SchemeGroupVersion.Version, "The API version of the build strategy to use.")
	cmd.Flags().StringVar(&o.StrategyKind, "strategy-kind", "", "The Kind of build strategy to use.")
	cmd.Flags().StringVar(&o.StrategyName, "strategy-name", "buildpacks-v3", "The name of the build strategy to use.")

	cmd.Flags().StringVar(&o.Dockerfile, "dockerfile", "", "The path of the Dockerfile to use, relative to the repository root.")

	cmd.Flags().StringVar(&o.BuilderImage, "input-image", "", "The builder image to use during the build process.")
	cmd.Flags().StringVar(&o.BuilderCredentialsSecretName, "input-credentials-secret", "", "The name of the Secret that contains credentials to pull the provided builder image.")

	cmd.Flags().StringVar(&o.OutputImage, "output-image", "", "The location to push the output image to.")
	cmd.MarkFlagRequired("output-image")
	cmd.Flags().StringVar(&o.OutputCredentialsSecretName, "output-credentials-secret", "", "The name of the Secret that contains credentials for the repository to push the built Image to.")

	cmd.Flags().StringArrayVarP(&o.Envs, "env", "e", []string{}, "specify a key-value pair for an environment variable to set for the build container")

	cmd.Flags().DurationVar(&o.Timeout.Duration, "timeout", time.Duration(0), "How long to let the build run before timing out.")

	return cmd
}

// NewBuildCreateCmd is a wrapper for newBuildCreateCmd
func NewBuildCreateCmd(ctx context.Context, ioStreams *genericclioptions.IOStreams, clients *types.ClientSets) *cobra.Command {
	o := &BuildCreateOptions{
		SharedOptions: types.SharedOptions{
			Clients: clients,
			Context: ctx,
			Streams: ioStreams,
		},
	}

	return newBuildCreateCmd(ctx, ioStreams, clients, o)
}

// Complete processes any data that is needed before Run executes
func (o *BuildCreateOptions) Complete(args []string) error {
	// Guard against index out of bound errors
	if len(args) == 0 {
		return errors.New("argument list is empty")
	}

	o.Name = args[0]

	o.Build = &buildv1alpha1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name: o.Name,
		},
		Spec: buildv1alpha1.BuildSpec{
			Source: buildv1alpha1.Source{
				URL:        o.SourceURL,
				Revision:   &o.SourceRevision,
				ContextDir: &o.SourceContextDir,
			},
			Strategy: &buildv1alpha1.Strategy{
				APIVersion: o.StrategyAPIVersion,
				Name:       o.StrategyName,
			},
			Dockerfile: &o.Dockerfile,
			Builder: &buildv1alpha1.Image{
				Image:       o.BuilderImage,
				Credentials: &corev1.LocalObjectReference{},
			},
			Output: buildv1alpha1.Image{
				Image:       o.OutputImage,
				Credentials: &corev1.LocalObjectReference{},
			},
			Timeout: &o.Timeout,
		},
	}

	o.Build.Spec.Env = append(o.Build.Spec.Env, util.StringSliceToEnvVarSlice(o.Envs)...)

	if len(o.SourceCredentialsSecretName) != 0 {
		o.Build.Spec.Source.Credentials = &corev1.LocalObjectReference{
			Name: o.SourceCredentialsSecretName,
		}
	}

	if len(o.StrategyKind) != 0 {
		strategyKind := buildv1alpha1.BuildStrategyKind(o.StrategyKind)
		if strategyKind != buildv1alpha1.ClusterBuildStrategyKind && strategyKind != buildv1alpha1.NamespacedBuildStrategyKind {
			return fmt.Errorf("%q is not a BuildStrategyKind", strategyKind)
		}
		o.Build.Spec.Strategy.Kind = &strategyKind
	} else {
		o.Build.Spec.Strategy.Kind = &clusterBuildStrategyKind
	}

	if len(o.BuilderCredentialsSecretName) != 0 {
		o.Build.Spec.Builder.Credentials = &corev1.LocalObjectReference{
			Name: o.BuilderCredentialsSecretName,
		}
	}

	if len(o.OutputCredentialsSecretName) != 0 {
		o.Build.Spec.Output.Credentials = &corev1.LocalObjectReference{
			Name: o.OutputCredentialsSecretName,
		}
	}

	return nil
}

// Run executes the command logic
func (o *BuildCreateOptions) Run() error {
	b, err := o.Clients.ShipwrightClientSet.ShipwrightV1alpha1().Builds(o.Clients.Namespace).Create(o.Context, o.Build, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	fmt.Printf("Created build %q\n", b.Name)

	return nil
}
