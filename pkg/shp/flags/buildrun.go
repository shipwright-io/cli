package flags

import (
	"time"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/pflag"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BuildRunSpecFlags BuildRun's spec represtantation as command-line flags.
func BuildRunSpecFlags(flags *pflag.FlagSet) *buildv1alpha1.BuildRunSpec {
	empty := ""
	spec := &buildv1alpha1.BuildRunSpec{
		BuildRef:       &buildv1alpha1.BuildRef{},
		ServiceAccount: &buildv1alpha1.ServiceAccount{Name: &empty},
		Timeout:        &metav1.Duration{},
		Output: &buildv1alpha1.Image{
			Credentials: &corev1.LocalObjectReference{},
		},
	}

	buildRefFlags(flags, spec.BuildRef)
	serviceAccountFlags(flags, spec.ServiceAccount)
	timeoutFlags(flags, spec.Timeout)
	outputFlags(flags, spec.Output)

	return spec
}

// buildRefFlags register flags for BuildRun's spec.buildRef attribute.
func buildRefFlags(flags *pflag.FlagSet, buildRef *buildv1alpha1.BuildRef) {
	flags.StringVar(
		&buildRef.Name,
		"buildref-name",
		"",
		"name of build resource to reference",
	)
	flags.StringVar(
		&buildRef.APIVersion,
		"buildref-apiversion",
		"",
		"API version of build resource to reference",
	)
}

// serviceAccountFlags register flags for BuildRun's spec.serviceAccount attribute.
func serviceAccountFlags(flags *pflag.FlagSet, sa *buildv1alpha1.ServiceAccount) {
	flags.StringVar(
		sa.Name,
		"sa-name",
		"",
		"service-account name",
	)
	flags.BoolVar(
		&sa.Generate,
		"sa-generate",
		false,
		"generate a service-account for the build",
	)
}

// serviceAccountFlags register flags for BuildRun's spec.timeout attribute.
func timeoutFlags(flags *pflag.FlagSet, timeout *metav1.Duration) {
	flags.DurationVar(
		&timeout.Duration,
		"timeout",
		time.Duration(0),
		"build process timeout",
	)
}

// serviceAccountFlags register flags for BuildRun's spec.output attribute.
func outputFlags(flags *pflag.FlagSet, output *buildv1alpha1.Image) {
	flags.StringVar(
		&output.Image,
		"output-image",
		"",
		"output image URL",
	)
	secretRefFlags(flags, output.Credentials)
}

// serviceAccountFlags register flags for BuildRun's spec.output.credentials attribute.
func secretRefFlags(flags *pflag.FlagSet, secretRef *corev1.LocalObjectReference) {
	flags.StringVar(
		&secretRef.Name,
		"output-credentials",
		"",
		"Kubernetes Secret with credentials for output container registry.",
	)
}
