package main

import (
	"fmt"
	"os"

	"github.com/shipwright-io/cli/pkg/shp/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// outputDirPath stores the path to the output directory.
var outputDirPath string

var root = &cobra.Command{
	Use:   "gendoc",
	Short: "Generate shp's command-line help messages as markdown",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		genericOpts := &genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stdout}
		shpCmd := cmd.NewCmdSHP(genericOpts)
		shpCmd.DisableAutoGenTag = true
		return doc.GenMarkdownTree(shpCmd, outputDirPath)
	},
}

func init() {
	root.Flags().StringVarP(
		&outputDirPath,
		"output-dir",
		"o",
		".",
		"Path to output directory for generated markdown files",
	)
}

func main() {
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}
