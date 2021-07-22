package resource

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"

	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/util"
)

// Resource is a wrapper around dynamic kubernetes client that
// provides means to work with objects stored in kubernetes
type Resource struct {
	gv       schema.GroupVersion
	kind     string
	resource string

	params *params.Params

	resourceInterface dynamic.ResourceInterface
}

func (r *Resource) getResourceInterface() (dynamic.ResourceInterface, error) {
	if r.resourceInterface != nil {
		return r.resourceInterface, nil
	}

	client, err := r.params.Client()
	if err != nil {
		return nil, err
	}

	r.resourceInterface = client.Resource(r.gv.WithResource(r.resource)).Namespace(r.params.Namespace())
	return r.resourceInterface, nil
}

// Create creates the object
func (r *Resource) Create(ctx context.Context, name string, obj interface{}) error {
	ri, err := r.getResourceInterface()
	if err != nil {
		return nil
	}

	return util.CreateObject(ctx, ri, name, r.gv.WithKind(r.kind), obj)
}

// Update execute update against informed object.
func (r *Resource) Update(ctx context.Context, name string, obj interface{}) error {
	ri, err := r.getResourceInterface()
	if err != nil {
		return err
	}
	return util.UpdateObject(ctx, ri, name, r.gv.WithKind(r.kind), obj)
}

func (r *Resource) Patch(ctx context.Context, name, op, path, value string) error {
	ri, err := r.getResourceInterface()
	if err != nil {
		return err
	}

	return util.PatchObject(ctx, ri, name, op, path, value)
}

// Delete deletes the object identified by name
func (r *Resource) Delete(ctx context.Context, name string) error {
	ri, err := r.getResourceInterface()
	if err != nil {
		return nil
	}

	return util.DeleteObject(ctx, ri, name, r.gv.WithResource(r.resource))
}

// List returns list of object type
func (r *Resource) List(ctx context.Context, result interface{}) error {
	return r.ListWithOptions(ctx, result, v1.ListOptions{})
}

// ListWithOptions returns list of object type narrowed down by ListOptions
func (r *Resource) ListWithOptions(ctx context.Context, result interface{}, options v1.ListOptions) error {
	ri, err := r.getResourceInterface()
	if err != nil {
		return nil
	}

	return util.ListObjectWithOptions(ctx, ri, result, options)
}

// Get returns the object identified by name
func (r *Resource) Get(ctx context.Context, name string, result interface{}) error {
	ri, err := r.getResourceInterface()
	if err != nil {
		return nil
	}

	return util.GetObject(ctx, ri, name, result)
}

// Watch returns a watch for informed list options.
func (r *Resource) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	ri, err := r.getResourceInterface()
	if err != nil {
		return nil, err
	}
	return ri.Watch(ctx, opts)
}

func newResource(p *params.Params, gv schema.GroupVersion, kind, resource string) *Resource {
	r := &Resource{
		gv:       gv,
		kind:     kind,
		resource: resource,
		params:   p,
	}

	return r
}
