package buildrun

import (
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewBuildRun creates and returns BuildRun kubernetes object for given arguments
func NewBuildRun(build *buildv1alpha1.Build, prefix string) *buildv1alpha1.BuildRun {
	return &buildv1alpha1.BuildRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: prefix + "-",
		},
		Spec: buildv1alpha1.BuildRunSpec{
			BuildRef: &buildv1alpha1.BuildRef{
				Name:       build.Name,
				APIVersion: buildv1alpha1.SchemeGroupVersion.Version,
			},
		},
	}
}
