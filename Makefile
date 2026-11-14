VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
GOLANGCI_LINT_VERSION := v2.14.0

.PHONY: test test-integration lint fmt tidy version

test:
	go test -race ./...

test-integration:
	go test -race -tags=integration ./...

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./...
	go run ./tools/codecheck .

fmt:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) fmt ./...

tidy:
	go mod tidy

version:
	@echo $(VERSION)
