package flags

import (
	"fmt"
	"time"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	BuildrefNameFlag = "buildref-name"
	SourceURLFlag    = "source-url"
	OutputImageFlag  = "output-image"
)

// sourceFlags flags for ".spec.source"
func sourceFlags(flags *pflag.FlagSet, source *buildv1alpha1.Source) {
	flags.StringVar(
		&source.URL,
		SourceURLFlag,
		"",
		"git repository source URL",
	)
	flags.Var(
		NewStringPointerValue(source.Revision),
		"source-revision",
		"git repository source revision",
	)
	flags.Var(
		NewStringPointerValue(source.ContextDir),
		"source-context-dir",
		"use a inner directory as context directory",
	)
	flags.StringVar(
		&source.Credentials.Name,
		"source-credentials-secret",
		"",
		"name of the secret with git repository credentials",
	)
}

// strategyFlags flags for ".spec.strategy".
func strategyFlags(flags *pflag.FlagSet, strategy *buildv1alpha1.Strategy) {
	flags.StringVar(
		&strategy.APIVersion,
		"strategy-apiversion",
		buildv1alpha1.SchemeGroupVersion.Version,
		"kubernetes api-version of the build-strategy resource",
	)
	flags.Var(
		NewStrategyKindValue(strategy.Kind),
		"strategy-kind",
		"build-strategy kind",
	)
	flags.StringVar(
		&strategy.Name,
		"strategy-name",
		"buildpacks-v3",
		"build-strategy name",
	)
}

// imageFlags flags for Shipwright's Image definition, using a prefix to avoid duplicated flags.
func imageFlags(flags *pflag.FlagSet, prefix string, image *buildv1alpha1.Image) {
	flags.StringVar(
		&image.Image,
		fmt.Sprintf("%s-image", prefix),
		"",
		"image employed during the building process",
	)
	flags.StringVar(
		&image.Credentials.Name,
		fmt.Sprintf("%s-credentials-secret", prefix),
		"",
		"name of the secret with builder-image pull credentials",
	)
}

// timeoutFlags register a timeout flag as time.Duration instance.
func timeoutFlags(flags *pflag.FlagSet, timeout *metav1.Duration) {
	flags.DurationVar(
		&timeout.Duration,
		"timeout",
		time.Duration(0),
		"build process timeout",
	)
}

// buildRefFlags register flags for BuildRun's spec.buildRef attribute.
func buildRefFlags(flags *pflag.FlagSet, buildRef *buildv1alpha1.BuildRef) {
	flags.StringVar(
		&buildRef.Name,
		BuildrefNameFlag,
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
	flags.Var(
		NewStringPointerValue(sa.Name),
		"sa-name",
		"service-account name",
	)
	flags.BoolVar(
		&sa.Generate,
		"sa-generate",
		false,
		"generate a service-account for the build",
	)
}
