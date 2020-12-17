package buildrun

import (
	"fmt"

	"github.com/otaviof/shp/pkg/shp/flags"
	"github.com/otaviof/shp/pkg/shp/util"
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

// BuildRun implements SubCommand interface, focusing on creating BuildRun resources.
type BuildRun struct {
	cmd  *cobra.Command              // instantiated on new
	spec *buildv1alpha1.BuildRunSpec // directly using flag values
	verb util.APIVerb                // action to run against resource

	name string // resource name, filled during complete
}

// TODO: move constant into the operator itself;

// BuildRunKind BuildRun API resource kind.
const BuildRunKind = "BuildRun"

// Complete the current BuildRun instance by validating and organizing args into attributes. When a
// "build-run" only the name is employed, while on a "build" the name is used for build-ref name.
func (b *BuildRun) Complete(client dynamic.Interface, ns string, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("not enough arguments, informed: '%v'", args)
	}
	noun := args[0]
	name := args[1]
	switch noun {
	case "build-run":
		b.name = name
	case "build":
		// FIXME: decide upon how to name build-run resources when the "run build" is issued;
		b.name = "should-be-random-name"
		b.spec.BuildRef.Name = name
	default:
		return fmt.Errorf("unknown noun '%s' in command-line parameters", noun)
	}
	return nil
}

// Validate apply validation logic against instantiated BuildRun and its spec.
func (b *BuildRun) Validate() error {
	// TODO: write the validation routine for BuildRun resources;
	return nil
}

// Run execute the primary logic, it creates a build-run resource in the cluster.
func (b *BuildRun) Run(client dynamic.Interface, ns string) error {
	u, err := util.ToUnstructured(b.name, BuildRunKind, &buildv1alpha1.BuildRun{Spec: *b.spec})
	if err != nil {
		return err
	}

	rs := buildRunResourceClient(client, ns)
	switch b.verb {
	case util.Create:
		_, err = rs.Create(u, metav1.CreateOptions{})
	case util.Delete:
		err = rs.Delete(b.name, &metav1.DeleteOptions{})
	case util.Update:
		// TODO: write the update routine
	default:
		return fmt.Errorf("%w: informed verb is '%s'", util.ErrUnknownVerb, b.verb)
	}
	return err
}

// Cmd share cobra Command instance.
func (b *BuildRun) Cmd() *cobra.Command {
	return b.cmd
}

// newBuildRun creates a BuildRun instance with informed variables and wire up sub-command flags.
func newBuildRun(cmd *cobra.Command, verb util.APIVerb) *BuildRun {
	return &BuildRun{
		cmd:  cmd,
		verb: verb,
		spec: flags.BuildRunSpecFlags(cmd.PersistentFlags()),
	}
}
