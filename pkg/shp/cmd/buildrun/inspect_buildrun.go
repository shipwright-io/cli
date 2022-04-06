package buildrun

import (
	"context"
	"fmt"
	"time"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	buildclientset "github.com/shipwright-io/build/pkg/client/clientset/versioned"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// InspectBuildRun tries to retrieve the informed BuildRun object to identify its status and log it
// on the ioStreams instance.
func InspectBuildRun(
	ctx context.Context,
	buildClientset buildclientset.Interface,
	name types.NamespacedName,
	ioStreams *genericclioptions.IOStreams,
) error {
	var br *buildv1alpha1.BuildRun
	err := wait.PollImmediate(1*time.Second, 15*time.Second, func() (done bool, err error) {
		br, err = buildClientset.ShipwrightV1alpha1().
			BuildRuns(name.Namespace).
			Get(ctx, name.Name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			return true, err
		}
		return br.IsDone(), nil
	})
	if err != nil {
		return err
	}
	switch {
	case br == nil || (br != nil && br.DeletionTimestamp != nil):
		fmt.Fprintf(ioStreams.ErrOut, "BuildRun %q has been deleted.\n", name)
	case br != nil && br.IsCanceled():
		fmt.Fprintf(ioStreams.ErrOut, "BuildRun %q has been canceled.\n", name)
	case br != nil && !br.HasStarted():
		fmt.Fprintf(ioStreams.ErrOut, "BuildRun %q has been marked as failed.\n", name)
	}
	return nil
}
