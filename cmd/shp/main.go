package main

import (
	"fmt"
	"os"

	"github.com/shipwright-io/cli/pkg/shp/cmd"
)

func main() {
	rootCmd := cmd.NewCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("[ERROR] %#v\n", err)
		os.Exit(1)
	}
}
