package buildrun

import (
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"k8s.io/client-go/dynamic"
)

func buildRunResourceClient(client dynamic.Interface, ns string) dynamic.ResourceInterface {
	gvr := buildv1alpha1.SchemeGroupVersion.WithResource("buildruns")
	return client.Resource(gvr).Namespace(ns)
}
