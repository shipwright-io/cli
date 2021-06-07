package flags

// StringPointerValue implements pflag.Value interface, to represent a pointer to an string as a
// command-line flag with Cobra.
type StringPointerValue struct {
	stringPtr *string
}

// String returns the value as string, so when the pointer is nil it returns a empty string.
func (s *StringPointerValue) String() string {
	if s.stringPtr == nil {
		return ""
	}
	return *s.stringPtr
}

// Set set a new value. When empty it keeps the current value set.
func (s *StringPointerValue) Set(str string) error {
	if str == "" {
		return nil
	}
	s.stringPtr = &str
	return nil
}

// Type analogous to the pflag "string".
func (s *StringPointerValue) Type() string {
	return "string"
}

// NewStringPointerValue instantiate StringPointerValue with the default pointer to string.
func NewStringPointerValue(stringPtr *string) *StringPointerValue {
	return &StringPointerValue{stringPtr: stringPtr}
}
