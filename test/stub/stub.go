package stub

import (
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BuildRunEmpty simple empty BuildRun instance.
func BuildRunEmpty() buildv1alpha1.BuildRun {
	return buildv1alpha1.BuildRun{}
}

// TestBuild returns instance of Build for testing purposes
func TestBuild(name, image, source string) *buildv1alpha1.Build {
	strategyKind := buildv1alpha1.ClusterBuildStrategyKind

	result := &buildv1alpha1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: buildv1alpha1.BuildSpec{
			StrategyRef: &buildv1alpha1.StrategyRef{
				Name: "buildah",
				Kind: &strategyKind,
			},
			Source: buildv1alpha1.GitSource{
				URL: source,
			},
			Output: buildv1alpha1.Image{
				ImageURL: image,
			},
		},
	}

	return result
}
