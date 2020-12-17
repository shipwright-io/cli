package cmd

import (
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	kcmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// Options global options informed to the command-line, includes generic kubectl flags. This should
// contain all information needed for instantiating a API client.
type Options struct {
	configFlags       *genericclioptions.ConfigFlags
	matchVersionFlags *kcmdutil.MatchVersionFlags
	factory           kcmdutil.Factory
	dryRun            bool
}

// AddFlags register global flags.
func (o *Options) AddFlags(flags *pflag.FlagSet) {
	flags.BoolVar(&o.dryRun, "dry-run", false, "dry-run mode, no changes are issued.")

	o.configFlags.AddFlags(flags)
	o.matchVersionFlags.AddFlags(flags)
}

// Factory exposes the kubectl's Factory to instantiate API clients and interact with configuration.
func (o *Options) Factory() kcmdutil.Factory {
	if o.factory != nil {
		return o.factory
	}
	o.factory = kcmdutil.NewFactory(o.matchVersionFlags)
	return o.factory
}

// NewOptions instantiate options and kubectl generic flags.
func NewOptions() *Options {
	configFlags := genericclioptions.NewConfigFlags(true)
	matchVersionFlags := kcmdutil.NewMatchVersionFlags(configFlags)

	return &Options{configFlags: configFlags, matchVersionFlags: matchVersionFlags}
}
