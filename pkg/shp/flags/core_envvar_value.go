package flags

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

// CoreEnvVarArrayValue implements pflag.Value interface, in order to store corev1.EnvVar key-value
// pairs used on Shipwright's BuildSpec.
type CoreEnvVarArrayValue struct {
	envs *[]corev1.EnvVar // pointer to the slice of EnvVar
}

// String prints out the string representation of the slice of EnvVar objects.
func (c *CoreEnvVarArrayValue) String() string {
	slice := []string{}
	for _, e := range *c.envs {
		slice = append(slice, fmt.Sprintf("%s=%s", e.Name, e.Value))
	}
	csv, _ := writeAsCSV(slice)
	return fmt.Sprintf("[%s]", csv)
}

// Set receives a key-value entry separated by equal sign ("=").
func (c *CoreEnvVarArrayValue) Set(value string) error {
	k, v, err := splitKeyValue(value)
	if err != nil {
		return err
	}
	for _, e := range *c.envs {
		if k == e.Name {
			return fmt.Errorf("environment variable '%s' is already set", k)
		}
	}
	*c.envs = append(*c.envs, corev1.EnvVar{Name: k, Value: v})
	return nil
}

// Type analogous to the pflag "stringArray" type, where each flag entry will be tranlated to a
// single array (slice) entry, therefore the comma (",") is accepted as part of the value, as any
// other special character.
func (c *CoreEnvVarArrayValue) Type() string {
	return "stringArray"
}

// NewCoreEnvVarArrayValue instantiate a CoreEnvVarSliceValue sharing the EnvVar pointer.
func NewCoreEnvVarArrayValue(envs *[]corev1.EnvVar) *CoreEnvVarArrayValue {
	return &CoreEnvVarArrayValue{envs: envs}
}
