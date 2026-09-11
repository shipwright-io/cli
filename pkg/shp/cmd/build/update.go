package build // nolint:revive

import (
	"fmt"

	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	"github.com/spf13/cobra"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd/runner"
	"github.com/shipwright-io/cli/pkg/shp/flags"
	"github.com/shipwright-io/cli/pkg/shp/params"
)

// UpdateCommand contains data provided by user to update an existing Build
type UpdateCommand struct {
	cmd *cobra.Command

	name         string
	buildSpec    *buildv1beta1.BuildSpec
	dockerfile   *string
	builderImage *string
}

func updateCmd() runner.SubCommand {
	cmd := &cobra.Command{
		Use:   "update <name> [flags]",
		Short: "Update an existing Build",
		Args:  cobra.ExactArgs(1),
	}

	buildSpecFlags, dockerfileFlag, builderImageFlag := flags.BuildSpecFromFlags(cmd.Flags())

	return &UpdateCommand{
		cmd:          cmd,
		buildSpec:    buildSpecFlags,
		dockerfile:   dockerfileFlag,
		builderImage: builderImageFlag,
	}
}

// Cmd returns cobra command object
func (c *UpdateCommand) Cmd() *cobra.Command {
	return c.cmd
}

// Complete fills object with user input data
func (c *UpdateCommand) Complete(_ *params.Params, _ *genericclioptions.IOStreams, args []string) error {
	c.name = args[0]
	return nil
}

// Validate checks user input data
func (c *UpdateCommand) Validate() error {
	if c.name == "" {
		return fmt.Errorf("name must be provided")
	}
	return nil
}

// Run executes the update of an existing Build instance
func (c *UpdateCommand) Run(params *params.Params, ioStreams *genericclioptions.IOStreams) error {
	clientset, err := params.ShipwrightClientSet()
	if err != nil {
		return err
	}

	ns := params.Namespace()
	build, err := clientset.ShipwrightV1beta1().Builds(ns).Get(c.cmd.Context(), c.name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			fmt.Fprintf(ioStreams.Out, "Build '%s' not found in namespace '%s'.\n", c.name, ns)
			return nil
		}
		return err
	}

	// Update Git Source URL if flag was explicitly provided
	if c.cmd.Flags().Changed(flags.SourceGitURLFlag) && c.buildSpec.Source != nil && c.buildSpec.Source.Git != nil {
		if build.Spec.Source == nil {
			build.Spec.Source = &buildv1beta1.Source{}
		}
		if build.Spec.Source.Git == nil {
			build.Spec.Source.Git = &buildv1beta1.Git{}
		}
		build.Spec.Source.Git.URL = c.buildSpec.Source.Git.URL
		build.Spec.Source.Type = buildv1beta1.GitType
	}

	// Update Git Revision if flag was explicitly provided
	if c.cmd.Flags().Changed(flags.SourceGitRevisionFlag) && c.buildSpec.Source != nil && c.buildSpec.Source.Git != nil {
		if build.Spec.Source == nil {
			build.Spec.Source = &buildv1beta1.Source{}
		}
		if build.Spec.Source.Git == nil {
			build.Spec.Source.Git = &buildv1beta1.Git{}
		}
		build.Spec.Source.Git.Revision = c.buildSpec.Source.Git.Revision
	}

	// Update ContextDir if flag was explicitly provided
	if c.cmd.Flags().Changed(flags.SourceContextDirFlag) && c.buildSpec.Source != nil {
		if build.Spec.Source == nil {
			build.Spec.Source = &buildv1beta1.Source{}
		}
		build.Spec.Source.ContextDir = c.buildSpec.Source.ContextDir
	}

	// Update Strategy Name if flag was explicitly provided
	if c.cmd.Flags().Changed(flags.StrategyNameFlag) {
		build.Spec.Strategy.Name = c.buildSpec.Strategy.Name
	}

	// Update Strategy Kind if flag was explicitly provided
	if c.cmd.Flags().Changed(flags.StrategyKindFlag) {
		build.Spec.Strategy.Kind = c.buildSpec.Strategy.Kind
	}

	// Update Output Image if flag was explicitly provided
	if c.cmd.Flags().Changed(flags.OutputImageFlag) {
		build.Spec.Output.Image = c.buildSpec.Output.Image
	}

	// Update dockerfile param if changed
	if c.dockerfile != nil && *c.dockerfile != "" && c.cmd.Flags().Changed(flags.DockerfileFlag) {
		updateOrAppendParam(&build.Spec.ParamValues, "dockerfile", *c.dockerfile)
	}

	// Update builder-image param if changed
	if c.builderImage != nil && *c.builderImage != "" && c.cmd.Flags().Changed(flags.BuilderImageFlag) {
		updateOrAppendParam(&build.Spec.ParamValues, "builder-image", *c.builderImage)
	}

	flags.SanitizeBuildSpec(&build.Spec)

	if _, err := clientset.ShipwrightV1beta1().Builds(ns).Update(c.cmd.Context(), build, metav1.UpdateOptions{}); err != nil {
		return err
	}

	fmt.Fprintf(ioStreams.Out, "Updated build %q\n", c.name)
	return nil
}

func updateOrAppendParam(paramValues *[]buildv1beta1.ParamValue, name string, val string) {
	for i, p := range *paramValues {
		if p.Name == name {
			v := val
			(*paramValues)[i].SingleValue = &buildv1beta1.SingleValue{Value: &v}
			return
		}
	}
	v := val
	*paramValues = append(*paramValues, buildv1beta1.ParamValue{
		Name: name,
		SingleValue: &buildv1beta1.SingleValue{Value: &v},
	})
}
