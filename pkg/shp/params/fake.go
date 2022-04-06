package params

import (
	"fmt"
	"time"

	buildclientset "github.com/shipwright-io/build/pkg/client/clientset/versioned"
	shpfake "github.com/shipwright-io/build/pkg/client/clientset/versioned/fake"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/testing"
)

type FakeParams struct {
	timeout           time.Duration
	coreObjects       []runtime.Object
	shipwrightObjects []runtime.Object

	clientset           *fake.Clientset    // kubernetes clientset
	shipwrightClientset *shpfake.Clientset // shipwright clientset
}

var _ Interface = &FakeParams{}

// generateNameReactor intercepts object creation action in order to simulate the "generateName" in
// Kubernetes, where a object created with the directive receives a random name suffix.
func generateNameReactor(action testing.Action) (bool, runtime.Object, error) {
	createAction, ok := action.(testing.CreateAction)
	if !ok {
		panic(fmt.Errorf("action is not a create action: %+v", action))
	}

	obj := createAction.GetObject()
	objMeta, err := meta.Accessor(obj)
	if err != nil {
		panic(err)
	}

	if objMeta.GetName() == "" {
		genName := objMeta.GetGenerateName()
		if genName == "" {
			panic(fmt.Errorf("object does not have a name or generateName: '%#v'", obj))
		}
		suffix := rand.String(5)
		objMeta.SetName(fmt.Sprintf("%s%s", genName, suffix))
	}
	return false, nil, nil
}

func (p *FakeParams) AddFlags(*pflag.FlagSet) {}

func (p *FakeParams) RESTConfig() (*rest.Config, error) {
	return nil, nil
}

func (p *FakeParams) ClientSet() (kubernetes.Interface, error) {
	if p.clientset != nil {
		return p.clientset, nil
	}

	p.clientset = fake.NewSimpleClientset(p.coreObjects...)
	return p.clientset, nil
}

func (p *FakeParams) RequestTimeout() (time.Duration, error) {
	return p.timeout, nil
}

func (p *FakeParams) ShipwrightClientSet() (buildclientset.Interface, error) {
	if p.shipwrightClientset != nil {
		return p.shipwrightClientset, nil
	}

	p.shipwrightClientset = shpfake.NewSimpleClientset(p.shipwrightObjects...)
	p.shipwrightClientset.PrependReactor("create", "buildruns", generateNameReactor)
	return p.shipwrightClientset, nil
}

func (p *FakeParams) Namespace() string {
	return metav1.NamespaceDefault
}

func NewFakeParams(
	timeout time.Duration,
	coreObjects []runtime.Object,
	shipwrightObjects []runtime.Object,
) *FakeParams {
	return &FakeParams{
		timeout:           timeout,
		coreObjects:       coreObjects,
		shipwrightObjects: shipwrightObjects,
	}
}
