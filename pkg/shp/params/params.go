package params

import (
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/scheme"

	buildclientset "github.com/shipwright-io/build/pkg/client/clientset/versioned"

	"github.com/spf13/pflag"
)

// Params is a place for Shipwright CLI to store its runtime parameters including configured dynamic
// client and global flags.
type Params struct {
	clientset      kubernetes.Interface     // kubernetes api-client, global instance
	buildClientset buildclientset.Interface // shipwright api-client, global instance

	configFlags *genericclioptions.ConfigFlags
	namespace   string
}

var _ Interface = &Params{}

// AddFlags accepts flags and adds program global flags to it
func (p *Params) AddFlags(flags *pflag.FlagSet) {
	p.configFlags.AddFlags(flags)
}

// RESTConfig returns the rest configuration based on local flags.
func (p *Params) RESTConfig() (*rest.Config, error) {
	clientConfig := p.configFlags.ToRawKubeConfigLoader()
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	p.namespace, _, err = clientConfig.Namespace()
	if err != nil {
		return nil, err
	}

	restConfig.APIPath = "/api"
	restConfig.GroupVersion = &corev1.SchemeGroupVersion
	restConfig.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{
		CodecFactory: scheme.Codecs,
	}
	return restConfig, nil
}

// ClientSet returns a kubernetes clientset.
func (p *Params) ClientSet() (kubernetes.Interface, error) {
	if p.clientset != nil {
		return p.clientset, nil
	}

	restConfig, err := p.RESTConfig()
	if err != nil {
		return nil, err
	}
	p.clientset, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return p.clientset, nil
}

// RequestTimeout returns the setting from k8s --request-timeout param
func (p *Params) RequestTimeout() (time.Duration, error) {
	timeoutDefault := 45 * time.Second
	if p.configFlags.Timeout == nil {
		return timeoutDefault, nil
	}
	// 0 or empty also mean no timeout
	to := strings.TrimSpace(*p.configFlags.Timeout)
	if len(to) == 0 || to == "0" || strings.HasPrefix(to, "0") {
		return timeoutDefault, nil
	}
	return time.ParseDuration(*p.configFlags.Timeout)
}

// ShipwrightClientSet returns a Shipwright Clientset
func (p *Params) ShipwrightClientSet() (buildclientset.Interface, error) {
	if p.buildClientset != nil {
		return p.buildClientset, nil
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
	p.buildClientset, err = buildclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return p.buildClientset, nil
}

// Namespace returns kubernetes namespace with alle the overrides
// from command line and kubernetes config
func (p *Params) Namespace() string {
	if len(p.namespace) == 0 {
		clientConfig := p.configFlags.ToRawKubeConfigLoader()
		clientConfig.ClientConfig()
		p.namespace, _, _ = clientConfig.Namespace()

	}
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
	namespace string,
) *Params {
	return &Params{
		clientset:      clientset,
		buildClientset: shpClientset,
		configFlags:    configFlags,
		namespace:      namespace,
	}
}
