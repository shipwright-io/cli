package cmd

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/spf13/pflag"
)

func TestCMD_Options(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	o := NewOptions()

	g.Expect(o.configFlags).NotTo(gomega.BeNil())
	g.Expect(o.matchVersionFlags).NotTo(gomega.BeNil())

	t.Run("AddFlags", func(t *testing.T) {
		flags := pflag.NewFlagSet("unit-test", pflag.ExitOnError)
		o.AddFlags(flags)

		dryRun, err := flags.GetBool("dry-run")
		g.Expect(err).To(gomega.BeNil())
		g.Expect(dryRun).To(gomega.BeFalse())
	})
}
