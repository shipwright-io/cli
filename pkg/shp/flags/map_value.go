package flags

import (
	"fmt"
)

// MapValue implements pflag.Value interface, in order to store key-value
// pairs used on Shipwright's BuildSpec which have map[string]string as field type.
type MapValue struct {
	kvMap map[string]string
}

// String prints out the string representation of the map.
func (m *MapValue) String() string {
	slice := []string{}
	for k, v := range m.kvMap {
		slice = append(slice, fmt.Sprintf("%s=%s", k, v))
	}
	csv, _ := writeAsCSV(slice)
	return fmt.Sprintf("[%s]", csv)
}

// Set receives a key-value entry separated by equal sign ("=").
func (m *MapValue) Set(value string) error {
	k, v, err := splitKeyValue(value)
	if err != nil {
		return err
	}
	m.kvMap[k] = v
	return nil
}

// Type analogous to the pflag "stringArray" type, where each flag entry will be translated to a
// single array (slice) entry, therefore the comma (",") is accepted as part of the value, as any
// other special character.
func (c *MapValue) Type() string {
	return "stringArray"
}

// NewMapValue instantiate a MapValue sharing the map.
func NewMapValue(m map[string]string) *MapValue {
	return &MapValue{kvMap: m}
}
