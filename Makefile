APP = shp
OUTPUT_DIR ?= _output

CMD = ./cmd/$(APP)/...
PKG = ./pkg/...

BIN ?= $(OUTPUT_DIR)/$(APP)
KUBECTL_BIN ?= $(OUTPUT_DIR)/kubectl-$(APP)

GO_FLAGS ?= -v -mod=vendor
GO_TEST_FLAGS ?= -race -cover

ARGS ?=

# Tekton and Shipwright Build Controller versions for CI
TEKTON_VERSION ?= v0.30.0
SHIPWRIGHT_VERSION ?= nightly-2022-01-21-1642741753

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
	./test/e2e/bats/core/bin/bats --recursive test/e2e/*.bats

# wait for KinD cluster to be on ready state, so tests can be performed
verify-kind:
	./hack/verify-kind.sh

# deploys Tekton and Shipwright Build Controller following the versions exported
install-shipwright:
	./hack/install-tekton.sh
	./hack/install-shipwright.sh

# Install golangci-lint via: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
.PHONY: sanity-check
sanity-check:
	golangci-lint run
