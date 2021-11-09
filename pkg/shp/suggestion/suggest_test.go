package suggestion

import (
	"fmt"
	"os"
	"testing"

	"github.com/onsi/gomega"
	"github.com/shipwright-io/cli/pkg/shp/cmd/build"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestSuggestion(t *testing.T) {
	g := gomega.NewWithT(t)

	genericOpts := &genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	cmd := build.Command(nil, genericOpts)

	err := SubcommandsRequiredWithSuggestions(cmd, []string{"cr"})

	expected := fmt.Sprintf("unknown command %q for %q\n\nDid you mean this?\n\t%s\n", "cr", "build", "create")

	g.Expect(err.Error()).To(gomega.Equal(expected))
}
