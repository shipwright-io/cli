package util

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

func StringSliceToEnvVarSlice(envs []string) ([]corev1.EnvVar, error) {
	envVars := []corev1.EnvVar{}

	for _, l := range envs {
		parts := strings.SplitN(l, "=", 2)
		if len(parts) == 1 {
			return envVars, fmt.Errorf("failed to parse key-value pair %q, not enough parts", l)
		}
		envVars = append(envVars, corev1.EnvVar{Name: parts[0], Value: parts[1]})
	}

	return envVars, nil
}
