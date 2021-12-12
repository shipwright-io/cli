package flags

import (
	"testing"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"

	o "github.com/onsi/gomega"
)

func TestNewMapValue(t *testing.T) {
	g := o.NewWithT(t)

	spec := &buildv1alpha1.BuildSpec{Output: buildv1alpha1.Image{
		Labels: map[string]string{},
	}}
	c := NewMapValue(spec.Output.Labels)

	// expect error when key-value is not split by equal sign
	err := c.Set("a")
	g.Expect(err).NotTo(o.BeNil())

	// setting a simple key-value entry
	err = c.Set("a=b")
	g.Expect(err).To(o.BeNil())
	g.Expect(len(spec.Output.Labels)).To(o.Equal(1))
	g.Expect(spec.Output.Labels["a"]).To(o.Equal("b"))

	// setting a key-value entry with special characters
	err = c.Set("b=c,d,e=f")
	g.Expect(err).To(o.BeNil())
	g.Expect(len(spec.Output.Labels)).To(o.Equal(2))
	g.Expect(spec.Output.Labels["b"]).To(o.Equal("c,d,e=f"))

	// setting a key-value entry with space on it
	err = c.Set("c=d e")
	g.Expect(err).To(o.BeNil())
	g.Expect(len(spec.Output.Labels)).To(o.Equal(3))
	g.Expect(spec.Output.Labels["c"]).To(o.Equal("d e"))

	// verify the map status
	g.Expect(c.kvMap).To(o.BeEquivalentTo(map[string]string{
		"a": "b",
		"b": "c,d,e=f",
		"c": "d e",
	}))

	// making sure the string representation produced is as expected
	s := c.String()
	g.Expect(s).To(o.BeElementOf(
		"[a=b,\"b=c,d,e=f\",c=d e]",
		"[a=b,c=d e,\"b=c,d,e=f\"]",
		"[\"b=c,d,e=f\",a=b,c=d e]",
		"[\"b=c,d,e=f\",c=d e,a=b]",
		"[c=d e,a=b,\"b=c,d,e=f\"]",
		"[c=d e,\"b=c,d,e=f\",a=b,\"b=c,d,e=f\"]",
	))
}
