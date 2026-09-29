package flags

import (
	"testing"

	"github.com/onsi/gomega"
	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewStepResourcesArrayValue(t *testing.T) {
	g := gomega.NewWithT(t)

	t.Run("single limit", func(_ *testing.T) {
		spec := buildv1beta1.BuildRunSpec{}
		v := NewStepResourcesArrayValue(&spec.StepResources)

		g.Expect(v.Set("build-and-push=limits.memory=2Gi")).To(gomega.BeNil())

		g.Expect(spec.StepResources).To(gomega.HaveLen(1))
		g.Expect(spec.StepResources[0].Name).To(gomega.Equal("build-and-push"))
		g.Expect(spec.StepResources[0].Resources.Limits).To(gomega.Equal(corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("2Gi"),
		}))
	})

	t.Run("requests and limits merge on same step", func(_ *testing.T) {
		spec := buildv1beta1.BuildRunSpec{}
		v := NewStepResourcesArrayValue(&spec.StepResources)

		g.Expect(v.Set("build-and-push=limits.memory=2Gi")).To(gomega.BeNil())
		g.Expect(v.Set("build-and-push=requests.cpu=500m")).To(gomega.BeNil())

		g.Expect(spec.StepResources).To(gomega.HaveLen(1))
		g.Expect(spec.StepResources[0].Resources.Limits).To(gomega.Equal(corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("2Gi"),
		}))
		g.Expect(spec.StepResources[0].Resources.Requests).To(gomega.Equal(corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("500m"),
		}))
	})

	t.Run("distinct steps append", func(_ *testing.T) {
		spec := buildv1beta1.BuildRunSpec{}
		v := NewStepResourcesArrayValue(&spec.StepResources)

		g.Expect(v.Set("step-a=limits.memory=1Gi")).To(gomega.BeNil())
		g.Expect(v.Set("step-b=limits.memory=2Gi")).To(gomega.BeNil())

		g.Expect(spec.StepResources).To(gomega.HaveLen(2))
		g.Expect(spec.StepResources[0].Name).To(gomega.Equal("step-a"))
		g.Expect(spec.StepResources[1].Name).To(gomega.Equal("step-b"))
	})

	t.Run("missing step separator", func(_ *testing.T) {
		spec := buildv1beta1.BuildRunSpec{}
		v := NewStepResourcesArrayValue(&spec.StepResources)

		g.Expect(v.Set("limits.memory=2Gi")).NotTo(gomega.BeNil())
	})

	t.Run("missing resource value", func(_ *testing.T) {
		spec := buildv1beta1.BuildRunSpec{}
		v := NewStepResourcesArrayValue(&spec.StepResources)

		g.Expect(v.Set("build-and-push=limits.memory")).NotTo(gomega.BeNil())
	})

	t.Run("unknown resource kind", func(_ *testing.T) {
		spec := buildv1beta1.BuildRunSpec{}
		v := NewStepResourcesArrayValue(&spec.StepResources)

		g.Expect(v.Set("build-and-push=foo.memory=2Gi")).NotTo(gomega.BeNil())
	})

	t.Run("invalid quantity", func(_ *testing.T) {
		spec := buildv1beta1.BuildRunSpec{}
		v := NewStepResourcesArrayValue(&spec.StepResources)

		g.Expect(v.Set("build-and-push=limits.memory=notaquantity")).NotTo(gomega.BeNil())
	})
}
