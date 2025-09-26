package flags

import (
	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	"github.com/spf13/pflag"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func pointerUInt(value uint) *uint {
	return &value
}

// BuildSpecFromFlags creates a BuildSpec instance based on command-line flags.
func BuildSpecFromFlags(flags *pflag.FlagSet) (*buildv1beta1.BuildSpec, *string, *string) {
	clusterBuildStrategyKind := buildv1beta1.ClusterBuildStrategyKind
	pruneOption := buildv1beta1.PruneNever
	spec := &buildv1beta1.BuildSpec{
		Source: &buildv1beta1.Source{
			ContextDir: new(string),
			Git: &buildv1beta1.Git{
				Revision:    new(string),
				CloneSecret: new(string),
			},
			OCIArtifact: &buildv1beta1.OCIArtifact{
				Prune:      &pruneOption,
				PullSecret: new(string),
			},
		},
		Strategy: buildv1beta1.Strategy{
			Kind: &clusterBuildStrategyKind,
		},
		Output: buildv1beta1.Image{
			Insecure:    ptr.To(false),
			PushSecret:  new(string),
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Timeout: &metav1.Duration{},
		Retention: &buildv1beta1.BuildRetention{
			FailedLimit:       pointerUInt(65535),
			SucceededLimit:    pointerUInt(65535),
			TTLAfterFailed:    &metav1.Duration{},
			TTLAfterSucceeded: &metav1.Duration{},
		},
		NodeSelector:  map[string]string{},
		SchedulerName: new(string),
	}

	sourceFlags(flags, spec.Source)
	strategyFlags(flags, &spec.Strategy)
	imageFlags(flags, "output", &spec.Output)
	timeoutFlags(flags, spec.Timeout)
	envFlags(flags, &spec.Env)
	paramValueFlag(flags, &spec.ParamValues)
	imageLabelsFlags(flags, spec.Output.Labels)
	imageAnnotationsFlags(flags, spec.Output.Annotations)
	buildRetentionFlags(flags, spec.Retention)
	buildNodeSelectorFlags(flags, spec.NodeSelector)
	buildSchedulerNameFlag(flags, spec.SchedulerName)
	var dockerfile, builderImage string
	dockerfileFlags(flags, &dockerfile)
	builderImageFlag(flags, &builderImage)

	return spec, &dockerfile, &builderImage
}

// SanitizeBuildSpec checks for empty inner data structures and replaces them with nil.
func SanitizeBuildSpec(b *buildv1beta1.BuildSpec) {
	if b == nil {
		return
	}

	if b.Source != nil {
		if b.Source.Git != nil {
			if b.Source.Git.URL == "" {
				b.Source.Git = nil
			}
		}

		if b.Source.Git != nil {
			if b.Source.Git.Revision != nil && *b.Source.Git.Revision == "" {
				b.Source.Git.Revision = nil
			}
			if b.Source.Git.CloneSecret != nil && *b.Source.Git.CloneSecret == "" {
				b.Source.Git.CloneSecret = nil
			}
		}

		if b.Source.ContextDir != nil && *b.Source.ContextDir == "" {
			b.Source.ContextDir = nil
		}
		if b.Source.OCIArtifact != nil && b.Source.OCIArtifact.Image == "" {
			b.Source.OCIArtifact = nil
		}
		if b.Source.OCIArtifact != nil && b.Source.OCIArtifact.PullSecret != nil {
			if *b.Source.OCIArtifact.PullSecret == "" {
				b.Source.OCIArtifact.PullSecret = nil
			}
		}

		if b.Source.Git == nil && b.Source.OCIArtifact == nil {
			b.Source = nil
		}
	}

	if len(b.Env) == 0 {
		b.Env = nil
	}

	if b.Timeout != nil && b.Timeout.Duration == 0 {
		b.Timeout = nil
	}

	if b.Output.PushSecret != nil && *b.Output.PushSecret == "" {
		b.Output.PushSecret = nil
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
	if b.SchedulerName != nil && *b.SchedulerName == "" {
		b.SchedulerName = nil
	}
}
