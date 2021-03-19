package util

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// ToUnstructured converts informed object to unstructured, using informed name and kind.
func toUnstructured(name string, gvk schema.GroupVersionKind, obj interface{}) (*unstructured.Unstructured, error) {
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	u := &unstructured.Unstructured{Object: data}
	u.SetGroupVersionKind(gvk)
	u.SetName(name)
	return u, nil
}

func fromUnstructured(u map[string]interface{}, obj interface{}) error {
	return runtime.DefaultUnstructuredConverter.FromUnstructured(u, obj)
}

func CreateObject(resource dynamic.ResourceInterface, name string, gvk schema.GroupVersionKind, obj interface{}) error {
	u, err := toUnstructured(name, gvk, obj)
	if err != nil {
		return err
	}

	_, err = resource.Create(context.TODO(), u, v1.CreateOptions{})
	return err
}

func GetObject(resource dynamic.ResourceInterface, name string, obj interface{}) error {
	u, err := resource.Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		return err
	}

	return fromUnstructured(u.UnstructuredContent(), obj)
}

func DeleteObject(resource dynamic.ResourceInterface, name string, gvr schema.GroupVersionResource) error {
	return resource.Delete(context.TODO(), name, v1.DeleteOptions{})
}

func ListObject(resource dynamic.ResourceInterface, result interface{}) error {
	u, err := resource.List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return err
	}

	return fromUnstructured(u.UnstructuredContent(), result)
}
