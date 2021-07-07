package flags

import (
	"fmt"
	"time"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// BuildrefNameFlag command-line flag.
	BuildrefNameFlag = "buildref-name"
	// BuilderImageFlag command-line flag.
	BuilderImageFlag = "builder-image"
	// BuilderCredentialsSecretFlag command-line flag.
	BuilderCredentialsSecretFlag = "builder-credentials-secret"
	// DockerfileFlag command-line flag.
	DockerfileFlag = "dockerfile"
	// SourceURLFlag command-line flag.
	SourceURLFlag = "source-url"
	// SourceRevisionFlag command-line flag.
	SourceRevisionFlag = "source-revision"
	// SourceContextDirFlag command-line flag.
	SourceContextDirFlag = "source-context-dir"
	// SourceCredentialsSecretFlag command-line flag.
	SourceCredentialsSecretFlag = "source-credentials-secret"
	// StrategyAPIVersionFlag command-line flag.
	StrategyAPIVersionFlag = "strategy-apiversion"
	// StrategyKindFlag command-line flag.
	StrategyKindFlag = "strategy-kind"
	// StrategyNameFlag command-line flag.
	StrategyNameFlag = "strategy-name"
	// OutputImageFlag command-line flag.
	OutputImageFlag = "output-image"
	// OutputCredentialsSecretFlag command-line flag.
	OutputCredentialsSecretFlag = "output-credentials-secret"
	// ServiceAccountNameFlag command-line flag.
	ServiceAccountNameFlag = "sa-name"
	// ServiceAccountGenerateFlag command-line flag.
	ServiceAccountGenerateFlag = "sa-generate"
	// TimeoutFlag command-line flag.
	TimeoutFlag = "timeout"
)

// sourceFlags flags for ".spec.source"
func sourceFlags(flags *pflag.FlagSet, source *buildv1alpha1.Source) {
	flags.StringVar(
		&source.URL,
		SourceURLFlag,
		"",
		"git repository source URL",
	)
	flags.StringVar(
		source.Revision,
		SourceRevisionFlag,
		"",
		"git repository source revision",
	)
	flags.StringVar(
		source.ContextDir,
		SourceContextDirFlag,
		"",
		"use a inner directory as context directory",
	)
	flags.StringVar(
		&source.Credentials.Name,
		SourceCredentialsSecretFlag,
		"",
		"name of the secret with git repository credentials",
	)
}

// strategyFlags flags for ".spec.strategy".
func strategyFlags(flags *pflag.FlagSet, strategy *buildv1alpha1.Strategy) {
	flags.StringVar(
		&strategy.APIVersion,
		StrategyAPIVersionFlag,
		buildv1alpha1.SchemeGroupVersion.Version,
		"kubernetes api-version of the build-strategy resource",
	)
	flags.Var(
		NewStrategyKindValue(strategy.Kind),
		StrategyKindFlag,
		"build-strategy kind",
	)
	flags.StringVar(
		&strategy.Name,
		StrategyNameFlag,
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

// dockerfileFlags register dockerfile flag as pointer to string.
func dockerfileFlags(flags *pflag.FlagSet, dockerfile *string) {
	flags.StringVar(
		dockerfile,
		DockerfileFlag,
		"",
		"path to dockerfile relative to repository",
	)
}

// timeoutFlags register a timeout flag as time.Duration instance.
func timeoutFlags(flags *pflag.FlagSet, timeout *metav1.Duration) {
	flags.DurationVar(
		&timeout.Duration,
		TimeoutFlag,
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
	flags.StringVar(
		sa.Name,
		ServiceAccountNameFlag,
		"",
		"Kubernetes service-account name",
	)
	flags.BoolVar(
		&sa.Generate,
		ServiceAccountGenerateFlag,
		false,
		"generate a Kubernetes service-account for the build",
	)
}
