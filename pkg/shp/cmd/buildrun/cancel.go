package buildrun

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/shipwright-io/cli/pkg/shp/cmd/types"
	"github.com/shipwright-io/cli/pkg/shp/reactor"
)

var (
	// Long description for the "buildrun cancel" command
	buildRunCancelLongDescription = templates.LongDesc(`
		Cancels a BuildRun
	`)

	// Examples for using the "buildrun cancel" command
	buildRunCancelExamples = templates.Examples(`
		$ shp buildrun cancel my-build
	`)
)

// BuildRunCancelOptions stores data passed to the command via command line flags
type BuildRunCancelOptions struct {
	types.SharedOptions
	PodWatcher *reactor.PodWatcher

	Name string
}

// newBuildRunCancelCmd creates the "buildrun cancel" command
func newBuildRunCancelCmd(ctx context.Context, ioStreams *genericclioptions.IOStreams, clients *types.ClientSets, o *BuildRunCancelOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cancel <name>",
		Short:   "Cancel BuildRun",
		Long:    buildRunCancelLongDescription,
		Example: buildRunCancelExamples,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(args))
			cmdutil.CheckErr(o.Run())
		},
	}
	return cmd
}

// NewBuildRunCancelCmd is a wrapper for newBuildRunCancelCmd
func NewBuildRunCancelCmd(ctx context.Context, ioStreams *genericclioptions.IOStreams, clients *types.ClientSets) *cobra.Command {
	o := &BuildRunCancelOptions{
		SharedOptions: types.SharedOptions{
			Clients: clients,
			Context: ctx,
			Streams: ioStreams,
		},
	}

	return newBuildRunCancelCmd(ctx, ioStreams, clients, o)
}

// Complete processes any data that is needed before Run executes
func (o *BuildRunCancelOptions) Complete(args []string) error {
	// Guard against index out of bound errors
	if len(args) == 0 {
		return errors.New("argument list is empty")
	}

	o.Name = args[0]

	return nil
}

// Run executes the command logic
func (o *BuildRunCancelOptions) Run() error {
	br, err := o.Clients.ShipwrightClientSet.ShipwrightV1alpha1().BuildRuns(o.Clients.Namespace).Get(o.Context, o.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("an error occurred getting BuildRun %q: %s", o.Name, err.Error())
	}

	if br.IsDone() {
		return fmt.Errorf("failed to cancel BuildRun %q: execution has already finished", o.Name)
	}

	type patchStringValue struct {
		Op    string `json:"op"`
		Path  string `json:"path"`
		Value string `json:"value"`
	}
	payload := []patchStringValue{{
		Op:    "replace",
		Path:  "/spec/state",
		Value: buildv1alpha1.BuildRunStateCancel,
	}}
	var data []byte
	if data, err = json.Marshal(payload); err != nil {
		return err
	}
	if _, err = o.Clients.ShipwrightClientSet.ShipwrightV1alpha1().BuildRuns(o.Clients.Namespace).Patch(o.Context, o.Name, ktypes.JSONPatchType, data, metav1.PatchOptions{}); err != nil {
		return err
	}

	fmt.Fprintf(o.Streams.Out, "BuildRun %q has been canceled\n", o.Name)

	return nil
}
