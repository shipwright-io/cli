package flags

import (
	"testing"
	"time"

	o "github.com/onsi/gomega"
	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func TestBuildRunSpecFromFlags(t *testing.T) {
	g := o.NewWithT(t)

	str := "something-random"
	expected := &buildv1beta1.BuildRunSpec{
		Build: buildv1beta1.ReferencedBuild{
			Name: &str,
		},
		ServiceAccount: &str,
		Timeout: &metav1.Duration{
			Duration: 1 * time.Second,
		},
		Output: &buildv1beta1.Image{
			PushSecret:  ptr.To("name"),
			Image:       str,
			Insecure:    ptr.To(false),
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Retention: &buildv1beta1.BuildRunRetention{
			TTLAfterFailed: &metav1.Duration{
				Duration: 48 * time.Hour,
			},
			TTLAfterSucceeded: &metav1.Duration{
				Duration: 30 * time.Minute,
			},
		},
		NodeSelector:     map[string]string{"kubernetes.io/hostname": "worker-1"},
		SchedulerName:    ptr.To("dolphinscheduler"),
		RuntimeClassName: ptr.To("kata"),
	}

	cmd := &cobra.Command{}
	flags := cmd.PersistentFlags()
	spec := BuildRunSpecFromFlags(flags)

	t.Run(".spec.buildRef", func(_ *testing.T) {
		err := flags.Set(BuildrefNameFlag, *expected.Build.Name)
		g.Expect(err).To(o.BeNil())

		g.Expect(expected.Build).To(o.Equal(spec.Build), "spec.buildRef")
	})

	t.Run(".spec.serviceAccount", func(_ *testing.T) {
		err := flags.Set(ServiceAccountNameFlag, *expected.ServiceAccount)
		g.Expect(err).To(o.BeNil())

		g.Expect(*expected.ServiceAccount).To(o.Equal(*spec.ServiceAccount), "spec.serviceAccount")
	})

	t.Run(".spec.timeout", func(_ *testing.T) {
		err := flags.Set(TimeoutFlag, expected.Timeout.Duration.String())
		g.Expect(err).To(o.BeNil())

		g.Expect(*expected.Timeout).To(o.Equal(*spec.Timeout), "spec.timeout")
	})

	t.Run(".spec.output", func(_ *testing.T) {
		err := flags.Set(OutputImageFlag, expected.Output.Image)
		g.Expect(err).To(o.BeNil())

		err = flags.Set(OutputCredentialsSecretFlag, *expected.Output.PushSecret)
		g.Expect(err).To(o.BeNil())

		g.Expect(*expected.Output).To(o.Equal(*spec.Output), "spec.output")
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

	t.Run(".spec.runtimeClassName", func(_ *testing.T) {
		err := flags.Set(RuntimeClassNameFlag, *expected.RuntimeClassName)
		g.Expect(err).To(o.BeNil())

		g.Expect(expected.RuntimeClassName).To(o.Equal(spec.RuntimeClassName), "spec.runtimeClassName")
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

func TestSanitizeBuildRunSpec(t *testing.T) {
	g := o.NewWithT(t)

	name := "name"
	completeBuildRunSpec := buildv1beta1.BuildRunSpec{
		ServiceAccount: &name,
		Output: &buildv1beta1.Image{
			PushSecret: ptr.To("name"),
			Image:      "image",
		},
	}

	testCases := []struct {
		name string
		in   buildv1beta1.BuildRunSpec
		out  buildv1beta1.BuildRunSpec
	}{{
		name: "all empty should stay empty",
		in:   buildv1beta1.BuildRunSpec{},
		out:  buildv1beta1.BuildRunSpec{},
	}, {
		name: "should clean-up service-account",
		in:   buildv1beta1.BuildRunSpec{ServiceAccount: ptr.To("")},
		out:  buildv1beta1.BuildRunSpec{},
	}, {
		name: "should clean-up output",
		in:   buildv1beta1.BuildRunSpec{Output: &buildv1beta1.Image{}},
		out:  buildv1beta1.BuildRunSpec{},
	}, {
		name: "should not clean-up complete objects",
		in:   completeBuildRunSpec,
		out:  completeBuildRunSpec,
	}, {
		name: "should clean-up 0s duration",
		in: buildv1beta1.BuildRunSpec{Timeout: &metav1.Duration{
			Duration: time.Duration(0),
		}},
		out: buildv1beta1.BuildRunSpec{Timeout: nil},
	}, {
		name: "should clean-up an empty retention",
		in: buildv1beta1.BuildRunSpec{
			Retention: &buildv1beta1.BuildRunRetention{},
		},
		out: buildv1beta1.BuildRunSpec{},
	}, {
		name: "should clean-up runtime-class-name",
		in:   buildv1beta1.BuildRunSpec{RuntimeClassName: ptr.To("")},
		out:  buildv1beta1.BuildRunSpec{},
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(_ *testing.T) {
			aCopy := tt.in.DeepCopy()
			SanitizeBuildRunSpec(aCopy)
			g.Expect(tt.out).To(o.Equal(*aCopy))
		})
	}
}
