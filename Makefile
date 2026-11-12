GO       ?= go
GO_TOOLS ?= $(shell $(GO) list tool)
# Packages outside the vendored YAML library, which stays as upstream wrote it.
GO_PKGS  ?= $(shell $(GO) list ./... | grep -v /internal/yaml)

DIST_DIR ?= dist

.PHONY: all
all: fmt lint test

.PHONY: fmt
fmt:
	@$(GO) fix $(GO_PKGS)
	@$(GO) tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint fmt --enable=gci,golines,gofumpt
	@$(GO) tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint run --fix -c .golangci.ruleguard.yml

.PHONY: build
build:
	@$(GO) build -o $(DIST_DIR)/yamlfmt .

.PHONY: install
install:
	@$(GO) install

.PHONY: lint
lint:
	@$(GO) tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint run

.PHONY: test
test:
	@$(GO) test -timeout 2m -race ./...

.PHONY: update
update:
	@clover run
	@$(GO) get $(GO_TOOLS) $(shell $(GO) list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)
	@$(GO) mod tidy
