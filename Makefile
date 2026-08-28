MODULE_NAME=httpmock

GOLANGCI_LINT_VERSION ?= v2.13.0

GO ?= go
GOLANGCI_LINT ?= $(shell go env GOPATH)/bin/golangci-lint-$(GOLANGCI_LINT_VERSION)

VENDOR_DIR = vendor
GOROOT_DIR = $(shell $(GO) env GOROOT)

GITHUB_OUTPUT ?= /dev/null

# Other config
NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

ifeq ($(V),1)
  Q = @set -x;
else
  Q = @
endif

.PHONY: $(VENDOR_DIR)
$(VENDOR_DIR):
	$(Q)mkdir -p $(VENDOR_DIR)
	$(Q)$(GO) mod vendor
	$(Q)$(GO) mod tidy

.PHONY: bump-deps
bump-deps:
	$(Q)$(GO) get -u ./...

.PHONY: tidy
tidy:
	$(Q)$(GO) mod tidy

.PHONY: lint
ifeq ($(V),1)
  GOLANGCI_LINT_FLAGS = -vvvv
else
  GOLANGCI_LINT_FLAGS =
endif

lint: $(GOLANGCI_LINT)
	@printf -- "$(OK_COLOR)==> lint$(NO_COLOR)\n"
	$(Q)GOROOT=$(GOROOT_DIR) PATH="$(GOROOT_DIR)/bin:$$PATH" $(GOLANGCI_LINT) run -c .golangci.yaml --color always $(GOLANGCI_LINT_FLAGS)

.PHONY: test
test: test-unit

## Run unit tests
.PHONY: test-unit
test-unit:
	@printf -- "$(OK_COLOR)==> unit test$(NO_COLOR)\n"
	$(Q)$(GO) test -gcflags=-l -coverprofile=unit.coverprofile -covermode=atomic -race ./...

.PHONY: $(GITHUB_OUTPUT)
$(GITHUB_OUTPUT):
	@echo "MODULE_NAME=$(MODULE_NAME)" >> "$@"
	@echo "GOLANGCI_LINT_VERSION=$(GOLANGCI_LINT_VERSION)" >> "$@"

$(GOLANGCI_LINT):
	@printf -- "$(OK_COLOR)==> Installing golangci-lint $(GOLANGCI_LINT_VERSION)$(NO_COLOR)\n"
	$(Q)curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b /tmp "$(GOLANGCI_LINT_VERSION)"
	$(Q)$(call install-dep,/tmp/golangci-lint,$(GOLANGCI_LINT))

define install-dep
	if [ "$(1)" != "$(2)" ]; then \
		mkdir -p $$(dirname $(2)) || true; \
		mv $(1) $(2); \
	fi

	chmod +x "$(2)"
endef
