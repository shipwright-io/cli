package cmd

import (
	"os"
	"testing"

	"github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
)

type mockedSubCommand struct{}

var testCmd = &cobra.Command{}

func (m *mockedSubCommand) Cmd() *cobra.Command {
	return testCmd
}

func (m *mockedSubCommand) Complete(client dynamic.Interface, ns string, args []string) error {
	return nil
}

func (m *mockedSubCommand) Validate() error {
	return nil
}

func (m *mockedSubCommand) Run(client dynamic.Interface, ns string) error {
	return nil
}

func TestCMD_Runner(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	o := NewOptions()
	testNs := "test"
	o.configFlags.Namespace = &testNs

	genericOpts := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	r := NewRunner(o, genericOpts, &mockedSubCommand{})

	t.Run("cmd", func(t *testing.T) {
		cmd := r.Cmd()

		g.Expect(cmd.RunE).ToNot(gomega.BeNil())
	})

	t.Run("dynamicClientNamespace", func(t *testing.T) {
		client, ns, err := r.dynamicClientNamespace()

		g.Expect(err).To(gomega.BeNil())
		g.Expect(ns).To(gomega.Equal(testNs))
		g.Expect(client).NotTo(gomega.BeNil())
	})

	t.Run("RunE", func(t *testing.T) {
		err := r.RunE(testCmd, []string{})

		g.Expect(err).To(gomega.BeNil())
	})
}
