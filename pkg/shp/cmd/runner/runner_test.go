package runner

import (
	"os"
	"testing"

	"github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/params"
)

type mockedSubCommand struct{}

var testCmd = &cobra.Command{}

func (m *mockedSubCommand) Cmd() *cobra.Command {
	return testCmd
}

func (m *mockedSubCommand) Complete(p *params.Params, ioStreams *genericclioptions.IOStreams, args []string) error {
	return nil
}

func (m *mockedSubCommand) Validate() error {
	return nil
}

func (m *mockedSubCommand) Run(p *params.Params, ioStreams *genericclioptions.IOStreams) error {
	return nil
}

func TestCMD_Runner(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	p := params.NewParams()

	genericStreams := &genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	r := NewRunner(p, genericStreams, &mockedSubCommand{})

	t.Run("cmd", func(t *testing.T) {
		cmd := r.Cmd()

		g.Expect(cmd.RunE).ToNot(gomega.BeNil())
	})

	t.Run("RunE", func(t *testing.T) {
		err := r.RunE(testCmd, []string{})

		g.Expect(err).To(gomega.BeNil())
	})
}
