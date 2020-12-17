package stub

import (
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
)

// BuildRunEmpty simple empty BuildRun instance.
func BuildRunEmpty() buildv1alpha1.BuildRun {
	return buildv1alpha1.BuildRun{}
}
