package flags

import (
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/pflag"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	if br.Timeout != nil && br.Timeout.Duration == 0 {
		br.Timeout = nil
	}
}
