package flags

import (
	"testing"

	"github.com/onsi/gomega"
	o "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

func TestStringPointerValue(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	flagName := "flag"
	value := "value"
	targetStr := "string"

	cmd := &cobra.Command{}
	flags := cmd.PersistentFlags()
	flags.Var(NewStringPointerValue(&targetStr), flagName, "")

	err := flags.Set(flagName, value)
	g.Expect(err).To(o.BeNil())

	v, err := flags.GetString(flagName)
	g.Expect(err).To(o.BeNil())
	g.Expect(value).To(o.Equal(v))
}
