package util

import (
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ToUnstructured converts informed object to unstructured, using informed name and kind.
func ToUnstructured(name, kind string, obj interface{}) (*unstructured.Unstructured, error) {
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	u := &unstructured.Unstructured{Object: data}
	gvk := buildv1alpha1.SchemeGroupVersion.WithKind(kind)
	u.SetGroupVersionKind(gvk)
	u.SetName(name)
	return u, nil
}
