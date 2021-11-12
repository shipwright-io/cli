package flags

import (
	"testing"

	. "github.com/onsi/gomega"
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
)

func TestStrategyKindValue(t *testing.T) {
	g := NewWithT(t)

	buildStrategyKind := buildv1alpha1.ClusterBuildStrategyKind
	v := NewStrategyKindValue(&buildStrategyKind)

	expected := buildv1alpha1.NamespacedBuildStrategyKind

	err := v.Set(string(expected))
	g.Expect(err).To(BeNil())

	g.Expect(string(expected)).To(Equal(v.String()))
	g.Expect(expected).To(Equal(buildStrategyKind))
}
