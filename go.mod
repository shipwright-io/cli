module github.com/shipwright-io/cli

go 1.15

require (
	github.com/google/uuid v1.2.0 // indirect
	github.com/mailru/easyjson v0.7.1 // indirect
	github.com/onsi/gomega v1.10.3
	github.com/pkg/errors v0.9.1
	github.com/shipwright-io/build v0.4.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/tektoncd/pipeline v0.23.0
	github.com/texttheater/golang-levenshtein/levenshtein v0.0.0-20200805054039-cae8b0eaed6c
	k8s.io/api v0.19.7
	k8s.io/apimachinery v0.19.7
	k8s.io/cli-runtime v0.19.7
	k8s.io/client-go v12.0.0+incompatible
	knative.dev/pkg v0.0.0-20210208131226-4b2ae073fa06 // indirect
)

replace k8s.io/client-go => k8s.io/client-go v0.19.7

replace github.com/go-openapi/spec => github.com/go-openapi/spec v0.19.3 // Needed, otherwise we will hit this https://github.com/knative/client/pull/1207#issuecomment-770845105
