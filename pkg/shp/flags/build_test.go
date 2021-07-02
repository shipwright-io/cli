package flags

import (
	"testing"
	"time"

	"github.com/onsi/gomega"
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	o "github.com/onsi/gomega"
)

func TestBuildSpecFromFlags(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	str := "something-random"
	credentials := corev1.LocalObjectReference{Name: "name"}
	buildStrategyKind := buildv1alpha1.ClusterBuildStrategyKind
	expected := &buildv1alpha1.BuildSpec{
		Source: buildv1alpha1.Source{
			Credentials: &credentials,
			URL:         str,
			Revision:    &str,
			ContextDir:  &str,
		},
		Strategy: &buildv1alpha1.Strategy{
			Name:       str,
			Kind:       &buildStrategyKind,
			APIVersion: buildv1alpha1.SchemeGroupVersion.Version,
		},
		Dockerfile: &str,
		Builder: &buildv1alpha1.Image{
			Credentials: &credentials,
			Image:       str,
		},
		Output: buildv1alpha1.Image{
			Credentials: &credentials,
			Image:       str,
		},
		Timeout: &metav1.Duration{
			Duration: 1 * time.Second,
		},
	}

	cmd := &cobra.Command{}
	flags := cmd.PersistentFlags()
	spec := BuildSpecFromFlags(flags)

	t.Run(".spec.source", func(t *testing.T) {
		err := flags.Set(SourceURLFlag, expected.Source.URL)
		g.Expect(err).To(o.BeNil())

		err = flags.Set(SourceRevisionFlag, *expected.Source.Revision)
		g.Expect(err).To(o.BeNil())

		err = flags.Set(SourceCredentialsSecretFlag, expected.Source.Credentials.Name)
		g.Expect(err).To(o.BeNil())

		err = flags.Set(StrategyAPIVersionFlag, expected.Strategy.APIVersion)
		g.Expect(err).To(o.BeNil())

		g.Expect(expected.Source).To(o.Equal(spec.Source), "spec.source")
	})

	t.Run(".spec.strategy", func(t *testing.T) {
		err := flags.Set(StrategyKindFlag, string(buildv1alpha1.ClusterBuildStrategyKind))
		g.Expect(err).To(o.BeNil())

		err = flags.Set(StrategyNameFlag, expected.Strategy.Name)
		g.Expect(err).To(o.BeNil())

		g.Expect(expected.Strategy).To(o.Equal(spec.Strategy), "spec.strategy")
	})

	t.Run(".spec.dockerfile", func(t *testing.T) {
		err := flags.Set(DockerfileFlag, *expected.Dockerfile)
		g.Expect(err).To(o.BeNil())

		g.Expect(spec.Dockerfile).NotTo(o.BeNil())
		g.Expect(*expected.Dockerfile).To(o.Equal(*spec.Dockerfile), "spec.dockerfile")
	})

	t.Run(".spec.builder", func(t *testing.T) {
		err := flags.Set(BuilderImageFlag, expected.Builder.Image)
		g.Expect(err).To(o.BeNil())

		err = flags.Set(BuilderCredentialsSecretFlag, expected.Builder.Credentials.Name)
		g.Expect(err).To(o.BeNil())

		g.Expect(*expected.Builder).To(o.Equal(*spec.Builder), "spec.builder")
	})

	t.Run(".spec.output", func(t *testing.T) {
		err := flags.Set(OutputImageFlag, expected.Output.Image)
		g.Expect(err).To(o.BeNil())

		err = flags.Set(OutputCredentialsSecretFlag, expected.Output.Credentials.Name)
		g.Expect(err).To(o.BeNil())

		g.Expect(expected.Output).To(o.Equal(spec.Output), "spec.output")
	})

	t.Run(".spec.timeout", func(t *testing.T) {
		err := flags.Set(TimeoutFlag, expected.Timeout.Duration.String())
		g.Expect(err).To(o.BeNil())

		g.Expect(*expected.Timeout).To(o.Equal(*spec.Timeout), "spec.timeout")
	})
}

func TestSanitizeBuildSpec(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	completeBuildSpec := buildv1alpha1.BuildSpec{
		Source: buildv1alpha1.Source{
			Credentials: &corev1.LocalObjectReference{Name: "name"},
		},
		Builder: &buildv1alpha1.Image{
			Credentials: &corev1.LocalObjectReference{Name: "name"},
			Image:       "image",
		},
	}

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
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			copy := tt.in.DeepCopy()
			SanitizeBuildSpec(copy)
			g.Expect(tt.out).To(o.Equal(*copy))
		})
	}
}
