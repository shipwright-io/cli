package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/onsi/gomega"
	"github.com/shipwright-io/cli/test/stub"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestCMD_NewCmdSHP(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	genericOpts := &genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	cmd := NewCmdSHP(genericOpts)

	out, err := stub.ExecuteCommand(cmd, "build", "cr")

	if err == nil {
		t.Errorf("No errors was defined. Output: %s", out)
	}

	expected := fmt.Sprintf("unknown command %q for %q\n\nDid you mean this?\n\t%s\n", "cr", "shp build", "create")

	g.Expect(err.Error()).To(gomega.Equal(expected))
}
