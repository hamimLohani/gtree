VERSION     ?= dev
BUILD_DATE  := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS     := -ldflags "-X github.com/hamimlohani/gtree/cmd.Version=$(VERSION) \
                          -X github.com/hamimlohani/gtree/cmd.BuildDate=$(BUILD_DATE) \
                          -s -w"

BINARY      := gtree
INSTALL_DIR := $(HOME)/.local/bin

# ─── Development targets ──────────────────────────────────────────────────────

.PHONY: build
build: ## Build for the local machine
	go build $(LDFLAGS) -o $(BINARY) .

.PHONY: install
install: build ## Install to $(INSTALL_DIR)
	@mkdir -p $(INSTALL_DIR)
	@cp $(BINARY) $(INSTALL_DIR)/$(BINARY)
	@echo "Installed to $(INSTALL_DIR)/$(BINARY)"

.PHONY: run
run: ## Build and run against the current directory
	go run . $(ARGS)

.PHONY: test
test: ## Run all unit tests
	go test -v -race ./...

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run ./...

.PHONY: tidy
tidy: ## Tidy go.mod / go.sum
	go mod tidy

.PHONY: clean
clean: ## Remove build artifacts
	rm -f $(BINARY)
	rm -rf dist/

# ─── Cross-compilation ────────────────────────────────────────────────────────

PLATFORMS := \
	darwin/amd64 \
	darwin/arm64  \
	linux/amd64   \
	linux/arm64

.PHONY: build-all
build-all: ## Build for all platforms into dist/
	@mkdir -p dist
	@$(foreach PLATFORM,$(PLATFORMS), \
		$(eval GOOS   := $(word 1,$(subst /, ,$(PLATFORM)))) \
		$(eval GOARCH := $(word 2,$(subst /, ,$(PLATFORM)))) \
		echo "Building $(GOOS)/$(GOARCH)..." && \
		GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) \
			-o dist/$(BINARY)-$(GOOS)-$(GOARCH) . && \
	) echo "Done."

# ─── Release (goreleaser) ────────────────────────────────────────────────────

.PHONY: release-dry
release-dry: ## Dry-run goreleaser (no publish)
	goreleaser release --snapshot --clean

.PHONY: release
release: ## Publish a release (requires GITHUB_TOKEN + git tag)
	goreleaser release --clean

# ─── Help ────────────────────────────────────────────────────────────────────

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
