package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
)

// SubCommand defines the methods for a sub-command wrapped with Runner.
type SubCommand interface {
	// Cmd shares the cobra.Command instance.
	Cmd() *cobra.Command
	// Complete aggregate data needed for the sub-command primary logic.
	Complete(client dynamic.Interface, ns string, args []string) error
	// Validate perform validation against the context collected.
	Validate() error
	// Run execute the primary sub-command logic.
	Run(client dynamic.Interface, ns string) error
}
