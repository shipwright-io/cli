package flags

import (
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

func TestBuildRunSpecFromFlags(t *testing.T) {
	g := NewWithT(t)

	str := "something-random"
	expected := &buildv1alpha1.BuildRunSpec{
		BuildRef: &buildv1alpha1.BuildRef{
			Name:       str,
			APIVersion: pointer.String(""),
		},
		ServiceAccount: &buildv1alpha1.ServiceAccount{
			Name:     &str,
			Generate: pointer.Bool(false),
		},
		Timeout: &metav1.Duration{
			Duration: 1 * time.Second,
		},
		Output: &buildv1alpha1.Image{
			Credentials: &corev1.LocalObjectReference{Name: "name"},
			Image:       str,
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Retention: &buildv1alpha1.BuildRunRetention{
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
	spec := BuildRunSpecFromFlags(flags)

	t.Run(".spec.buildRef", func(t *testing.T) {
		err := flags.Set(BuildrefNameFlag, expected.BuildRef.Name)
		g.Expect(err).To(BeNil())

		g.Expect(*expected.BuildRef).To(Equal(*spec.BuildRef), "spec.buildRef")
	})

	t.Run(".spec.serviceAccount", func(t *testing.T) {
		err := flags.Set(ServiceAccountNameFlag, *expected.ServiceAccount.Name)
		g.Expect(err).To(BeNil())

		err = flags.Set(ServiceAccountGenerateFlag, strconv.FormatBool(*expected.ServiceAccount.Generate))
		g.Expect(err).To(BeNil())

		g.Expect(*expected.ServiceAccount).To(Equal(*spec.ServiceAccount), "spec.serviceAccount")
	})

	t.Run(".spec.timeout", func(t *testing.T) {
		err := flags.Set(TimeoutFlag, expected.Timeout.Duration.String())
		g.Expect(err).To(BeNil())

		g.Expect(*expected.Timeout).To(Equal(*spec.Timeout), "spec.timeout")
	})

	t.Run(".spec.output", func(t *testing.T) {
		err := flags.Set(OutputImageFlag, expected.Output.Image)
		g.Expect(err).To(BeNil())

		err = flags.Set(OutputCredentialsSecretFlag, expected.Output.Credentials.Name)
		g.Expect(err).To(BeNil())

		g.Expect(*expected.Output).To(Equal(*spec.Output), "spec.output")
	})

	t.Run(".spec.retention.ttlAfterFailed", func(t *testing.T) {
		err := flags.Set(RetentionTTLAfterFailedFlag, expected.Retention.TTLAfterFailed.Duration.String())
		g.Expect(err).To(BeNil())

		g.Expect(*expected.Retention.TTLAfterFailed).To(Equal(*spec.Retention.TTLAfterFailed), "spec.retention.ttlAfterFailed")
	})

	t.Run(".spec.retention.ttlAfterSucceeded", func(t *testing.T) {
		err := flags.Set(RetentionTTLAfterSucceededFlag, expected.Retention.TTLAfterSucceeded.Duration.String())
		g.Expect(err).To(BeNil())

		g.Expect(*expected.Retention.TTLAfterSucceeded).To(Equal(*spec.Retention.TTLAfterSucceeded), "spec.retention.ttlAfterSucceeded")
	})
}

func TestSanitizeBuildRunSpec(t *testing.T) {
	g := NewWithT(t)

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
	}, {
		name: "should clean-up 0s duration",
		in: buildv1alpha1.BuildRunSpec{Timeout: &metav1.Duration{
			Duration: time.Duration(0),
		}},
		out: buildv1alpha1.BuildRunSpec{Timeout: nil},
	}, {
		name: "should clean-up an empty retention",
		in: buildv1alpha1.BuildRunSpec{
			Retention: &buildv1alpha1.BuildRunRetention{},
		},
		out: buildv1alpha1.BuildRunSpec{},
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			aCopy := tt.in.DeepCopy()
			SanitizeBuildRunSpec(aCopy)
			g.Expect(tt.out).To(Equal(*aCopy))
		})
	}
}
