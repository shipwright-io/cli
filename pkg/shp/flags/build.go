package flags

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/spf13/pflag"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
)

// BuildSpecFromFlags creates a BuildSpec instance based on command-line flags.
func BuildSpecFromFlags(flags *pflag.FlagSet) *buildv1alpha1.BuildSpec {
	clusterBuildStrategyKind := buildv1alpha1.ClusterBuildStrategyKind
	spec := &buildv1alpha1.BuildSpec{
		Source: buildv1alpha1.Source{
			Credentials:     &corev1.LocalObjectReference{},
			Revision:        pointer.String(""),
			ContextDir:      pointer.String(""),
			BundleContainer: &buildv1alpha1.BundleContainer{},
		},
		Strategy: &buildv1alpha1.Strategy{
			Kind:       &clusterBuildStrategyKind,
			APIVersion: buildv1alpha1.SchemeGroupVersion.Version,
		},
		Dockerfile: pointer.String(""),
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
	dockerfileFlags(flags, spec.Dockerfile)
	imageFlags(flags, "builder", spec.Builder)
	imageFlags(flags, "output", &spec.Output)
	timeoutFlags(flags, spec.Timeout)

	return spec
}

// SanitizeBuildSpec checks for empty inner data structures and replaces them with nil.
func SanitizeBuildSpec(b *buildv1alpha1.BuildSpec) {
	if b == nil {
		return
	}
	if b.Source.Credentials != nil && b.Source.Credentials.Name == "" {
		b.Source.Credentials = nil
	}

	if b.Source.BundleContainer != nil && b.Source.BundleContainer.Image == "" {
		b.Source.BundleContainer = nil
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
