APP = shp
OUTPUT_DIR ?= _output

CMD = ./cmd/$(APP)/...
PKG = ./pkg/...

BIN ?= $(OUTPUT_DIR)/$(APP)
KUBECTL_BIN ?= $(OUTPUT_DIR)/kubectl-$(APP)

GO_FLAGS ?= -v -mod=vendor
GO_TEST_FLAGS ?= -race -cover

GO_PATH ?= $(shell go env GOPATH)
GO_CACHE ?= $(shell go env GOCACHE)

ARGS ?=

# container registry deployment namespace
REGISTRY_NAMESPACE ?= registry

# hostname and namespace for the end-to-end tests producing container images
OUTPUT_HOSTNAME ?= registry.registry.svc.cluster.local:32222
OUTPUT_NAMESPACE ?= shipwright-io

# Tekton and Shipwright Build Controller versions for CI
TEKTON_VERSION ?= v0.38.3
SHIPWRIGHT_VERSION ?= v0.11.0

.EXPORT_ALL_VARIABLES:

.PHONY: $(BIN)
$(BIN):
	go build $(GO_FLAGS) -o $(BIN) $(CMD)

build: $(BIN)

install: build
	install -m 0755 $(BIN) /usr/local/bin/

# creates a kubectl prefixed shp binary, "kubectl-shp", and when installed under $PATH, will be
# visible as "kubectl shp".
.PHONY: kubectl
kubectl: BIN = $(KUBECTL_BIN)
kubectl: $(BIN)

kubectl-install: BIN = $(KUBECTL_BIN)
kubectl-install: kubectl install

clean:
	rm -rf "$(OUTPUT_DIR)"

run:
	go run $(GO_FLAGS) $(CMD) $(ARGS)

# runs all tests, unit and end-to-end.
test: test-unit test-e2e

.PHONY: test-unit
test-unit:
	go test $(GO_FLAGS) $(GO_TEST_FLAGS) $(CMD) $(PKG) $(ARGS)

# looks for *.bats files in the test/e2e directory and runs them
test-e2e:
	bats --recursive test/e2e/*.bats

# runs act, with optional arguments
.PHONY: act
act:
	@act --secret="GITHUB_TOKEN=${GITHUB_TOKEN}" $(ARGS)

# Install golangci-lint via: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
.PHONY: sanity-check
sanity-check:
	golangci-lint run

# generates the command-line help messages as markdown files, and in order to have a generic ${HOME}
# rendered, it exports a fake home directory ("~"), and preserves the original GOPATH and GOCACHE to avoid
# creating bogus files on the project directory.
.PHONY: generate-docs
generate-docs:
	GOPATH="$(GO_PATH)" GOCACHE="$(GO_CACHE)" HOME="~" go run cmd/help/main.go --output-dir=./docs

# checks if the generated documentation files are out of sync.
.PHONY: verify-docs
verify-docs: generate-docs
	./hack/verify-docs.sh
