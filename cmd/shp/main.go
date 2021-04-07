package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/shipwright-io/cli/pkg/shp/cmd"
)

// ApplicationName application name.
const ApplicationName = "shp"

func main() {
	flags := pflag.NewFlagSet(ApplicationName, pflag.ExitOnError)
	pflag.CommandLine = flags

	streams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	rootCmd := cmd.NewCmdSHP(&streams)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}
