VENDOR_DIR = vendor

GOLANGCI_LINT_VERSION ?= v1.48.0

GO ?= go
GOLANGCI_LINT ?= golangci-lint-$(GOLANGCI_LINT_VERSION)

.PHONY: $(VENDOR_DIR)
$(VENDOR_DIR):
	@mkdir -p $(VENDOR_DIR)
	@$(GO) mod vendor

.PHONY: lint
lint: bin/$(GOLANGCI_LINT) $(VENDOR_DIR)
	@bin/$(GOLANGCI_LINT) run -c .golangci.yaml

.PHONY: test
test: test-unit

## Run unit tests
.PHONY: test-unit
test-unit:
	@echo ">> unit test"
	@$(GO) test -gcflags=-l -coverprofile=unit.coverprofile -covermode=atomic -race ./...

bin/$(GOLANGCI_LINT):
	@echo "$(OK_COLOR)==> Installing golangci-lint $(GOLANGCI_LINT_VERSION)$(NO_COLOR)"; \
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./bin "$(GOLANGCI_LINT_VERSION)"
	@mv ./bin/golangci-lint bin/$(GOLANGCI_LINT)
