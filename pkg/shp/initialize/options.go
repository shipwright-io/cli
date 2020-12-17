package initialize

import "github.com/spf13/pflag"

type Options struct{}

func (o *Options) AddFlags(flags *pflag.FlagSet) {}

func NewOptions() *Options {
	return &Options{}
}
