package params

import (
	"time"

	buildclientset "github.com/shipwright-io/build/pkg/client/clientset/versioned"
	"github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Interface interface {
	AddFlags(*pflag.FlagSet)
	RESTConfig() (*rest.Config, error)
	ClientSet() (kubernetes.Interface, error)
	RequestTimeout() (time.Duration, error)
	ShipwrightClientSet() (buildclientset.Interface, error)
	Namespace() string
}
