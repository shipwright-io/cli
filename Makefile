APP = shp
OUTPUT_DIR ?= _output

CMD = ./cmd/$(APP)/...
PKG = ./pkg/...

BIN ?= $(OUTPUT_DIR)/$(APP)

GO_FLAGS ?= -v -mod=vendor
GO_TEST_FLAGS ?= -race -cover

ARGS ?=

default: $(BIN)

.PHONY: $(BIN)
$(BIN):
	go build $(GO_FLAGS) -o $(BIN) $(CMD)

# Creates the application binary under output directory.
#
# 	make build
build: $(BIN)

# Executes the application main binary from source-code with "go run". It takes "ARGS" as the
# command-line direct arguments.
#
# 	make run ARGS='--help'
run:
	go run $(GO_FLAGS) $(CMD) $(ARGS)

# Single target to run all tests.
#
#	make test
test: test-unit

# Execute unit-tests.
#
#	make test-unit
.PHONY: test-unit
test-unit:
	go test $(GO_FLAGS) $(GO_TEST_FLAGS) $(CMD) $(PKG) $(ARGS)
