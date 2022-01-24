package flags

import (
	"github.com/spf13/pflag"
)

// FollowFlag register the (log) follow flag, recording the value on the informed boolean pointer.
func FollowFlag(flags *pflag.FlagSet, follow *bool) {
	flags.BoolVarP(
		follow,
		"follow",
		"F",
		*follow,
		"Start a build and watch its log until it completes or fails.",
	)
}
