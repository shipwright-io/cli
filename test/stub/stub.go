package stub

import (
	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BuildRunEmpty simple empty BuildRun instance.
func BuildRunEmpty() buildv1beta1.BuildRun {
	return buildv1beta1.BuildRun{}
}

// TestBuild returns instance of Build for testing purposes
func TestBuild(name, image, source string) *buildv1beta1.Build {
	strategyKind := buildv1beta1.ClusterBuildStrategyKind

	result := &buildv1beta1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: buildv1beta1.BuildSpec{
			Strategy: buildv1beta1.Strategy{
				Name: "buildah",
				Kind: &strategyKind,
			},
			Source: &buildv1beta1.Source{
				Type: buildv1beta1.GitType,
				Git: &buildv1beta1.Git{
					URL: source,
				},
			},
			Output: buildv1beta1.Image{
				Image: image,
			},
		},
	}

	return result
}
