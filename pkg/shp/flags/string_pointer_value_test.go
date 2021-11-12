package flags

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

func TestStringPointerValue(t *testing.T) {
	g := NewWithT(t)

	flagName := "flag"
	value := "value"
	targetStr := "string"

	cmd := &cobra.Command{}
	flags := cmd.PersistentFlags()
	flags.Var(NewStringPointerValue(&targetStr), flagName, "")

	err := flags.Set(flagName, value)
	g.Expect(err).To(BeNil())

	v, err := flags.GetString(flagName)
	g.Expect(err).To(BeNil())
	g.Expect(value).To(Equal(v))
}
