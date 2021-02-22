package stub

import (
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
)

// NewFakeClient creates a fake client with Shipwright's Build scheme.
func NewFakeClient() dynamic.Interface {
	scheme := runtime.NewScheme()
	buildv1alpha1.SchemeBuilder.AddToScheme(scheme)
	return fake.NewSimpleDynamicClient(scheme)
}
