module github.com/otaviof/shp

go 1.15

require (
	github.com/Sirupsen/logrus v0.0.0-00010101000000-000000000000 // indirect
	github.com/onsi/gomega v1.10.2
	github.com/shipwright-io/build v0.1.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	golang.org/x/net v0.0.0-20200904194848-62affa334b73 // indirect
	golang.org/x/text v0.3.3 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	k8s.io/api v0.17.6
	k8s.io/apimachinery v0.17.6
	k8s.io/cli-runtime v0.17.6
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kubectl v0.17.6
)

replace github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.6.0
