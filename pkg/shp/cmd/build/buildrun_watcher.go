package build

import (
	"context"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/cli/pkg/shp/cmd/buildrun"
	"github.com/shipwright-io/cli/pkg/shp/resource"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

//TODO hope to create a similar helper method on BuildRun in api when do cancel build
func isDone(br *buildv1alpha1.BuildRun) bool {
	c := br.Status.GetCondition(buildv1alpha1.Succeeded)
	return c != nil && c.GetStatus() != corev1.ConditionUnknown
}

func getBuildRunWatcherAddModFunc() func(br *buildv1alpha1.BuildRun) (*buildv1alpha1.BuildRun, error) {
	f := func(br *buildv1alpha1.BuildRun) (*buildv1alpha1.BuildRun, error) {
		if isDone(br) {
			return br, nil
		}
		if br.Status.LatestTaskRunRef != nil {
			return br, nil
		}
		return nil, nil
	}
	return f
}

func getBuildRunWatcherSkipFunc(buildRunName string) func(br *buildv1alpha1.BuildRun) bool {
	return func(br *buildv1alpha1.BuildRun) bool {
		if br.Name != buildRunName {
			return true
		}
		return false
	}
}

func getBuildRunWatcherDeleteFunc(buildRunName string) func(br *buildv1alpha1.BuildRun) (*buildv1alpha1.BuildRun, error) {
	return func(br *buildv1alpha1.BuildRun) (*buildv1alpha1.BuildRun, error) {
		return br, nil
	}
}

func waitForBuildRunToHaveTaskRun(ctx context.Context, buildRunName string, brr *resource.Resource, ioStreams *genericclioptions.IOStreams) (*buildv1alpha1.BuildRun, error) {
	brw, err := buildrun.NewBuildRunWatcher(ctx, brr, &v1.ListOptions{}, ioStreams)
	if err != nil {
		return nil, err
	}
	f := getBuildRunWatcherAddModFunc()
	brw.WithSkipBuildRunFunc(getBuildRunWatcherSkipFunc(buildRunName)).
		WithOnBuildRunDeletedFunc(getBuildRunWatcherDeleteFunc(buildRunName)).
		WithOnBuildRunAddedFunc(f).
		WithOnBuildRunModifiedFunc(f)

	return brw.Start()
}
