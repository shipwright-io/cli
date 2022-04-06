package params

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/spf13/pflag"
)

func TestParamsCreation(t *testing.T) {
	g := gomega.NewWithT(t)

	flagset := pflag.NewFlagSet("name", 0)

	shpParams := NewParams()
	shpParams.AddFlags(flagset)

	testNs := "test"
	shpParams.configFlags.Namespace = &testNs

	client, err := shpParams.ClientSet()
	g.Expect(err).To(gomega.BeNil(), "Must not be an error during client creation")
	g.Expect(client).ToNot(gomega.BeNil(), "Client must not be nil")

	t.Run("Namespace", func(_ *testing.T) {
		ns := shpParams.Namespace()

		g.Expect(ns).To(gomega.Equal("test"))
	})

}
