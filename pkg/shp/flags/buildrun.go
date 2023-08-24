package flags

import (
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/pflag"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// BuildRunSpecFromFlags creates a BuildRun spec from command-line flags.
func BuildRunSpecFromFlags(flags *pflag.FlagSet) *buildv1alpha1.BuildRunSpec {
	spec := &buildv1alpha1.BuildRunSpec{
		BuildRef: &buildv1alpha1.BuildRef{
			APIVersion: ptr.To[string](""),
		},
		ServiceAccount: &buildv1alpha1.ServiceAccount{
			Name:     ptr.To[string](""),
			Generate: ptr.To[bool](false),
		},
		Timeout: &metav1.Duration{},
		Output: &buildv1alpha1.Image{
			Credentials: &corev1.LocalObjectReference{},
			Insecure:    ptr.To[bool](false),
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Env: []corev1.EnvVar{},
		Retention: &buildv1alpha1.BuildRunRetention{
			TTLAfterFailed:    &metav1.Duration{},
			TTLAfterSucceeded: &metav1.Duration{},
		},
	}

	buildRefFlags(flags, spec.BuildRef)
	serviceAccountFlags(flags, spec.ServiceAccount)
	timeoutFlags(flags, spec.Timeout)
	imageFlags(flags, "output", spec.Output)
	envFlags(flags, &spec.Env)
	imageLabelsFlags(flags, spec.Output.Labels)
	imageAnnotationsFlags(flags, spec.Output.Annotations)
	buildRunRetentionFlags(flags, spec.Retention)

	return spec
}

// SanitizeBuildRunSpec checks for empty inner data structures and replaces them with nil.
func SanitizeBuildRunSpec(br *buildv1alpha1.BuildRunSpec) {
	if br == nil {
		return
	}
	if br.BuildRef != nil {
		if br.BuildRef.APIVersion != nil && *br.BuildRef.APIVersion == "" {
			br.BuildRef.APIVersion = nil
		}

		if br.BuildRef.Name == "" && br.BuildRef.APIVersion == nil {
			br.BuildRef = nil
		}
	}
	if br.ServiceAccount != nil {
		if (br.ServiceAccount.Name == nil || *br.ServiceAccount.Name == "") &&
			(br.ServiceAccount.Generate == nil || !*br.ServiceAccount.Generate) {
			br.ServiceAccount = nil
		}
	}
	if br.Output != nil {
		if br.Output.Credentials != nil && br.Output.Credentials.Name == "" {
			br.Output.Credentials = nil
		}
		if br.Output.Insecure != nil && !*br.Output.Insecure {
			br.Output.Insecure = nil
		}
		if br.Output.Image == "" && br.Output.Credentials == nil {
			br.Output = nil
		}
	}
	if br.Timeout != nil && br.Timeout.Duration == 0 {
		br.Timeout = nil
	}

	if len(br.Env) == 0 {
		br.Env = nil
	}
	if br.Retention != nil {
		if br.Retention.TTLAfterFailed != nil && br.Retention.TTLAfterFailed.Duration == 0 {
			br.Retention.TTLAfterFailed = nil
		}
		if br.Retention.TTLAfterSucceeded != nil && br.Retention.TTLAfterSucceeded.Duration == 0 {
			br.Retention.TTLAfterSucceeded = nil
		}
		if br.Retention.TTLAfterFailed == nil && br.Retention.TTLAfterSucceeded == nil {
			br.Retention = nil
		}
	}
}
