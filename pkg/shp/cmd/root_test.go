package cmd

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/onsi/gomega"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd/types"
	"github.com/shipwright-io/cli/test/stub"
)

func TestCMD_NewCmdSHP(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	genericOpts := &genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	p := &types.ClientSets{}

	cmd := NewRootCmd(context.Background(), genericOpts, p)

	out, err := stub.ExecuteCommand(cmd, "build", "cr")

	if err == nil {
		t.Errorf("No errors was defined. Output: %s", out)
	}

	expected := fmt.Sprintf("unknown command %q for %q\n\nDid you mean this?\n\t%s\n", "cr", "shp build", "create")

	g.Expect(err.Error()).To(gomega.Equal(expected))
}
