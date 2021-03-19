APP = shp
OUTPUT_DIR ?= _output

CMD = ./cmd/$(APP)/...
PKG = ./pkg/...

BIN ?= $(OUTPUT_DIR)/$(APP)
KUBECTL_BIN ?= $(OUTPUT_DIR)/kubectl-$(APP)

GO_FLAGS ?= -v -mod=vendor
GO_TEST_FLAGS ?= -race -cover

ARGS ?=

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

test: test-unit

.PHONY: test-unit
test-unit:
	go test $(GO_FLAGS) $(GO_TEST_FLAGS) $(CMD) $(PKG) $(ARGS)

travis:
	./hack/install-kubectl.sh
	./hack/install-kind.sh