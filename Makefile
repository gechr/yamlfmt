GO       ?= go
GO_TOOLS ?= $(shell $(GO) list tool)

DIST_DIR ?= dist

.PHONY: all
all: fmt lint test

.PHONY: fmt
fmt:
	@$(GO) fix ./...
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
	@$(GO) test -timeout 30s -race ./...

.PHONY: update
update:
	@clover run
	@$(GO) get $(GO_TOOLS) $(shell $(GO) list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)
	@$(GO) mod tidy
