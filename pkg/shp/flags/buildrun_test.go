package flags

import (
	"testing"

	"github.com/onsi/gomega"
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	v1 "k8s.io/api/core/v1"

	o "github.com/onsi/gomega"
)

func TestSanitizeBuildRunSpec(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	name := "name"
	completeBuildRunSpec := buildv1alpha1.BuildRunSpec{
		ServiceAccount: &buildv1alpha1.ServiceAccount{Name: &name},
		Output: &buildv1alpha1.Image{
			Credentials: &v1.LocalObjectReference{Name: "name"},
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
