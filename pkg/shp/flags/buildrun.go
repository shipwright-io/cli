package flags

import (
	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	"github.com/spf13/pflag"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// BuildRunSpecFromFlags creates a BuildRun spec from command-line flags.
func BuildRunSpecFromFlags(flags *pflag.FlagSet) *buildv1beta1.BuildRunSpec {
	spec := &buildv1beta1.BuildRunSpec{
		Build: buildv1beta1.ReferencedBuild{
			Name: ptr.To(""),
		},
		ServiceAccount: ptr.To(""),
		Timeout:        &metav1.Duration{},
		Output: &buildv1beta1.Image{
			PushSecret:  ptr.To(""),
			Insecure:    ptr.To(false),
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Env: []corev1.EnvVar{},
		Retention: &buildv1beta1.BuildRunRetention{
			TTLAfterFailed:    &metav1.Duration{},
			TTLAfterSucceeded: &metav1.Duration{},
		},
		NodeSelector: map[string]string{},
	}

	buildRefFlags(flags, &spec.Build)
	serviceAccountFlags(flags, spec.ServiceAccount)
	timeoutFlags(flags, spec.Timeout)
	imageFlags(flags, "output", spec.Output)
	envFlags(flags, &spec.Env)
	paramValueFlag(flags, &spec.ParamValues)
	imageLabelsFlags(flags, spec.Output.Labels)
	imageAnnotationsFlags(flags, spec.Output.Annotations)
	buildRunRetentionFlags(flags, spec.Retention)
	buildNodeSelectorFlags(flags, spec.NodeSelector)
	return spec
}

// SanitizeBuildRunSpec checks for empty inner data structures and replaces them with nil.
func SanitizeBuildRunSpec(br *buildv1beta1.BuildRunSpec) {
	if br == nil {
		return
	}
	if br.Build.Name != nil && *br.Build.Name == "" {
		br.Build.Name = nil
	}
	if br.ServiceAccount != nil && *br.ServiceAccount == "" {
		br.ServiceAccount = nil
	}
	if br.Output != nil {
		if br.Output.PushSecret != nil && *br.Output.PushSecret == "" {
			br.Output.PushSecret = nil
		}
		if br.Output.Insecure != nil && !*br.Output.Insecure {
			br.Output.Insecure = nil
		}
		if br.Output.Image == "" && br.Output.PushSecret == nil {
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
