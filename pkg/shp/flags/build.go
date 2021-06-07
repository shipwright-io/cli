package flags

import (
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BuildSpecFlags creates a BuildSpec instance based on command-line flags.
func BuildSpecFlags(flags *pflag.FlagSet) *buildv1alpha1.BuildSpec {
	clusterBuildStrategyKind := buildv1alpha1.ClusterBuildStrategyKind
	spec := &buildv1alpha1.BuildSpec{
		Source: buildv1alpha1.Source{
			Credentials: &corev1.LocalObjectReference{},
		},
		Strategy: &buildv1alpha1.Strategy{
			Kind:       &clusterBuildStrategyKind,
			APIVersion: buildv1alpha1.SchemeGroupVersion.Version,
		},
		Dockerfile: nil,
		Builder: &buildv1alpha1.Image{
			Credentials: &corev1.LocalObjectReference{},
		},
		Output: buildv1alpha1.Image{
			Credentials: &corev1.LocalObjectReference{},
		},
		Timeout: &metav1.Duration{},
	}

	sourceFlags(flags, &spec.Source)
	strategyFlags(flags, spec.Strategy)

	flags.Var(
		NewStringPointerValue(spec.Dockerfile),
		"dockerfile",
		"path to dockerfile relative to repository",
	)

	imageFlags(flags, "builder", spec.Builder)
	imageFlags(flags, "output", &spec.Output)

	timeoutFlags(flags, spec.Timeout)

	return spec
}

// SanitizeBuildSpec checks for empty inner data structure and replaces them with nil.
func SanitizeBuildSpec(b *buildv1alpha1.BuildSpec) {
	if b == nil {
		return
	}
	if b.Source.Credentials != nil && b.Source.Credentials.Name == "" {
		b.Source.Credentials = nil
	}
	if b.Builder != nil {
		if b.Builder.Credentials != nil && b.Builder.Credentials.Name == "" {
			b.Builder.Credentials = nil
		}
		if b.Builder.Image == "" && b.Builder.Credentials == nil {
			b.Builder = nil
		}
	}
}
