package util

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/otaviof/shp/test/stub"
)

func TestUtil_ToUnstructured(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	name := "test"
	kind := "BuildRun"
	br := stub.BuildRunEmpty()

	u, err := ToUnstructured(name, kind, &br)

	g.Expect(err).To(gomega.BeNil())
	g.Expect(u.GetName()).To(gomega.Equal(name))
	g.Expect(u.GetKind()).To(gomega.Equal(kind))
}
