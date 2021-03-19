module github.com/shipwright-io/cli

go 1.15

require (
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1
	github.com/shipwright-io/build v0.3.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.19.7
	k8s.io/apimachinery v0.19.7
	k8s.io/cli-runtime v0.19.7
	k8s.io/client-go v12.0.0+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.19.7
