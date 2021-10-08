package suggestion

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/onsi/gomega"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd/build"
	"github.com/shipwright-io/cli/pkg/shp/cmd/types"
)

func TestSuggestion(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	genericOpts := &genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	p := &types.ClientSets{}
	cmd := build.Command(context.Background(), genericOpts, p)

	err := SubcommandsRequiredWithSuggestions(cmd, []string{"cr"})

	expected := fmt.Sprintf("unknown command %q for %q\n\nDid you mean this?\n\t%s\n", "cr", "build", "create")

	g.Expect(err.Error()).To(gomega.Equal(expected))
}
