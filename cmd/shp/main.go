package main

import (
	goflag "flag"
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"

	"github.com/shipwright-io/cli/pkg/shp/cmd"
)

// ApplicationName application name.
const ApplicationName = "shp"

func main() {
	initGoFlags()
	initPFlags()

	streams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	rootCmd := cmd.NewCmdSHP(&streams)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

// initGoFlags initializes the flag sets for klog.
// Any flags for "-h" or "--help" are ignored because pflag will show the usage later with all subcommands.
func initGoFlags() {
	flagset := goflag.NewFlagSet(ApplicationName, goflag.ContinueOnError)
	goflag.CommandLine = flagset
	klog.InitFlags(flagset)

	args := []string{}
	for _, arg := range os.Args[1:] {
		if arg != "-h" && arg != "--help" {
			args = append(args, arg)
		}
	}
	flagset.Parse(args)
}

// initPFlags initializes the pflags used by Cobra subcommands.
func initPFlags() {
	flags := pflag.NewFlagSet(ApplicationName, pflag.ExitOnError)
	flags.AddGoFlagSet(goflag.CommandLine)
	pflag.CommandLine = flags
}
