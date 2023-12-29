package flags

import (
	"testing"

	o "github.com/onsi/gomega"
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
)

func TestSourcePruneOption(t *testing.T) {
	g := o.NewWithT(t)

	// Check for type and defaults
	g.Expect(pruneOptionFlag{}.Type()).To(o.Equal("pruneOption"))
	g.Expect(pruneOptionFlag{}.String()).To(o.Equal(string(buildv1alpha1.PruneNever)))

	var obj buildv1alpha1.PruneOption
	v := pruneOptionFlag{ref: &obj}

	// Check the supported values
	g.Expect(v.Set(string(buildv1alpha1.PruneNever))).Should(o.Succeed())
	g.Expect(v.String()).To(o.Equal(string(buildv1alpha1.PruneNever)))
	g.Expect(v.Set(string(buildv1alpha1.PruneAfterPull))).Should(o.Succeed())
	g.Expect(v.String()).To(o.Equal(string(buildv1alpha1.PruneAfterPull)))

	// Check that invalid values fail with the flag
	g.Expect(v.Set("invalid")).ToNot(o.Succeed())
}
