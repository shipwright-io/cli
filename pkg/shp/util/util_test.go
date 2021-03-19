package util

import (
	"testing"

	"github.com/onsi/gomega"
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"

	"github.com/shipwright-io/cli/test/stub"
)

const (
	buildRunName = "test"
)

func TestUtil_ToUnstructured(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	name := "test"
	kind := "BuildRun"
	gvk := buildv1alpha1.SchemeBuilder.GroupVersion.WithKind(kind)
	br := stub.BuildRunEmpty()

	br.Name = buildRunName

	u, err := toUnstructured(name, gvk, &br)

	g.Expect(err).To(gomega.BeNil())
	g.Expect(u.GetName()).To(gomega.Equal(name))
	g.Expect(u.GetKind()).To(gomega.Equal(kind))

	var brNew buildv1alpha1.BuildRun
	fromUnstructured(u.UnstructuredContent(), &brNew)
	g.Expect(brNew.Name).To(gomega.Equal(name))
}
