package buildrun

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
	// Long description for the "buildrun create" command
	buildRunCreateLongDescription = templates.LongDesc(`
		Creates a BuildRun
	`)

	// Examples for using the "buildrun create" command
	buildRunCreateExamples = templates.Examples(`
		$ shp buildrun create my-buildrun --buildref-name=my-build
	`)
)

// BuildRunCreateOptions stores data passed to the command via command line flags
type BuildRunCreateOptions struct {
	types.SharedOptions

	BuildRun *buildv1alpha1.BuildRun

	Name string

	BuildRefName       string
	BuildRefAPIVersion string

	ServiceAccountName     string
	ServiceAccountGenerate bool

	OutputImage                 string
	OutputCredentialsSecretName string

	Envs []string

	Timeout metav1.Duration
}

// newBuildRunCreateCmd creates the "buildrun create" command
func newBuildRunCreateCmd(ctx context.Context, ioStreams *genericclioptions.IOStreams, clients *types.ClientSets, o *BuildRunCreateOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create <name> [flags]",
		Short:   "Creates a BuildRun instance.",
		Long:    buildRunCreateLongDescription,
		Args:    cobra.ExactArgs(1),
		Example: buildRunCreateExamples,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(args))
			cmdutil.CheckErr(o.Run())
		},
	}

	cmd.Flags().StringVar(&o.BuildRefName, "buildref-name", "", "name of build resource to reference")
	cmd.MarkFlagRequired("buildref-name")
	cmd.Flags().StringVar(&o.BuildRefAPIVersion, "buildref-apiversion", "", "API version of build resource to reference")

	cmd.Flags().StringVar(&o.ServiceAccountName, "sa-name", "", "Kubernetes service-account name")
	cmd.Flags().BoolVar(&o.ServiceAccountGenerate, "sa-generate", false, "generate a Kubernetes service-account for the build")

	cmd.Flags().StringVar(&o.OutputImage, "output-image", "", "The location to push the output image to.")
	cmd.Flags().StringVar(&o.OutputCredentialsSecretName, "output-credentials-secret", "", "The name of the Secret that contains credentials for the repository to push the built Image to.")

	cmd.Flags().StringArrayVarP(&o.Envs, "env", "e", []string{}, "specify a key-value pair for an environment variable to set for the build container")

	cmd.Flags().DurationVar(&o.Timeout.Duration, "timeout", time.Duration(0), "How long to let the build run before timing out.")

	return cmd
}

// NewBuildRunCreateCmd is a wrapper for newBuildRunCreateCmd
func NewBuildRunCreateCmd(ctx context.Context, ioStreams *genericclioptions.IOStreams, clients *types.ClientSets) *cobra.Command {
	o := &BuildRunCreateOptions{
		SharedOptions: types.SharedOptions{
			Clients: clients,
			Context: ctx,
			Streams: ioStreams,
		},
	}

	return newBuildRunCreateCmd(ctx, ioStreams, clients, o)
}

// Complete processes any data that is needed before Run executes
func (o *BuildRunCreateOptions) Complete(args []string) error {
	// Guard against index out of bound errors
	if len(args) == 0 {
		return errors.New("argument list is empty")
	}

	o.Name = args[0]

	o.BuildRun = &buildv1alpha1.BuildRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: o.Name,
		},
		Spec: buildv1alpha1.BuildRunSpec{
			BuildRef: &buildv1alpha1.BuildRef{
				Name:       o.BuildRefName,
				APIVersion: o.BuildRefAPIVersion,
			},
			ServiceAccount: &buildv1alpha1.ServiceAccount{},
			Timeout:        &o.Timeout,
			Output: &buildv1alpha1.Image{
				Image:       o.OutputImage,
				Credentials: &corev1.LocalObjectReference{},
			},
		},
	}

	o.BuildRun.Spec.Env = append(o.BuildRun.Spec.Env, util.StringSliceToEnvVarSlice(o.Envs)...)

	if len(o.OutputCredentialsSecretName) != 0 {
		o.BuildRun.Spec.Output.Credentials = &corev1.LocalObjectReference{
			Name: o.OutputCredentialsSecretName,
		}
	}

	if len(o.ServiceAccountName) != 0 {
		o.BuildRun.Spec.ServiceAccount.Name = &o.ServiceAccountName
		o.BuildRun.Spec.ServiceAccount.Generate = false
	} else {
		o.BuildRun.Spec.ServiceAccount.Generate = true
	}

	return nil
}

// Run executes the command logic
func (o *BuildRunCreateOptions) Run() error {
	if _, err := o.Clients.ShipwrightClientSet.ShipwrightV1alpha1().BuildRuns(o.Clients.Namespace).Create(o.Context, o.BuildRun, metav1.CreateOptions{}); err != nil {
		return err
	}

	fmt.Fprintf(o.Streams.Out, "BuildRun %q created for Build %q\n", o.Name, o.BuildRun.Spec.BuildRef.Name)

	return nil
}
