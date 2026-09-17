VERSION     ?= dev
BUILD_DATE  := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
MODULE      := github.com/hamimlohani/gtree
LDFLAGS     := -ldflags "-X $(MODULE)/cmd.Version=$(VERSION) \
                          -X $(MODULE)/cmd.BuildDate=$(BUILD_DATE) \
                          -s -w"

BINARY      := gtree
INSTALL_DIR := $(HOME)/.local/bin
COVERAGE    := coverage.out

# ─── Development ──────────────────────────────────────────────────────────────

.PHONY: build
build: ## Build binary for the local machine
	go build $(LDFLAGS) -o $(BINARY) .
	@echo "Built ./$(BINARY)"

.PHONY: run
run: ## Build and run (pass extra args via ARGS="...")
	go run $(LDFLAGS) . $(ARGS)

.PHONY: install
install: build ## Install to $(INSTALL_DIR)
	@mkdir -p $(INSTALL_DIR)
	@cp $(BINARY) $(INSTALL_DIR)/$(BINARY)
	@echo "Installed $(INSTALL_DIR)/$(BINARY)"

.PHONY: uninstall
uninstall: ## Remove installed binary
	@rm -f $(INSTALL_DIR)/$(BINARY)
	@echo "Removed $(INSTALL_DIR)/$(BINARY)"

# ─── Quality ──────────────────────────────────────────────────────────────────

.PHONY: test
test: ## Run unit + integration tests with race detector
	go test -v -race ./...

.PHONY: test-short
test-short: ## Run only fast unit tests (no git subprocess)
	go test -v -race -short ./...

.PHONY: cover
cover: ## Run tests and open HTML coverage report
	go test -race -coverprofile=$(COVERAGE) ./...
	go tool cover -html=$(COVERAGE)

.PHONY: lint
lint: ## Run golangci-lint (install: brew install golangci-lint)
	golangci-lint run ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: check
check: vet test ## Run vet + tests (CI-friendly, no color needed)

.PHONY: tidy
tidy: ## Tidy go.mod / go.sum
	go mod tidy

.PHONY: clean
clean: ## Remove build artifacts and coverage files
	@rm -f $(BINARY) $(COVERAGE) coverage.html
	@rm -rf dist/
	@echo "Cleaned."

# ─── Cross-compilation ────────────────────────────────────────────────────────

PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64

.PHONY: build-all
build-all: ## Build for all platforms → dist/
	@mkdir -p dist
	@$(foreach PLATFORM,$(PLATFORMS), \
		$(eval _OS   := $(word 1,$(subst /, ,$(PLATFORM)))) \
		$(eval _ARCH := $(word 2,$(subst /, ,$(PLATFORM)))) \
		printf "  %-20s " "$(_OS)/$(_ARCH)" && \
		CGO_ENABLED=0 GOOS=$(_OS) GOARCH=$(_ARCH) \
			go build $(LDFLAGS) -o dist/$(BINARY)-$(_OS)-$(_ARCH) . && \
		echo "✓"; \
	)
	@echo "\ndist/:"
	@ls -lh dist/

.PHONY: build-darwin
build-darwin: ## Build for macOS (amd64 + arm64)
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64  go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64  go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64 .
	@echo "Built macOS binaries in dist/"

.PHONY: build-linux
build-linux: ## Build for Linux (amd64 + arm64)
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64  go build $(LDFLAGS) -o dist/$(BINARY)-linux-arm64 .
	@echo "Built Linux binaries in dist/"

# ─── Release ──────────────────────────────────────────────────────────────────

.PHONY: snapshot
snapshot: ## Dry-run goreleaser snapshot (no publish, no git tag needed)
	goreleaser release --snapshot --clean

.PHONY: release-dry
release-dry: ## Goreleaser release dry-run (validates .goreleaser.yml)
	goreleaser release --skip=publish --clean

.PHONY: release
release: ## Publish a release via goreleaser (requires GITHUB_TOKEN + pushed tag)
	@if [ -z "$(GITHUB_TOKEN)" ]; then echo "Error: GITHUB_TOKEN is not set"; exit 1; fi
	goreleaser release --clean

.PHONY: tag
tag: ## Create and push a version tag. Usage: make tag VERSION=v1.2.3
	@if [ -z "$(VERSION)" ] || [ "$(VERSION)" = "dev" ]; then \
		echo "Usage: make tag VERSION=v1.2.3"; exit 1; fi
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	@echo "Tagged and pushed $(VERSION)"

# ─── Help ─────────────────────────────────────────────────────────────────────

.PHONY: help
help: ## Show this help
	@printf "\n\033[1mgtree — available targets\033[0m\n\n"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'
	@echo ""

.DEFAULT_GOAL := help
