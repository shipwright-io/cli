package flags

import (
	"fmt"

	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
)

// ParamArrayValue implements pflag.Value interface, in order to store ParamValue key-value
// pairs used on Shipwright's BuildSpec.
type ParamArrayValue struct {
	params *[]buildv1beta1.ParamValue // pointer to the slice of ParamValue
}

// String prints out the string representation of the slice of EnvVar objects.
func (p *ParamArrayValue) String() string {
	slice := []string{}
	for _, e := range *p.params {
		slice = append(slice, fmt.Sprintf("%s=%v", e.Name, e.Value))
	}
	csv, _ := writeAsCSV(slice)
	return fmt.Sprintf("[%s]", csv)
}

// Set receives a key-value entry separated by equal sign ("=").
func (p *ParamArrayValue) Set(value string) error {
	k, v, err := splitKeyValue(value)
	if err != nil {
		return err
	}
	for _, e := range *p.params {
		if k == e.Name {
			return fmt.Errorf("environment variable '%s' is already set", k)
		}
	}
	*p.params = append(*p.params, buildv1beta1.ParamValue{Name: k, SingleValue: &buildv1beta1.SingleValue{Value: &v}})
	return nil
}

// Type analogous to the pflag "stringArray" type, where each flag entry will be tranlated to a
// single array (slice) entry, therefore the comma (",") is accepted as part of the value, as any
// other special character.
func (p *ParamArrayValue) Type() string {
	return "stringArray"
}

// NewCoreEnvVarArrayValue instantiate a ParamValSliceValue sharing the EnvVar pointer.
func NewParamArrayValue(params *[]buildv1beta1.ParamValue) *ParamArrayValue {
	return &ParamArrayValue{params: params}
}
