package flags

import (
	"fmt"
	"time"

	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	"github.com/spf13/pflag"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// BuildrefNameFlag command-line flag.
	BuildrefNameFlag = "buildref-name"
	// BuilderImageFlag command-line flag.
	BuilderImageFlag = "builder-image"
	// DockerfileFlag command-line flag.
	DockerfileFlag = "dockerfile"
	// EnvFlag command-line flag.
	EnvFlag = "env"
	// SourceURLFlag command-line flag.
	SourceURLFlag = "source-git-url"
	// SourceRevisionFlag command-line flag.
	SourceRevisionFlag = "source-git-revision"
	// SourceContextDirFlag command-line flag.
	SourceContextDirFlag = "source-context-dir"
	// SourceCredentialsSecretFlag command-line flag.
	SourceCredentialsSecretFlag = "source-git-clone-secret" // #nosec G101
	// SourceBundleImageFlag command-line flag
	SourceOCIArtifactImageFlag = "source-oci-artifact-image"
	// SourceBundlePruneFlag command-line flag
	SourceOCIArtifactPruneFlag = "source-oci-artifact-prune"
	// SourceOCIArtifactPullSecretFlag command-line flag
	SourceOCIArtifactPullSecretFlag = "source-oci-artifact-pull-secret" // #nosec G101
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
	// ParameterValueFlag command-line flag.
	ParamValueFlag = "param-value"
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
	// NodeSelectorFlag command-line flag.
	NodeSelectorFlag = "node-selector"
)

// sourceFlags flags for ".spec.source"
func sourceFlags(flags *pflag.FlagSet, source *buildv1beta1.Source) {
	flags.StringVar(
		&source.Git.URL,
		SourceURLFlag,
		"",
		"git repository source URL",
	)
	flags.StringVar(
		&source.Git.URL,
		"source-url",
		"",
		"alias for source-git-url",
	)
	flags.MarkDeprecated("source-url", fmt.Sprintf("please use --%s instead", SourceURLFlag))

	flags.StringVar(
		source.Git.Revision,
		SourceRevisionFlag,
		"",
		"git repository source revision",
	)
	flags.StringVar(
		source.Git.Revision,
		"source-revision",
		"",
		"alias for source-git-revision",
	)
	flags.MarkDeprecated("source-revision", fmt.Sprintf("please use --%s instead", SourceRevisionFlag))

	flags.StringVar(
		source.ContextDir,
		SourceContextDirFlag,
		"",
		"use a inner directory as context directory",
	)

	flags.StringVar(
		source.Git.CloneSecret,
		SourceCredentialsSecretFlag,
		"",
		"name of the secret with credentials to access the git source, e.g. git credentials",
	)
	flags.StringVar(
		source.Git.CloneSecret,
		"source-credentials-secret",
		"",
		"name of the secret with credentials to access the source, e.g. credentials",
	)
	flags.MarkDeprecated("source-credentials-secret", fmt.Sprintf("please use --%s instead", SourceCredentialsSecretFlag))

	flags.StringVar(
		&source.OCIArtifact.Image,
		SourceOCIArtifactImageFlag,
		"",
		"source OCI artifact image reference, e.g. ghcr.io/shipwright-io/sample-go/source-bundle:latest",
	)
	flags.StringVar(
		&source.OCIArtifact.Image,
		"source-bundle-image",
		"",
		"source bundle image location, e.g. ghcr.io/shipwright-io/sample-go/source-bundle:latest",
	)
	flags.MarkDeprecated("source-bundle-image", fmt.Sprintf("please use --%s instead", SourceOCIArtifactImageFlag))

	flags.StringVar(
		source.OCIArtifact.PullSecret,
		SourceOCIArtifactPullSecretFlag,
		"",
		"name of the secret with credentials to access the OCI artifact image, e.g. registry credentials",
	)

	flags.Var(
		pruneOptionFlag{ref: source.OCIArtifact.Prune},
		SourceOCIArtifactPruneFlag,
		fmt.Sprintf("source OCI artifact image prune option, either %s, or %s", buildv1beta1.PruneNever, buildv1beta1.PruneAfterPull),
	)
	flags.Var(
		pruneOptionFlag{ref: source.OCIArtifact.Prune},
		"source-bundle-prune",
		fmt.Sprintf("source bundle prune option, either %s, or %s", buildv1beta1.PruneNever, buildv1beta1.PruneAfterPull),
	)
	flags.MarkDeprecated("source-bundle-prune", fmt.Sprintf("please use --%s instead", SourceOCIArtifactPruneFlag))
}

// strategyFlags flags for ".spec.strategy".
func strategyFlags(flags *pflag.FlagSet, strategy *buildv1beta1.Strategy) {
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
func imageFlags(flags *pflag.FlagSet, prefix string, image *buildv1beta1.Image) {
	flags.StringVar(
		&image.Image,
		fmt.Sprintf("%s-image", prefix),
		"",
		"image employed during the building process",
	)
	flags.StringVar(
		image.PushSecret,
		fmt.Sprintf("%s-image-push-secret", prefix),
		"",
		"name of the secret with output image push credentials",
	)
	flags.StringVar(
		image.PushSecret,
		fmt.Sprintf("%s-credentials-secret", prefix),
		"",
		"name of the secret with output image push credentials",
	)
	flags.MarkDeprecated(fmt.Sprintf("%s-credentials-secret", prefix), fmt.Sprintf("please use --%s-image-push-secret instead", prefix))

	if prefix == "output" {
		flags.BoolVar(
			image.Insecure,
			fmt.Sprintf("%s-insecure", prefix),
			false,
			"flag to indicate an insecure container registry",
		)
	}
}

// dockerfileFlags register dockerfile flag as an environment variable.
func dockerfileFlags(flags *pflag.FlagSet, dockerfile *string) {
	flags.StringVar(
		dockerfile,
		DockerfileFlag,
		"",
		"path to dockerfile relative to repository",
	)
	flags.MarkDeprecated("dockerfile", "dockerfile parameter is deprecated")
}

// builderImageFlag register builder-image flag as an environment variable..
func builderImageFlag(flags *pflag.FlagSet, builderImage *string) {
	flags.StringVar(
		builderImage,
		BuilderImageFlag,
		"",
		"path to dockerfile relative to repository",
	)
	flags.MarkDeprecated("builder-image", "builder-image flag is deprecated, and will be removed in a future release. Use an appropriate parameter for the build strategy instead.")
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
func buildRefFlags(flags *pflag.FlagSet, buildRef *buildv1beta1.ReferencedBuild) {
	flags.StringVar(
		buildRef.Name,
		BuildrefNameFlag,
		"",
		"name of build resource to reference",
	)
}

// serviceAccountFlags register flags for BuildRun's spec.serviceAccount attribute.
func serviceAccountFlags(flags *pflag.FlagSet, sa *string) {
	flags.StringVar(
		sa,
		ServiceAccountNameFlag,
		"",
		"Kubernetes service-account name",
	)
	var ignore bool
	flags.BoolVar(
		&ignore,
		ServiceAccountGenerateFlag,
		false,
		"generate a Kubernetes service-account for the build",
	)
	flags.MarkDeprecated("sa-generate", fmt.Sprintf("this flag has no effect, please use --%s for service account", ServiceAccountNameFlag))

}

// buildNodeSelectorFlags registers flags for adding BuildSpec.NodeSelector
func buildNodeSelectorFlags(flags *pflag.FlagSet, nodeSelectorLabels map[string]string) {
	flags.Var(NewMapValue(nodeSelectorLabels), NodeSelectorFlag, "set of key-value pairs that correspond to labels of a node to match")
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

// parameterValueFlag registers flags for adding BuildSpec.ParamValues
func paramValueFlag(flags *pflag.FlagSet, paramValue *[]buildv1beta1.ParamValue) {
	flags.VarP(
		NewParamArrayValue(paramValue),
		ParamValueFlag,
		"",
		"set of key-value pairs to pass as parameters to the buildStrategy",
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

func buildRetentionFlags(flags *pflag.FlagSet, buildRetention *buildv1beta1.BuildRetention) {
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

func buildRunRetentionFlags(flags *pflag.FlagSet, buildRunRetention *buildv1beta1.BuildRunRetention) {
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
