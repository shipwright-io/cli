package stub

import (
	"log"

	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
)

// NewFakeClient creates a fake client with Shipwright's Build scheme.
func NewFakeClient() dynamic.Interface {
	scheme := runtime.NewScheme()
	if err := buildv1beta1.SchemeBuilder.AddToScheme(scheme); err != nil {
		log.Fatal(err)
	}
	return fake.NewSimpleDynamicClient(scheme)
}
