package main

import (
	goflag "flag"
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"

	"github.com/shipwright-io/cli/pkg/shp/cmd"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// ApplicationName application name.
const ApplicationName = "shp"

var hiddenLogFlags = []string{
	"add_dir_header",
	"alsologtostderr",
	"log_backtrace_at",
	"log_dir",
	"log_file",
	"log_file_max_size",
	"logtostderr",
	"one_output",
	"skip_headers",
	"skip_log_headers",
	"stderrthreshold",
	"v",
	"vmodule",
}

func main() {
	if err := initGoFlags(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
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
func initGoFlags() error {
	flagset := goflag.NewFlagSet(ApplicationName, goflag.ContinueOnError)
	goflag.CommandLine = flagset
	klog.InitFlags(flagset)

	args := []string{}
	for _, arg := range os.Args[1:] {
		if arg != "-h" && arg != "--help" {
			args = append(args, arg)
		}
	}
	return flagset.Parse(args)
}

// initPFlags initializes the pflags used by Cobra subcommands.
func initPFlags() {
	flags := pflag.NewFlagSet(ApplicationName, pflag.ExitOnError)
	flags.AddGoFlagSet(goflag.CommandLine)
	pflag.CommandLine = flags

	for _, flag := range hiddenLogFlags {
		if err := flags.MarkHidden(flag); err != nil {
			panic(err)
		}
	}
}
