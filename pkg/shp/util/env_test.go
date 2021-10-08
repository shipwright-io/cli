package util

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestStringSliceToEnvVar(t *testing.T) {
	type args struct {
		envs []string
	}
	tests := []struct {
		name string
		args args
		want []corev1.EnvVar
	}{
		{
			name: "no envs",
			args: args{
				envs: []string{},
			},
			want: []corev1.EnvVar{},
		},
		{
			name: "one env",
			args: args{
				envs: []string{"my-name=my-value"},
			},
			want: []corev1.EnvVar{
				{Name: "my-name", Value: "my-value"},
			},
		},
		{
			name: "multiple envs",
			args: args{
				envs: []string{"name-one=value-one", "name-two=value-two", "name-three=value-three"},
			},
			want: []corev1.EnvVar{
				{Name: "name-one", Value: "value-one"},
				{Name: "name-two", Value: "value-two"},
				{Name: "name-three", Value: "value-three"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringSliceToEnvVarSlice(tt.args.envs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StringSliceToEnvVar() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
