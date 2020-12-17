package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
)

// Runner execute the sub-command lifecycle, wrapper around sub-commands.
type Runner struct {
	opts      *Options                    // global options
	ioStreams genericclioptions.IOStreams // input, output and error io streams
	subCmd    SubCommand                  // sub-command instance
}

// Cmd is a wrapper around sub-command's Cobra, it wires up global flags and set a single RunE
// executor to self.
func (r *Runner) Cmd() *cobra.Command {
	cmd := r.subCmd.Cmd()
	cmd.RunE = r.RunE
	r.opts.AddFlags(cmd.PersistentFlags())
	return cmd
}

// dynamicClientNamespace instantiate a dynamic client, and configure the target namespace. When
// --namespace is not informed, it uses the default configured locally.
func (r *Runner) dynamicClientNamespace() (dynamic.Interface, string, error) {
	f := r.opts.Factory()
	configLoader := f.ToRawKubeConfigLoader()

	namespace := *r.opts.configFlags.Namespace
	if namespace == "" {
		var err error
		ns, _, err := configLoader.Namespace()
		if err != nil {
			return nil, "", err
		}
		namespace = ns
	}

	restConfig, err := configLoader.ClientConfig()
	if err != nil {
		return nil, "", err
	}
	client, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, "", err
	}
	return client, namespace, nil
}

// RunE cobra.Command's RunE implementation focusing on sub-commands lifecycle. To achieve it, a
// dynamic client and configured namespace are informed.
func (r *Runner) RunE(cmd *cobra.Command, args []string) error {
	client, ns, err := r.dynamicClientNamespace()
	if err != nil {
		return err
	}

	if err = r.subCmd.Complete(client, ns, args); err != nil {
		return err
	}
	if err = r.subCmd.Validate(); err != nil {
		return err
	}
	return r.subCmd.Run(client, ns)
}

// NewRunner instantiate a Runner.
func NewRunner(opts *Options, ioStreams genericclioptions.IOStreams, subCmd SubCommand) *Runner {
	return &Runner{opts: opts, ioStreams: ioStreams, subCmd: subCmd}
}
