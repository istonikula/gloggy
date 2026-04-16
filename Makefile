BIN := dist/gloggy

.PHONY: help
help:
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":[^:]*?## "}; {printf "%-20s %s\n", $$1, $$2}'

.PHONY: install
install: ## Install to $GOBIN (or $GOPATH/bin)
	go install ./cmd/gloggy

.PHONY: build
build: ## Build the distributable
	go build -o ${BIN} ./cmd/gloggy

.PHONY: test
test: ## Run short tests
	go test -short -count=1 ./...

.PHONY: integration
integration: build ## Run integration tests
	go test -count=1 ./tests/integration/...

.PHONY: test-all
test-all: ## Run all tests
	go test -count=1 ./...

.PHONY: clean
clean: ## Clean intermediate build products and remove distributable
	go clean
	$(RM) -f ${BIN}
