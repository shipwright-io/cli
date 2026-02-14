package flags

import (
	"strconv"
	"testing"
	"time"

	o "github.com/onsi/gomega"
	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func TestBuildSpecFromFlags(t *testing.T) {
	g := o.NewWithT(t)

	buildStrategyKind := buildv1beta1.ClusterBuildStrategyKind
	bundlePruneOption := buildv1beta1.PruneNever
	expected := &buildv1beta1.BuildSpec{
		Source: &buildv1beta1.Source{
			Git: &buildv1beta1.Git{
				URL:         "https://some.url",
				Revision:    ptr.To("some-rev"),
				CloneSecret: ptr.To("name"),
			},
			ContextDir: ptr.To("some-contextdir"),
			OCIArtifact: &buildv1beta1.OCIArtifact{
				Prune:      &bundlePruneOption,
				PullSecret: ptr.To("pull-secret"),
			},
		},
		Strategy: buildv1beta1.Strategy{
			Name: "strategy-name",
			Kind: &buildStrategyKind,
		},

		Output: buildv1beta1.Image{
			PushSecret:  ptr.To("name"),
			Image:       "output-image",
			Insecure:    ptr.To(false),
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Timeout: &metav1.Duration{
			Duration: 1 * time.Second,
		},
		Retention: &buildv1beta1.BuildRetention{
			FailedLimit:    pointerUInt(10),
			SucceededLimit: pointerUInt(5),
			TTLAfterFailed: &metav1.Duration{
				Duration: 48 * time.Hour,
			},
			TTLAfterSucceeded: &metav1.Duration{
				Duration: 30 * time.Minute,
			},
		},
		NodeSelector:  map[string]string{"kubernetes.io/hostname": "worker-1"},
		SchedulerName: ptr.To("dolphinscheduler"),
	}

	cmd := &cobra.Command{}
	flags := cmd.PersistentFlags()
	spec, dockerfile, builderImage := BuildSpecFromFlags(flags)

	t.Run(".spec.source", func(_ *testing.T) {
		err := flags.Set(SourceGitURLFlag, expected.Source.Git.URL)
		g.Expect(err).To(o.BeNil())

		err = flags.Set(SourceGitRevisionFlag, *expected.Source.Git.Revision)
		g.Expect(err).To(o.BeNil())

		err = flags.Set(SourceContextDirFlag, *expected.Source.ContextDir)
		g.Expect(err).To(o.BeNil())

		err = flags.Set(SourceGitCloneSecretFlag, *expected.Source.Git.CloneSecret)
		g.Expect(err).To(o.BeNil())

		err = flags.Set(SourceOCIArtifactPullSecretFlag, *expected.Source.OCIArtifact.PullSecret)
		g.Expect(err).To(o.BeNil())

		g.Expect(expected.Source).To(o.Equal(spec.Source), "spec.source")
	})

	t.Run(".spec.strategy", func(_ *testing.T) {
		err := flags.Set(StrategyKindFlag, string(buildv1beta1.ClusterBuildStrategyKind))
		g.Expect(err).To(o.BeNil())

		err = flags.Set(StrategyNameFlag, expected.Strategy.Name)
		g.Expect(err).To(o.BeNil())

		g.Expect(expected.Strategy).To(o.Equal(spec.Strategy), "spec.strategy")
	})

	t.Run("dockerfile", func(_ *testing.T) {
		err := flags.Set(DockerfileFlag, "test-dockerfile")
		g.Expect(err).To(o.BeNil())
		g.Expect(*dockerfile).To(o.Equal("test-dockerfile"))
	})

	t.Run("builderImage", func(_ *testing.T) {
		err := flags.Set(BuilderImageFlag, "test-builder-image")
		g.Expect(err).To(o.BeNil())
		g.Expect(*builderImage).To(o.Equal("test-builder-image"))
	})

	t.Run(".spec.output", func(_ *testing.T) {
		err := flags.Set(OutputImageFlag, expected.Output.Image)
		g.Expect(err).To(o.BeNil())

		err = flags.Set(OutputCredentialsSecretFlag, *expected.Output.PushSecret)
		g.Expect(err).To(o.BeNil())

		err = flags.Set(OutputInsecureFlag, strconv.FormatBool(*expected.Output.Insecure))
		g.Expect(err).To(o.BeNil())

		g.Expect(expected.Output).To(o.Equal(spec.Output), "spec.output")
	})

	t.Run(".spec.nodeSelector", func(_ *testing.T) {
		err := flags.Set(NodeSelectorFlag, "kubernetes.io/hostname=worker-1")
		g.Expect(err).To(o.BeNil())
		// g.Expect(expected.NodeSelector).To(o.HaveKeyWithValue("kubernetes.io/hostname",spec.NodeSelector["kubernetes.io/hostname"]), ".spec.nodeSelector")
		g.Expect(expected.NodeSelector).To(o.Equal(spec.NodeSelector), ".spec.nodeSelector")
	})

	t.Run(".spec.schedulerName", func(_ *testing.T) {
		err := flags.Set(SchedulerNameFlag, *expected.SchedulerName)
		g.Expect(err).To(o.BeNil())

		g.Expect(expected.SchedulerName).To(o.Equal(spec.SchedulerName), "spec.schedulerName")
	})

	t.Run(".spec.timeout", func(_ *testing.T) {
		err := flags.Set(TimeoutFlag, expected.Timeout.Duration.String())
		g.Expect(err).To(o.BeNil())

		g.Expect(*expected.Timeout).To(o.Equal(*spec.Timeout), "spec.timeout")
	})

	t.Run(".spec.retention.failedLimit", func(_ *testing.T) {
		err := flags.Set(RetentionFailedLimitFlag, strconv.FormatUint(uint64(*expected.Retention.FailedLimit), 10))
		g.Expect(err).To(o.BeNil())

		g.Expect(*expected.Retention.FailedLimit).To(o.Equal(*spec.Retention.FailedLimit), "spec.retention.failedLimit")
	})

	t.Run(".spec.retention.succeededLimit", func(_ *testing.T) {
		err := flags.Set(RetentionSucceededLimitFlag, strconv.FormatUint(uint64(*expected.Retention.SucceededLimit), 10))
		g.Expect(err).To(o.BeNil())

		g.Expect(*expected.Retention.SucceededLimit).To(o.Equal(*spec.Retention.SucceededLimit), "spec.retention.succeededLimit")
	})

	t.Run(".spec.retention.ttlAfterFailed", func(_ *testing.T) {
		err := flags.Set(RetentionTTLAfterFailedFlag, expected.Retention.TTLAfterFailed.Duration.String())
		g.Expect(err).To(o.BeNil())

		g.Expect(*expected.Retention.TTLAfterFailed).To(o.Equal(*spec.Retention.TTLAfterFailed), "spec.retention.ttlAfterFailed")
	})

	t.Run(".spec.retention.ttlAfterSucceeded", func(_ *testing.T) {
		err := flags.Set(RetentionTTLAfterSucceededFlag, expected.Retention.TTLAfterSucceeded.Duration.String())
		g.Expect(err).To(o.BeNil())

		g.Expect(*expected.Retention.TTLAfterSucceeded).To(o.Equal(*spec.Retention.TTLAfterSucceeded), "spec.retention.ttlAfterSucceeded")
	})
}

func TestSanitizeBuildSpec(t *testing.T) {
	g := o.NewWithT(t)

	completeBuildSpec := buildv1beta1.BuildSpec{
		Source: &buildv1beta1.Source{
			Type: buildv1beta1.GitType,
			Git: &buildv1beta1.Git{
				URL:         "test-url",
				CloneSecret: ptr.To("name"),
			},
		},
	}

	emptyString := ""

	testCases := []struct {
		name string
		in   buildv1beta1.BuildSpec
		out  buildv1beta1.BuildSpec
	}{
		{
			name: "all empty should stay empty",
			in:   buildv1beta1.BuildSpec{},
			out:  buildv1beta1.BuildSpec{},
		}, {
			name: "should clean-up `.spec.source.credentials`",
			in: buildv1beta1.BuildSpec{Source: &buildv1beta1.Source{
				Git: &buildv1beta1.Git{
					CloneSecret: new(string),
				},
			}},
			out: buildv1beta1.BuildSpec{},
		},
		{
			name: "should not clean-up complete objects",
			in:   completeBuildSpec,
			out:  completeBuildSpec,
		}, {
			name: "should clean-up 0s duration",
			in: buildv1beta1.BuildSpec{Timeout: &metav1.Duration{
				Duration: time.Duration(0),
			}},
			out: buildv1beta1.BuildSpec{Timeout: nil},
		},
		{
			name: "should clean-up an empty revision",
			in: buildv1beta1.BuildSpec{Source: &buildv1beta1.Source{
				Type: buildv1beta1.GitType,
				Git: &buildv1beta1.Git{
					URL:      "test-url",
					Revision: &emptyString,
				},
			}},
			out: buildv1beta1.BuildSpec{Source: &buildv1beta1.Source{
				Type: buildv1beta1.GitType,
				Git: &buildv1beta1.Git{
					URL:      "test-url",
					Revision: nil,
				},
			}},
		}, {
			name: "should clean-up an empty retention",
			in: buildv1beta1.BuildSpec{
				Retention: &buildv1beta1.BuildRetention{},
			},
			out: buildv1beta1.BuildSpec{},
		}, {
			name: "should clean-up an empty source contextDir",
			in: buildv1beta1.BuildSpec{
				Source: &buildv1beta1.Source{
					ContextDir: new(string),
				},
			},
			out: buildv1beta1.BuildSpec{},
		}, {
			name: "should clean-up an empty source URL",
			in: buildv1beta1.BuildSpec{
				Source: &buildv1beta1.Source{
					Git: &buildv1beta1.Git{
						URL: "",
					},
				},
			},
			out: buildv1beta1.BuildSpec{},
		}, {
			name: "should clean-up a false output insecure",
			in: buildv1beta1.BuildSpec{
				Output: buildv1beta1.Image{
					Image:    "some",
					Insecure: ptr.To(false),
				},
			},
			out: buildv1beta1.BuildSpec{
				Output: buildv1beta1.Image{
					Image: "some",
				},
			},
		}, {
			name: "should not clean-up a true output insecure",
			in: buildv1beta1.BuildSpec{
				Output: buildv1beta1.Image{
					Image:    "some",
					Insecure: ptr.To(true),
				},
			},
			out: buildv1beta1.BuildSpec{
				Output: buildv1beta1.Image{
					Image:    "some",
					Insecure: ptr.To(true),
				},
			},
		}}

	for _, tt := range testCases {
		t.Run(tt.name, func(_ *testing.T) {
			aCopy := tt.in.DeepCopy()
			SanitizeBuildSpec(aCopy)
			g.Expect(tt.out).To(o.Equal(*aCopy))
		})
	}
}
