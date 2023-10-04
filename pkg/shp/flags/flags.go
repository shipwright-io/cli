package flags

import (
	"fmt"
	"time"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/pflag"

	corev1 "k8s.io/api/core/v1"
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
	// EnvFlag command-line flag.
	EnvFlag = "env"
	// SourceURLFlag command-line flag.
	SourceURLFlag = "source-url"
	// SourceRevisionFlag command-line flag.
	SourceRevisionFlag = "source-revision"
	// SourceContextDirFlag command-line flag.
	SourceContextDirFlag = "source-context-dir"
	// SourceCredentialsSecretFlag command-line flag.
	SourceCredentialsSecretFlag = "source-credentials-secret" // #nosec G101
	// SourceBundleImageFlag command-line flag
	SourceBundleImageFlag = "source-bundle-image"
	// SourceBundlePruneFlag command-line flag
	SourceBundlePruneFlag = "source-bundle-prune"
	// StrategyAPIVersionFlag command-line flag.
	StrategyAPIVersionFlag = "strategy-apiversion"
	// StrategyKindFlag command-line flag.
	StrategyKindFlag = "strategy-kind"
	// StrategyNameFlag command-line flag.
	StrategyNameFlag = "strategy-name"
	// OutputImageFlag command-line flag.
	OutputImageFlag = "output-image"
	// OutputInsecure command-line flag.
	OutputInsecureFlag = "output-insecure"
	// OutputCredentialsSecretFlag command-line flag.
	OutputCredentialsSecretFlag = "output-credentials-secret" // #nosec G101
	// ServiceAccountNameFlag command-line flag.
	ServiceAccountNameFlag = "sa-name"
	// ServiceAccountGenerateFlag command-line flag.
	ServiceAccountGenerateFlag = "sa-generate"
	// TimeoutFlag command-line flag.
	TimeoutFlag = "timeout"
	// OutputImageLabelsFlag command-line flag.
	OutputImageLabelsFlag = "output-image-label"
	// OutputImageAnnotationsFlag command-line flag.
	OutputImageAnnotationsFlag = "output-image-annotation"
	// RetentionFailedLimitFlag command-line flag.
	RetentionFailedLimitFlag = "retention-failed-limit"
	// RetentionSucceededLimitFlag command-line flag.
	RetentionSucceededLimitFlag = "retention-succeeded-limit"
	// RetentionTTLAfterFailedFlag command-line flag.
	RetentionTTLAfterFailedFlag = "retention-ttl-after-failed"
	// RetentionTTLAfterSucceededFlag command-line flag.
	RetentionTTLAfterSucceededFlag = "retention-ttl-after-succeeded"
)

// sourceFlags flags for ".spec.source"
func sourceFlags(flags *pflag.FlagSet, source *buildv1alpha1.Source) {
	flags.StringVar(
		source.URL,
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
		"name of the secret with credentials to access the source, e.g. git or registry credentials",
	)
	flags.StringVar(
		&source.BundleContainer.Image,
		SourceBundleImageFlag,
		"",
		"source bundle image location, e.g. ghcr.io/shipwright-io/sample-go/source-bundle:latest",
	)
	flags.Var(
		pruneOptionFlag{ref: source.BundleContainer.Prune},
		SourceBundlePruneFlag,
		fmt.Sprintf("source bundle prune option, either %s, or %s", buildv1alpha1.PruneNever, buildv1alpha1.PruneAfterPull),
	)
}

// strategyFlags flags for ".spec.strategy".
func strategyFlags(flags *pflag.FlagSet, strategy *buildv1alpha1.Strategy) {
	flags.StringVar(
		strategy.APIVersion,
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
	if prefix == "output" {
		flags.BoolVar(
			image.Insecure,
			fmt.Sprintf("%s-insecure", prefix),
			false,
			"flag to indicate an insecure container registry",
		)
	}
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
		buildRef.APIVersion,
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
		sa.Generate,
		ServiceAccountGenerateFlag,
		false,
		"generate a Kubernetes service-account for the build",
	)
}

// envFlags registers flags for adding corev1.EnvVars.
func envFlags(flags *pflag.FlagSet, envs *[]corev1.EnvVar) {
	flags.VarP(
		NewCoreEnvVarArrayValue(envs),
		"env",
		"e",
		"specify a key-value pair for an environment variable to set for the build container",
	)
}

// imageLabelsFlags registers flags for output image labels.
func imageLabelsFlags(flags *pflag.FlagSet, labels map[string]string) {
	flags.VarP(
		NewMapValue(labels),
		OutputImageLabelsFlag,
		"",
		"specify a set of key-value pairs that correspond to labels to set on the output image",
	)
}

// imageLabelsFlags registers flags for output image annotations.
func imageAnnotationsFlags(flags *pflag.FlagSet, annotations map[string]string) {
	flags.VarP(
		NewMapValue(annotations),
		OutputImageAnnotationsFlag,
		"",
		"specify a set of key-value pairs that correspond to annotations to set on the output image",
	)
}

func buildRetentionFlags(flags *pflag.FlagSet, buildRetention *buildv1alpha1.BuildRetention) {
	flags.UintVar(
		buildRetention.FailedLimit,
		RetentionFailedLimitFlag,
		65535,
		"number of failed BuildRuns to be kept",
	)
	flags.UintVar(
		buildRetention.SucceededLimit,
		RetentionSucceededLimitFlag,
		65535,
		"number of succeeded BuildRuns to be kept",
	)
	flags.DurationVar(
		&buildRetention.TTLAfterFailed.Duration,
		RetentionTTLAfterFailedFlag,
		time.Duration(0),
		"duration to delete a failed BuildRun after completion",
	)
	flags.DurationVar(
		&buildRetention.TTLAfterSucceeded.Duration,
		RetentionTTLAfterSucceededFlag,
		time.Duration(0),
		"duration to delete a succeeded BuildRun after completion",
	)
}

func buildRunRetentionFlags(flags *pflag.FlagSet, buildRunRetention *buildv1alpha1.BuildRunRetention) {
	flags.DurationVar(
		&buildRunRetention.TTLAfterFailed.Duration,
		RetentionTTLAfterFailedFlag,
		time.Duration(0),
		"duration to delete the BuildRun after it failed",
	)
	flags.DurationVar(
		&buildRunRetention.TTLAfterSucceeded.Duration,
		RetentionTTLAfterSucceededFlag,
		time.Duration(0),
		"duration to delete the BuildRun after it succeeded",
	)
}
