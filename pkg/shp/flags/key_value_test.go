package flags

import (
	"testing"
)

func TestSplitKeyValue(t *testing.T) {
	testCases := []struct {
		name    string
		value   string
		k       string
		v       string
		wantErr bool
	}{{
		"error when no value is informed",
		"k",
		"",
		"",
		true,
	}, {
		"error when only value is informed",
		"=v",
		"",
		"",
		true,
	}, {
		"success on spliting 'key=value'",
		"k=v",
		"k",
		"v",
		false,
	}, {
		"success on splitting 'key=value=value' (double equal sign)",
		"k=v=v",
		"k",
		"v=v",
		false,
	}}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			k, v, err := splitKeyValue(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("splitKeyValue() error='%v', wantErr '%v'", err, tt.wantErr)
				return
			}
			if k != tt.k {
				t.Errorf("splitKeyValue() key='%v', want '%v'", k, tt.k)
			}
			if v != tt.v {
				t.Errorf("splitKeyValue() value='%v', want '%v'", v, tt.v)
			}
		})
	}
}
