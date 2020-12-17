package cmd

import (
	"os"
	"testing"

	"github.com/onsi/gomega"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestCMD_NewCmdSHP(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	genericOpts := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	cmd := NewCmdSHP(genericOpts)

	g.Expect(cmd).NotTo(gomega.BeNil())
}
