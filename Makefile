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

###
### Start of sanity check targets
###

.PHONY: govet
govet:
	@echo "Checking go vet"
	@go vet ./...

# Install it via: go install github.com/gordonklaus/ineffassign@latest
.PHONY: ineffassign
ineffassign:
	@echo "Checking ineffassign"
	@ineffassign ./...

# Install it via: go install golang.org/x/lint/golint@latest
# See https://github.com/golang/lint/issues/320 for details regarding the grep
.PHONY: golint
golint:
	@echo "Checking golint"
	@go list ./... | grep -v -e /vendor -e /test | xargs -L1 golint -set_exit_status

# Install it via: go install github.com/securego/gosec/v2/cmd/gosec@latest
.PHONY: gosec
gosec:
	@echo "Checking gosec"
	gosec -confidence medium -severity high ./...

# Install it via: go install github.com/client9/misspell/cmd/misspell@latest
.PHONY: misspell
misspell:
	@echo "Checking misspell"
	@find . -type f -not -path './vendor/*' -not -path './.git/*' -not -path './build/*' -print0 | xargs -0 misspell -source=text -error

# Install it via: go install honnef.co/go/tools/cmd/staticcheck@latest
.PHONY: staticcheck
staticcheck:
	@echo "Checking staticcheck"
	@go list ./... | grep -v /test | xargs staticcheck

.PHONY: sanity-check
sanity-check: ineffassign golint gosec govet misspell staticcheck
