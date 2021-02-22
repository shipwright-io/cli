package resource

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/shipwright-io/cli/pkg/shp/params"
	"github.com/shipwright-io/cli/pkg/shp/util"
)

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

func (r *Resource) Create(name string, obj interface{}) error {
	ri, err := r.getResourceInterface()
	if err != nil {
		return nil
	}

	return util.CreateObject(ri, name, r.gv.WithKind(r.kind), obj)
}

func (r *Resource) Delete(name string) error {
	ri, err := r.getResourceInterface()
	if err != nil {
		return nil
	}

	return util.DeleteObject(ri, name, r.gv.WithResource(r.resource))
}

func (r *Resource) List(result interface{}) error {
	ri, err := r.getResourceInterface()
	if err != nil {
		return nil
	}

	return util.ListObject(ri, result)
}

func (r *Resource) Get(name string, result interface{}) error {
	ri, err := r.getResourceInterface()
	if err != nil {
		return nil
	}

	return util.GetObject(ri, name, result)
}

func NewShpResource(p *params.Params, gv schema.GroupVersion, kind, resource string) *Resource {
	sr := &Resource{
		gv:       gv,
		kind:     kind,
		resource: resource,
		params:   p,
	}

	return sr
}
