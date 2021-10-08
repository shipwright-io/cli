package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"

	buildclientset "github.com/shipwright-io/build/pkg/client/clientset/versioned"

	"github.com/shipwright-io/cli/pkg/shp/cmd"
	"github.com/shipwright-io/cli/pkg/shp/cmd/types"
)

// ApplicationName is the name of the application
const ApplicationName = "shp"

func main() {
	flags := pflag.NewFlagSet(ApplicationName, pflag.ExitOnError)
	pflag.CommandLine = flags

	clients := types.ClientSets{}

	configFlags := genericclioptions.NewConfigFlags(true)
	clientConfig := configFlags.ToRawKubeConfigLoader()
	config, err := clientConfig.ClientConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
	if clients.Namespace, _, err = clientConfig.Namespace(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
	if clients.KubernetesClientSet, err = kubernetes.NewForConfig(config); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
	if clients.ShipwrightClientSet, err = buildclientset.NewForConfig(config); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	// Set basic input/output streams for the application
	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}

	// Create a context for the application
	ctx := context.Background()

	rootCmd := cmd.NewRootCmd(ctx, &ioStreams, &clients)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}
