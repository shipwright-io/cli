module github.com/shipwright-io/cli

go 1.15

require (
	github.com/onsi/gomega v1.10.3
	github.com/pkg/errors v0.9.1
	github.com/shipwright-io/build v0.3.1-0.20210330182238-23d2672f2f61
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.19.7
	k8s.io/apimachinery v0.19.7
	k8s.io/cli-runtime v0.19.7
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog/v2 v2.5.0
)

replace k8s.io/client-go => k8s.io/client-go v0.19.7

replace github.com/go-openapi/spec => github.com/go-openapi/spec v0.19.3 // Needed, otherwise we will hit this https://github.com/knative/client/pull/1207#issuecomment-770845105
