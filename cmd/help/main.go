package main

import (
	"fmt"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"os"

	"github.com/shipwright-io/cli/pkg/shp/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// outputDirPath stores the path to the output directory.
var outputDirPath string

var hiddenFlags = []string{
	"as-user-extra",
}

var root = &cobra.Command{
	Use:   "gendoc",
	Short: "Generate shp's command-line help messages as markdown",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		genericOpts := &genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stdout}
		shpCmd := cmd.NewCmdSHP(genericOpts)
		shpCmd.DisableAutoGenTag = true
		for _, flag := range hiddenFlags {
			_ = shpCmd.Flags().MarkHidden(flag)
			_ = shpCmd.PersistentFlags().MarkHidden(flag)
		}
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
