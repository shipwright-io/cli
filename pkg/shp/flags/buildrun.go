package flags

import (
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/pflag"
)

// BuildRunSpecFromFlags creates a BuildRun spec from command-line flags.
func BuildRunSpecFromFlags(flags *pflag.FlagSet) *buildv1alpha1.BuildRunSpec {
	empty := ""
	spec := &buildv1alpha1.BuildRunSpec{
		BuildRef: &buildv1alpha1.BuildRef{},
		ServiceAccount: &buildv1alpha1.ServiceAccount{
			Name: &empty,
		},
		Timeout: &metav1.Duration{},
		Output: &buildv1alpha1.Image{
			Credentials: &corev1.LocalObjectReference{},
		},
	}

	buildRefFlags(flags, spec.BuildRef)
	serviceAccountFlags(flags, spec.ServiceAccount)
	timeoutFlags(flags, spec.Timeout)
	imageFlags(flags, "output", spec.Output)

	return spec
}

// SanitizeBuildRunSpec checks for empty inner data structures and replaces them with nil.
func SanitizeBuildRunSpec(br *buildv1alpha1.BuildRunSpec) {
	if br == nil {
		return
	}
	if br.ServiceAccount != nil {
		if (br.ServiceAccount.Name == nil || *br.ServiceAccount.Name == "") &&
			br.ServiceAccount.Generate == false {
			br.ServiceAccount = nil
		}
	}
	if br.Output != nil {
		if br.Output.Credentials != nil && br.Output.Credentials.Name == "" {
			br.Output.Credentials = nil
		}
		if br.Output.Image == "" && br.Output.Credentials == nil {
			br.Output = nil
		}
	}
}

// BuildRunOpts contain the end-user settings for creating buildruns
type BuildRunOpts struct {
	LocalSourceDir string
	PruneBundle    bool
	Wait           bool
	Follow         bool
}

// BuildRunOptsFromFlags sets a BuildRunOpts based on the command-line flags
func BuildRunOptsFromFlags(flags *pflag.FlagSet) *BuildRunOpts {
	opts := &BuildRunOpts{}

	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	flags.BoolVar(&opts.Wait, "wait", false, "wait until BuildRun runs to completion")
	flags.BoolVarP(&opts.Follow, "follow", "F", false, "creates buildrun and watch its log until it completes or fails (this implies wait)")
	flags.BoolVar(&opts.PruneBundle, "prune-bundle", false, "prune source code bundle after use (this implies wait)")
	flags.StringVar(&opts.LocalSourceDir, "source-directory", cwd, "directory to be used for local source code")

	return opts
}
