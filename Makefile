# Faber — build & dev tasks.
# On Windows, run these in Git Bash, or install make (e.g. `scoop install make`).

BINARY := faber
ifeq ($(OS),Windows_NT)
	BINARY := faber.exe
endif

PKG := ./...

.PHONY: all build test run fmt vet tidy check clean help

all: build ## Build the binary (default)

build: ## Compile the faber binary
	go build -o ./out/$(BINARY) ./cmd/faber

test: ## Run all tests
	go test $(PKG)

run: ## Build and start the MCP server over stdio
	go run ./cmd/faber mcp start

fmt: ## Format all Go source
	go fmt $(PKG)

vet: ## Run go vet static checks
	go vet $(PKG)

tidy: ## Sync go.mod / go.sum
	go mod tidy

check: fmt vet test ## Format, vet, and test in one go

clean: ## Remove build artifacts
	go clean
	-rm -rf out

help: ## List available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-8s\033[0m %s\n", $$1, $$2}'
