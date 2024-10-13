package flags

import (
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/pflag"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

func pointerUInt(value uint) *uint {
	return &value
}

// BuildSpecFromFlags creates a BuildSpec instance based on command-line flags.
func BuildSpecFromFlags(flags *pflag.FlagSet) *buildv1alpha1.BuildSpec {
	clusterBuildStrategyKind := buildv1alpha1.ClusterBuildStrategyKind
	bundlePruneOption := buildv1alpha1.PruneNever
	spec := &buildv1alpha1.BuildSpec{
		Source: buildv1alpha1.Source{
			Credentials:     &corev1.LocalObjectReference{},
			Revision:        pointer.String(""),
			ContextDir:      pointer.String(""),
			URL:             pointer.String(""),
			BundleContainer: &buildv1alpha1.BundleContainer{Prune: &bundlePruneOption},
		},
		Strategy: buildv1alpha1.Strategy{
			Kind:       &clusterBuildStrategyKind,
			APIVersion: &buildv1alpha1.SchemeGroupVersion.Version,
		},
		Dockerfile: pointer.String(""),
		Builder: &buildv1alpha1.Image{
			Credentials: &corev1.LocalObjectReference{},
		},
		Output: buildv1alpha1.Image{
			Credentials: &corev1.LocalObjectReference{},
			Insecure:    pointer.Bool(false),
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Timeout: &metav1.Duration{},
		Retention: &buildv1alpha1.BuildRetention{
			FailedLimit:       pointerUInt(65535),
			SucceededLimit:    pointerUInt(65535),
			TTLAfterFailed:    &metav1.Duration{},
			TTLAfterSucceeded: &metav1.Duration{},
		},
	}

	sourceFlags(flags, &spec.Source)
	strategyFlags(flags, &spec.Strategy)
	dockerfileFlags(flags, spec.Dockerfile)
	imageFlags(flags, "builder", spec.Builder)
	imageFlags(flags, "output", &spec.Output)
	timeoutFlags(flags, spec.Timeout)
	envFlags(flags, &spec.Env)
	paramValueFlag(flags, &spec.ParamValues)
	imageLabelsFlags(flags, spec.Output.Labels)
	imageAnnotationsFlags(flags, spec.Output.Annotations)
	buildRetentionFlags(flags, spec.Retention)

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
	if b.Source.ContextDir != nil && *b.Source.ContextDir == "" {
		b.Source.ContextDir = nil
	}
	if b.Source.Revision != nil && *b.Source.Revision == "" {
		b.Source.Revision = nil
	}
	if b.Source.URL != nil && *b.Source.URL == "" {
		b.Source.URL = nil
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
		if len(b.Env) == 0 {
			b.Env = nil
		}
	}
	if b.Timeout != nil && b.Timeout.Duration == 0 {
		b.Timeout = nil
	}
	if b.Dockerfile != nil && *b.Dockerfile == "" {
		b.Dockerfile = nil
	}
	if b.Output.Credentials != nil && b.Output.Credentials.Name == "" {
		b.Output.Credentials = nil
	}
	if b.Output.Insecure != nil && !*b.Output.Insecure {
		b.Output.Insecure = nil
	}
	if b.Retention != nil {
		if b.Retention.FailedLimit != nil && *b.Retention.FailedLimit == 65535 {
			b.Retention.FailedLimit = nil
		}
		if b.Retention.SucceededLimit != nil && *b.Retention.SucceededLimit == 65535 {
			b.Retention.SucceededLimit = nil
		}
		if b.Retention.TTLAfterFailed != nil && b.Retention.TTLAfterFailed.Duration == 0 {
			b.Retention.TTLAfterFailed = nil
		}
		if b.Retention.TTLAfterSucceeded != nil && b.Retention.TTLAfterSucceeded.Duration == 0 {
			b.Retention.TTLAfterSucceeded = nil
		}
		if b.Retention.FailedLimit == nil && b.Retention.SucceededLimit == nil && b.Retention.TTLAfterFailed == nil && b.Retention.TTLAfterSucceeded == nil {
			b.Retention = nil
		}
	}
}
