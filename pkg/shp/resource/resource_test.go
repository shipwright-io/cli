package resource

import (
	"testing"

	"github.com/onsi/gomega"
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"

	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/test/stub"
)

func TestResource(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	kind := "Build"
	resource := "builds"

	name := "name"
	source := "source"
	image := "image-url"

	p := params.NewParams()

	buildResource := NewShpResource(p, buildv1alpha1.SchemeGroupVersion, kind, resource)
	ri, err := buildResource.getResourceInterface()

	g.Expect(ri).NotTo(gomega.BeNil(), "ResourceInterface should not be nil")
	g.Expect(err).To(gomega.BeNil(), "Error must be nil")

	buildResource.resourceInterface = stub.NewFakeClient().Resource(buildv1alpha1.SchemeGroupVersion.WithResource(resource))

	build := stub.TestBuild(name, image, source)
	err = buildResource.Create("name", build)

	g.Expect(err).To(gomega.BeNil(), "Error from creation must be nil")

	t.Run("Resource Get", func(t *testing.T) {
		var build1 buildv1alpha1.Build
		err = buildResource.Get(name, &build1)

		g.Expect(err).To(gomega.BeNil(), "Error from creation must be nil")
		g.Expect(build1.Name).To(gomega.Equal(name))
		g.Expect(build1.Spec.Source.URL).To(gomega.Equal(source))
	})

	t.Run("Resource List", func(t *testing.T) {
		var buildList buildv1alpha1.BuildList
		err = buildResource.List(&buildList)

		g.Expect(err).To(gomega.BeNil(), "Error from List must be nil")
		g.Expect(len(buildList.Items)).To(gomega.Equal(1))
	})

	err = buildResource.Delete(name)
	g.Expect(err).To(gomega.BeNil(), "Error from Delete must be nil")

}
