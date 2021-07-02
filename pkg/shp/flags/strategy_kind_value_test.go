package flags

import (
	"testing"

	"github.com/onsi/gomega"
	o "github.com/onsi/gomega"
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
)

func TestStrategyKindValue(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	buildStrategyKind := buildv1alpha1.ClusterBuildStrategyKind
	v := NewStrategyKindValue(&buildStrategyKind)

	expected := buildv1alpha1.NamespacedBuildStrategyKind

	err := v.Set(string(expected))
	g.Expect(err).To(o.BeNil())

	g.Expect(string(expected)).To(o.Equal(v.String()))
	g.Expect(expected).To(o.Equal(buildStrategyKind))
}
