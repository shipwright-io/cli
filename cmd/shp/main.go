package main

import (
	"fmt"
	"os"

	"github.com/otaviof/shp/pkg/shp/cmd"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// ApplicationName application name.
const ApplicationName = "kubectl-shp"

func main() {
	flags := pflag.NewFlagSet(ApplicationName, pflag.ExitOnError)
	pflag.CommandLine = flags

	streams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	rootCmd := cmd.NewCmdSHP(streams)
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("[ERROR] %#v\n", err)
		os.Exit(1)
	}
}
