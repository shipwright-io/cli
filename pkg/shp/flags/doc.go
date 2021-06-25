// Package flags contains command-line flags that can be reused over the project, taking real
// Shipwright Build-Controller resources as a direct representation as command-line flags.
//
// For instance:
//
// 	 cmd := &cobra.Command{}
//   br := flags.BuildRunSpecFromFlags(cmd.Flags())
//   flags.SanitizeBuildRunSpec(&br.Spec)
//
// The snippet above shows how to decorate an existing cobra.Command instance with flags, and return
// an instantiated object, which will be receive the inputted values. And, to make sure inner items
// are set to nil when all empty, to the resource in question can be used directly against Kubernetes
// API.
package flags
