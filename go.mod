module upgrade-all-services-cli-plugin

go 1.24.6

require (
	code.cloudfoundry.org/cli v7.1.0+incompatible
	code.cloudfoundry.org/jsonry v1.1.4
	github.com/blang/semver/v4 v4.0.0
	github.com/hashicorp/go-version v1.7.0
	github.com/onsi/ginkgo/v2 v2.23.4
	github.com/onsi/gomega v1.38.0
)

require (
	github.com/BurntSushi/toml v1.4.1-0.20240526193622-a339e1f7089c // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20250403155104-27863c87afa6 // indirect
	github.com/maxbrunsfeld/counterfeiter/v6 v6.11.2 // indirect
	go.uber.org/automaxprocs v1.6.0 // indirect
	golang.org/x/exp/typeparams v0.0.0-20231108232855-2478ac86f678 // indirect
	golang.org/x/mod v0.25.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	golang.org/x/tools v0.33.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	honnef.co/go/tools v0.6.1 // indirect
)

tool (
	github.com/maxbrunsfeld/counterfeiter/v6
	github.com/onsi/ginkgo/v2/ginkgo
	golang.org/x/tools/cmd/goimports
	honnef.co/go/tools/cmd/staticcheck
)
