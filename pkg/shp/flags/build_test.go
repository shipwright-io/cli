package flags

import (
	"strconv"
	"testing"
	"time"

	o "github.com/onsi/gomega"
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

func TestBuildSpecFromFlags(t *testing.T) {
	g := o.NewWithT(t)

	credentials := corev1.LocalObjectReference{Name: "name"}
	buildStrategyKind := buildv1alpha1.ClusterBuildStrategyKind
	bundlePruneOption := buildv1alpha1.PruneNever
	expected := &buildv1alpha1.BuildSpec{
		Source: buildv1alpha1.Source{
			Credentials:     &credentials,
			URL:             pointer.String("https://some.url"),
			Revision:        pointer.String("some-rev"),
			ContextDir:      pointer.String("some-contextdir"),
			BundleContainer: &buildv1alpha1.BundleContainer{Prune: &bundlePruneOption},
		},
		Strategy: buildv1alpha1.Strategy{
			Name:       "strategy-name",
			Kind:       &buildStrategyKind,
			APIVersion: &buildv1alpha1.SchemeGroupVersion.Version,
		},
		Dockerfile: pointer.String("some-dockerfile"),
		Builder: &buildv1alpha1.Image{
			Credentials: &credentials,
			Image:       "builder-image",
		},
		Output: buildv1alpha1.Image{
			Credentials: &credentials,
			Image:       "output-image",
			Insecure:    pointer.Bool(false),
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Timeout: &metav1.Duration{
			Duration: 1 * time.Second,
		},
		Retention: &buildv1alpha1.BuildRetention{
			FailedLimit:    pointerUInt(10),
			SucceededLimit: pointerUInt(5),
			TTLAfterFailed: &metav1.Duration{
				Duration: 48 * time.Hour,
			},
			TTLAfterSucceeded: &metav1.Duration{
				Duration: 30 * time.Minute,
			},
		},
	}

	cmd := &cobra.Command{}
	flags := cmd.PersistentFlags()
	spec := BuildSpecFromFlags(flags)

	t.Run(".spec.source", func(_ *testing.T) {
		err := flags.Set(SourceURLFlag, *expected.Source.URL)
		g.Expect(err).To(o.BeNil())

		err = flags.Set(SourceRevisionFlag, *expected.Source.Revision)
		g.Expect(err).To(o.BeNil())

		err = flags.Set(SourceContextDirFlag, *expected.Source.ContextDir)
		g.Expect(err).To(o.BeNil())

		err = flags.Set(SourceCredentialsSecretFlag, expected.Source.Credentials.Name)
		g.Expect(err).To(o.BeNil())

		err = flags.Set(StrategyAPIVersionFlag, *expected.Strategy.APIVersion)
		g.Expect(err).To(o.BeNil())

		g.Expect(expected.Source).To(o.Equal(spec.Source), "spec.source")
	})

	t.Run(".spec.strategy", func(_ *testing.T) {
		err := flags.Set(StrategyKindFlag, string(buildv1alpha1.ClusterBuildStrategyKind))
		g.Expect(err).To(o.BeNil())

		err = flags.Set(StrategyNameFlag, expected.Strategy.Name)
		g.Expect(err).To(o.BeNil())

		g.Expect(expected.Strategy).To(o.Equal(spec.Strategy), "spec.strategy")
	})

	t.Run(".spec.dockerfile", func(_ *testing.T) {
		err := flags.Set(DockerfileFlag, *expected.Dockerfile)
		g.Expect(err).To(o.BeNil())

		g.Expect(spec.Dockerfile).NotTo(o.BeNil())
		g.Expect(*expected.Dockerfile).To(o.Equal(*spec.Dockerfile), "spec.dockerfile")
	})

	t.Run(".spec.builder", func(_ *testing.T) {
		err := flags.Set(BuilderImageFlag, expected.Builder.Image)
		g.Expect(err).To(o.BeNil())

		err = flags.Set(BuilderCredentialsSecretFlag, expected.Builder.Credentials.Name)
		g.Expect(err).To(o.BeNil())

		g.Expect(*expected.Builder).To(o.Equal(*spec.Builder), "spec.builder")
	})

	t.Run(".spec.output", func(_ *testing.T) {
		err := flags.Set(OutputImageFlag, expected.Output.Image)
		g.Expect(err).To(o.BeNil())

		err = flags.Set(OutputCredentialsSecretFlag, expected.Output.Credentials.Name)
		g.Expect(err).To(o.BeNil())

		err = flags.Set(OutputInsecureFlag, strconv.FormatBool(*expected.Output.Insecure))
		g.Expect(err).To(o.BeNil())

		g.Expect(expected.Output).To(o.Equal(spec.Output), "spec.output")
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

	completeBuildSpec := buildv1alpha1.BuildSpec{
		Source: buildv1alpha1.Source{
			Credentials: &corev1.LocalObjectReference{Name: "name"},
		},
		Builder: &buildv1alpha1.Image{
			Credentials: &corev1.LocalObjectReference{Name: "name"},
			Image:       "image",
		},
	}

	emptyString := ""

	testCases := []struct {
		name string
		in   buildv1alpha1.BuildSpec
		out  buildv1alpha1.BuildSpec
	}{{
		name: "all empty should stay empty",
		in:   buildv1alpha1.BuildSpec{},
		out:  buildv1alpha1.BuildSpec{},
	}, {
		name: "should clean-up `.spec.source.credentials`",
		in: buildv1alpha1.BuildSpec{Source: buildv1alpha1.Source{
			Credentials: &corev1.LocalObjectReference{},
		}},
		out: buildv1alpha1.BuildSpec{},
	}, {
		name: "should clean-up `.spec.builder.credentials`",
		in: buildv1alpha1.BuildSpec{Builder: &buildv1alpha1.Image{
			Credentials: &corev1.LocalObjectReference{},
		}},
		out: buildv1alpha1.BuildSpec{},
	}, {
		name: "should clean-up `.spec.builder.image`",
		in:   buildv1alpha1.BuildSpec{Builder: &buildv1alpha1.Image{}},
		out:  buildv1alpha1.BuildSpec{},
	}, {
		name: "should not clean-up complete objects",
		in:   completeBuildSpec,
		out:  completeBuildSpec,
	}, {
		name: "should clean-up 0s duration",
		in: buildv1alpha1.BuildSpec{Timeout: &metav1.Duration{
			Duration: time.Duration(0),
		}},
		out: buildv1alpha1.BuildSpec{Timeout: nil},
	}, {
		name: "should clean-up an empty Dockerfile",
		in:   buildv1alpha1.BuildSpec{Dockerfile: &emptyString},
		out:  buildv1alpha1.BuildSpec{Dockerfile: nil},
	}, {
		name: "should clean-up an empty revision",
		in: buildv1alpha1.BuildSpec{Source: buildv1alpha1.Source{
			Revision: &emptyString,
		}},
		out: buildv1alpha1.BuildSpec{Source: buildv1alpha1.Source{
			Revision: nil,
		}},
	}, {
		name: "should clean-up an empty retention",
		in: buildv1alpha1.BuildSpec{
			Retention: &buildv1alpha1.BuildRetention{},
		},
		out: buildv1alpha1.BuildSpec{},
	}, {
		name: "should clean-up an empty source contextDir",
		in: buildv1alpha1.BuildSpec{
			Source: buildv1alpha1.Source{
				ContextDir: pointer.String(""),
			},
		},
		out: buildv1alpha1.BuildSpec{
			Source: buildv1alpha1.Source{},
		},
	}, {
		name: "should clean-up an empty source URL",
		in: buildv1alpha1.BuildSpec{
			Source: buildv1alpha1.Source{
				URL: pointer.String(""),
			},
		},
		out: buildv1alpha1.BuildSpec{
			Source: buildv1alpha1.Source{},
		},
	}, {
		name: "should clean-up a false output insecure",
		in: buildv1alpha1.BuildSpec{
			Output: buildv1alpha1.Image{
				Image:    "some",
				Insecure: pointer.Bool(false),
			},
		},
		out: buildv1alpha1.BuildSpec{
			Output: buildv1alpha1.Image{
				Image: "some",
			},
		},
	}, {
		name: "should not clean-up a true output insecure",
		in: buildv1alpha1.BuildSpec{
			Output: buildv1alpha1.Image{
				Image:    "some",
				Insecure: pointer.Bool(true),
			},
		},
		out: buildv1alpha1.BuildSpec{
			Output: buildv1alpha1.Image{
				Image:    "some",
				Insecure: pointer.Bool(true),
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
