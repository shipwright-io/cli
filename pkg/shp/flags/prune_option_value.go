package flags

import (
	"fmt"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
)

// pruneOptionFlag serves as an adapter to make the Build spec source bundle
// container prune option to be used as a command-line flag (pflag.Value).
type pruneOptionFlag struct {
	ref *buildv1alpha1.PruneOption
}

// Set translates the provided input string into one of the supported prune
// options, or fails with an error in cases of an unsupported value
func (p pruneOptionFlag) Set(val string) error {
	var pruneOption = buildv1alpha1.PruneOption(val)
	switch pruneOption {
	case buildv1alpha1.PruneNever, buildv1alpha1.PruneAfterPull:
		*p.ref = pruneOption
		return nil

	default:
		return fmt.Errorf("supported values are %s, or %s",
			buildv1alpha1.PruneNever,
			buildv1alpha1.PruneAfterPull,
		)
	}
}

// String returns a string representation of a prune option, defaults to the
// spec source container bundle prune default of `Never` in case it is not set
func (p pruneOptionFlag) String() string {
	if p.ref == nil {
		return string(buildv1alpha1.PruneNever)
	}

	return string(*p.ref)
}

// Type returns the type string, which is printed in the usage help output
func (p pruneOptionFlag) Type() string {
	return "pruneOption"
}
