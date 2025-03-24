package flags

import (
	"testing"

	o "github.com/onsi/gomega"
	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
)

func TestStrategyKindValue(t *testing.T) {
	g := o.NewWithT(t)

	buildStrategyKind := buildv1beta1.ClusterBuildStrategyKind
	v := NewStrategyKindValue(&buildStrategyKind)

	expected := buildv1beta1.NamespacedBuildStrategyKind

	err := v.Set(string(expected))
	g.Expect(err).To(o.BeNil())

	g.Expect(string(expected)).To(o.Equal(v.String()))
	g.Expect(expected).To(o.Equal(buildStrategyKind))
}
