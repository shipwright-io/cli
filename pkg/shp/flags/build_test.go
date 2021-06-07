package flags

import (
	"testing"

	"github.com/onsi/gomega"
	o "github.com/onsi/gomega"
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

func TestSanitizeBuildSpec(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	completeBuildSpec := buildv1alpha1.BuildSpec{
		Source: buildv1alpha1.Source{
			Credentials: &v1.LocalObjectReference{Name: "name"},
		},
		Builder: &buildv1alpha1.Image{
			Credentials: &v1.LocalObjectReference{Name: "name"},
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
			Credentials: &v1.LocalObjectReference{},
		}},
		out: buildv1alpha1.BuildSpec{},
	}, {
		name: "should clean-up `.spec.builder.credentials`",
		in: buildv1alpha1.BuildSpec{Builder: &buildv1alpha1.Image{
			Credentials: &v1.LocalObjectReference{},
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
