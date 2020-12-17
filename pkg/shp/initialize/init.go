package initialize

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
)

// TODO; write me!

type Initialize struct {
	cmd *cobra.Command
}

func (i *Initialize) Complete(client dynamic.Interface, ns string, args []string) error {
	return nil
}

func (i *Initialize) Validate() error {
	return nil
}

func (i *Initialize) Run(client dynamic.Interface, ns string) error {
	return nil
}

func (i *Initialize) Cmd() *cobra.Command {
	return i.cmd
}

func NewInitialize() *Initialize {
	return &Initialize{cmd: &cobra.Command{
		Use:   "init [directory]",
		Short: "create a Build resource based on local repository",
	}}
}
