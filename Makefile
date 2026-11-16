VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
GOLANGCI_LINT_VERSION := v2.14.0
GOLANGCI_LINT ?= go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
ADAPTERS := adapter/gin adapter/gorm adapter/pgx adapter/sql
MODULES := . $(ADAPTERS)
CORE := github.com/DarkoKlisuric/godebugbar

.PHONY: test test-integration lint fmt tidy version

test:
	@for m in $(MODULES); do (cd $$m && go test -race ./...) || exit 1; done

test-integration:
	@for m in $(MODULES); do (cd $$m && go test -race -tags=integration ./...) || exit 1; done

lint:
	@for m in $(MODULES); do (cd $$m && $(GOLANGCI_LINT) run ./...) || exit 1; done
	go run ./tools/codecheck .

fmt:
	@for m in $(MODULES); do (cd $$m && $(GOLANGCI_LINT) fmt ./...) || exit 1; done

tidy:
	go mod tidy
	@for m in $(ADAPTERS); do (cd $$m && go mod edit -replace=$(CORE)=../.. && GOWORK=off go mod tidy; status=$$?; go mod edit -dropreplace=$(CORE); exit $$status) || exit 1; done

version:
	@echo $(VERSION)
