package types

import (
	"context"

	shipwright "github.com/shipwright-io/build/pkg/client/clientset/versioned"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

type ClientSets struct {
	ShipwrightClientSet shipwright.Interface
	KubernetesClientSet kubernetes.Interface
	Namespace           string
}
type SharedOptions struct {
	Clients *ClientSets
	Context context.Context
	Streams *genericclioptions.IOStreams
}
