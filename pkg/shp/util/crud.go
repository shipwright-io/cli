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

// CreateObject creates the object with dynamic client
func CreateObject(ctx context.Context, resource dynamic.ResourceInterface, name string, gvk schema.GroupVersionKind, obj interface{}) error {
	u, err := toUnstructured(name, gvk, obj)
	if err != nil {
		return err
	}

	result, err := resource.Create(ctx, u, v1.CreateOptions{})
	if err != nil {
		return err
	}
	return fromUnstructured(result.Object, obj)
}

// UpdateObject updates informed object reference.
func UpdateObject(ctx context.Context, resource dynamic.ResourceInterface, name string, gvk schema.GroupVersionKind, obj interface{}) error {
	u, err := toUnstructured(name, gvk, obj)
	if err != nil {
		return err
	}

	result, err := resource.Update(ctx, u, v1.UpdateOptions{})
	if err != nil {
		return err
	}
	return fromUnstructured(result.Object, obj)
}

// GetObject returns the object using dynamic client
func GetObject(ctx context.Context, resource dynamic.ResourceInterface, name string, obj interface{}) error {
	u, err := resource.Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		return err
	}

	return fromUnstructured(u.UnstructuredContent(), obj)
}

// DeleteObject deletes the object with dynamic client
func DeleteObject(ctx context.Context, resource dynamic.ResourceInterface, name string, gvr schema.GroupVersionResource) error {
	return resource.Delete(context.TODO(), name, v1.DeleteOptions{})
}

// ListObject lists objects using dynamic client
func ListObject(ctx context.Context, resource dynamic.ResourceInterface, result interface{}) error {
	return ListObjectWithOptions(ctx, resource, result, v1.ListOptions{})
}

// ListObjectWithOptions lists objects using dynamic client narrowed down by ListOptions
func ListObjectWithOptions(ctx context.Context, resource dynamic.ResourceInterface, result interface{}, options v1.ListOptions) error {
	u, err := resource.List(context.TODO(), options)
	if err != nil {
		return err
	}

	return fromUnstructured(u.UnstructuredContent(), result)
}
