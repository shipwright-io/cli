package buildrun

import (
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/onsi/gomega"
	"github.com/otaviof/shp/pkg/shp/util"
	"github.com/otaviof/shp/test/stub"
	"github.com/spf13/cobra"
)

// TestBuildRun_newBuildRunCreate tests the "create" verb on the build-run sub-command handler.
func TestBuildRun_newBuildRunCreate(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	cmd := &cobra.Command{}
	b := newBuildRun(cmd, util.Create)

	ns := "test"
	name := "test"
	client := stub.NewFakeClient()

	// expect cobra.Command to be passed along
	t.Run("Cmd", func(t *testing.T) {
		cmd := b.Cmd()
		g.Expect(cmd).NotTo(gomega.BeNil())
	})

	// expect to have the BuildRun instance to not return error, and have the attribute "name" set
	// accordingly to the arguments informed
	t.Run("Complete", func(t *testing.T) {
		err := b.Complete(client, ns, []string{"build-run", name})

		g.Expect(err).To(gomega.BeNil())
		g.Expect(b.name).To(gomega.Equal(name))
	})

	// expect validation not to return error
	t.Run("Validate", func(t *testing.T) {
		err := b.Validate()

		g.Expect(err).To(gomega.BeNil())
	})

	// expect Run method to create the BuildRun resource in the fake client, using shared namespace
	// and resource names
	t.Run("Run", func(t *testing.T) {
		err := b.Run(client, ns)

		g.Expect(err).To(gomega.BeNil())

		rs := buildRunResourceClient(client, ns)
		buildRun, err := rs.Get(name, v1.GetOptions{})

		g.Expect(err).To(gomega.BeNil())
		g.Expect(buildRun).NotTo(gomega.BeNil())
		g.Expect(buildRun.GetName()).To(gomega.Equal(name))
	})
}
