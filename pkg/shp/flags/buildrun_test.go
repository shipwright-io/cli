package flags

import (
	"fmt"
	"testing"
	"time"

	"github.com/onsi/gomega"
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	o "github.com/onsi/gomega"
)

func TestBuildRunSpecFromFlags(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	str := "something-random"
	expected := &buildv1alpha1.BuildRunSpec{
		BuildRef: &buildv1alpha1.BuildRef{
			Name: str,
		},
		ServiceAccount: &buildv1alpha1.ServiceAccount{
			Name:     &str,
			Generate: false,
		},
		Timeout: &metav1.Duration{
			Duration: 1 * time.Second,
		},
		Output: &buildv1alpha1.Image{
			Credentials: &corev1.LocalObjectReference{Name: "name"},
			Image:       str,
		},
	}

	cmd := &cobra.Command{}
	flags := cmd.PersistentFlags()
	spec := BuildRunSpecFromFlags(flags)

	t.Run(".spec.buildRef", func(t *testing.T) {
		err := flags.Set(BuildrefNameFlag, expected.BuildRef.Name)
		g.Expect(err).To(o.BeNil())

		g.Expect(*expected.BuildRef).To(o.Equal(*spec.BuildRef), "spec.buildRef")
	})

	t.Run(".spec.serviceAccount", func(t *testing.T) {
		err := flags.Set(ServiceAccountNameFlag, *expected.ServiceAccount.Name)
		g.Expect(err).To(o.BeNil())

		generate := fmt.Sprintf("%v", expected.ServiceAccount.Generate)
		err = flags.Set(ServiceAccountGenerateFlag, generate)
		g.Expect(err).To(o.BeNil())

		g.Expect(*expected.ServiceAccount).To(o.Equal(*spec.ServiceAccount), "spec.serviceAccount")
	})

	t.Run(".spec.timeout", func(t *testing.T) {
		err := flags.Set(TimeoutFlag, expected.Timeout.Duration.String())
		g.Expect(err).To(o.BeNil())

		g.Expect(*expected.Timeout).To(o.Equal(*spec.Timeout), "spec.timeout")
	})

	t.Run(".spec.output", func(t *testing.T) {
		err := flags.Set(OutputImageFlag, expected.Output.Image)
		g.Expect(err).To(o.BeNil())

		err = flags.Set(OutputCredentialsSecretFlag, expected.Output.Credentials.Name)
		g.Expect(err).To(o.BeNil())

		g.Expect(*expected.Output).To(o.Equal(*spec.Output), "spec.output")
	})
}

func TestSanitizeBuildRunSpec(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	name := "name"
	completeBuildRunSpec := buildv1alpha1.BuildRunSpec{
		ServiceAccount: &buildv1alpha1.ServiceAccount{Name: &name},
		Output: &buildv1alpha1.Image{
			Credentials: &corev1.LocalObjectReference{Name: "name"},
			Image:       "image",
		},
	}

	testCases := []struct {
		name string
		in   buildv1alpha1.BuildRunSpec
		out  buildv1alpha1.BuildRunSpec
	}{{
		name: "all empty should stay empty",
		in:   buildv1alpha1.BuildRunSpec{},
		out:  buildv1alpha1.BuildRunSpec{},
	}, {
		name: "should clean-up service-account",
		in:   buildv1alpha1.BuildRunSpec{ServiceAccount: &buildv1alpha1.ServiceAccount{}},
		out:  buildv1alpha1.BuildRunSpec{},
	}, {
		name: "should clean-up output",
		in:   buildv1alpha1.BuildRunSpec{Output: &buildv1alpha1.Image{}},
		out:  buildv1alpha1.BuildRunSpec{},
	}, {
		name: "should not clean-up complete objects",
		in:   completeBuildRunSpec,
		out:  completeBuildRunSpec,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			copy := tt.in.DeepCopy()
			SanitizeBuildRunSpec(copy)
			g.Expect(tt.out).To(o.Equal(*copy))
		})
	}
}
