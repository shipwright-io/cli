package main

import (
	"flag"
	"os"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"

	"github.com/shipwright-io/cli/pkg/shp/cmd"
)

// ApplicationName application name.
const ApplicationName = "shp"

func main() {
	klogFlags := flag.NewFlagSet(ApplicationName, flag.ExitOnError)
	klog.InitFlags(klogFlags)

	streams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	rootCmd := cmd.NewCmdSHP(&streams)
	if err := rootCmd.Execute(); err != nil {
		klog.Errorf("Error: %v", err)
		os.Exit(1)
	}
}
