package resource

import (
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"

	v1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	corev1 "k8s.io/api/core/v1"

	"github.com/shipwright-io/cli/pkg/shp/params"
)

// GetBuildRunResource returns Resource for working with BuildRun objects
func GetBuildRunResource(p *params.Params) *Resource {
	return newResource(
		p,
		buildv1alpha1.SchemeGroupVersion,
		"BuildRun",
		"buildruns",
	)
}

// GetPodResource returns Resource for working with Pod objects
func GetPodResource(p *params.Params) *Resource {
	return newResource(
		p,
		corev1.SchemeGroupVersion,
		"Pod",
		"pods",
	)
}

// GetBuildResource returns Resource for working with Build objects
func GetBuildResource(p *params.Params) *Resource {
	return newResource(
		p,
		buildv1alpha1.SchemeBuilder.GroupVersion,
		"Build",
		"builds",
	)
}

// GetTaskRunResource returns a Resource for working with taskRun objects
func GetTaskRunResource(p *params.Params) *Resource {
	return newResource(
		p,
		v1beta1.SchemeGroupVersion,
		"taskRun",
		"taskruns",
	)
}
