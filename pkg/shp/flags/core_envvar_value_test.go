package flags

import (
	"testing"

	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	corev1 "k8s.io/api/core/v1"

	o "github.com/onsi/gomega"
)

func TestCoreEnvVarSliceValue(t *testing.T) {
	g := o.NewWithT(t)

	spec := &buildv1beta1.BuildSpec{Env: []corev1.EnvVar{}}
	c := NewCoreEnvVarArrayValue(&spec.Env)

	// expect error when key-value is not split by equal sign
	err := c.Set("a")
	g.Expect(err).NotTo(o.BeNil())

	// setting a simple key-value entry
	err = c.Set("a=b")
	g.Expect(err).To(o.BeNil())
	g.Expect(len(spec.Env)).To(o.Equal(1))
	g.Expect(spec.Env[0].Name).To(o.Equal("a"))
	g.Expect(spec.Env[0].Value).To(o.Equal("b"))

	// setting a key-value entry with special characters
	err = c.Set("b=c,d,e=f")
	g.Expect(err).To(o.BeNil())
	g.Expect(len(spec.Env)).To(o.Equal(2))
	g.Expect(spec.Env[1].Name).To(o.Equal("b"))
	g.Expect(spec.Env[1].Value).To(o.Equal("c,d,e=f"))

	// setting a key-value entry with space on it
	err = c.Set("c=d e")
	g.Expect(err).To(o.BeNil())
	g.Expect(len(spec.Env)).To(o.Equal(3))
	g.Expect(spec.Env[2].Name).To(o.Equal("c"))
	g.Expect(spec.Env[2].Value).To(o.Equal("d e"))

	// on trying to insert a repeated value, it should error
	err = c.Set("a=b")
	g.Expect(err).NotTo(o.BeNil())

	// making sure the string representation produced is as expected
	s := c.String()
	g.Expect(s).To(o.Equal("[a=b,\"b=c,d,e=f\",c=d e]"))
}
