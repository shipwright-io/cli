package params

import (
	"math"
	"strings"
	"time"

	buildclientset "github.com/shipwright-io/build/pkg/client/clientset/versioned"
	"github.com/spf13/pflag"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

// Params is a place for Shipwright CLI to store its runtime parameters including configured dynamic
// client and global flags.
type Params struct {
	clientset    kubernetes.Interface
	shpClientset buildclientset.Interface

	configFlags *genericclioptions.ConfigFlags
	namespace   string
}

// AddFlags accepts flags and adds program global flags to it
func (p *Params) AddFlags(flags *pflag.FlagSet) {
	p.configFlags.AddFlags(flags)
}

// ClientSet returns a kubernetes clientset.
func (p *Params) ClientSet() (kubernetes.Interface, error) {
	if p.clientset != nil {
		return p.clientset, nil
	}

	clientConfig := p.configFlags.ToRawKubeConfigLoader()
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	p.namespace, _, err = clientConfig.Namespace()
	if err != nil {
		return nil, err
	}

	p.clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return p.clientset, nil
}

// RequestTimeout returns the setting from k8s --request-timeout param
func (p *Params) RequestTimeout() (time.Duration, error) {
	if p.configFlags.Timeout == nil {
		return math.MaxInt64, nil
	}
	// 0 or empty also mean no timeout
	to := strings.TrimSpace(*p.configFlags.Timeout)
	if len(to) == 0 || to == "0" || strings.HasPrefix(to, "0") {
		return math.MaxInt64, nil
	}
	return time.ParseDuration(*p.configFlags.Timeout)
}

// ShipwrightClientSet returns a Shipwright Clientset
func (p *Params) ShipwrightClientSet() (buildclientset.Interface, error) {
	if p.shpClientset != nil {
		return p.shpClientset, nil
	}
	clientConfig := p.configFlags.ToRawKubeConfigLoader()
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	p.namespace, _, err = clientConfig.Namespace()
	if err != nil {
		return nil, err
	}
	p.shpClientset, err = buildclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return p.shpClientset, nil
}

// Namespace returns kubernetes namespace with alle the overrides
// from command line and kubernetes config
func (p *Params) Namespace() string {
	return p.namespace
}

// NewParams creates a new instance of ShipwrightParams and returns it as
// an interface value
func NewParams() *Params {
	p := &Params{}
	p.configFlags = genericclioptions.NewConfigFlags(true)

	return p
}

// NewParamsForTest creates an instance of Params for testing purpose
func NewParamsForTest(clientset kubernetes.Interface,
	shpClientset buildclientset.Interface,
	configFlags *genericclioptions.ConfigFlags,
	namespace string) *Params {
	return &Params{
		clientset:    clientset,
		shpClientset: shpClientset,
		configFlags:  configFlags,
		namespace:    namespace,
	}
}
