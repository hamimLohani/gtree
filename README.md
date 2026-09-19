# 🌳 gtree

> Scan your directory tree for git repositories and see their live status at a glance —
> like `gh` meets `tree`, built for developers who manage many repos.

```
 🌳 gtree  ~/Developer  35 repos

├── Arduino-Project/
│   ├── ESP32_SmartBus_Complete         [main]  ✓          6mo ago
│   ├── Smart-Multi-plug                [main]  ✓          1mo ago
│   └── SmartHome-esp32-code            [main]  ✗ ~1       2mo ago
├── Shell/
│   └── Project/
│       ├── gtree                       [main]  ✗ ~7  ↑2   just now
│       └── run-cplex                   [main]  ✓          8mo ago
└── Websites/
    ├── Student-Aid/
    │   └── student-aid                 [main]  ✓     ↑1   1w ago
    └── portfolio                       [main]  ✓          5d ago

────────────────────────────────────────────────────────────────────────────────
  35 repos  •  24 clean  •  11 dirty  •  2 unpushed ↑
```

> 📽 **Demo GIF** — _record one with [vhs](https://github.com/charmbracelet/vhs) and drop it here_

---

## ✨ Features

### 🔍 Scanning
| Feature | Details |
|---|---|
| **Recursive discovery** | Walks directory tree finding `.git` repos at any depth |
| **Configurable depth** | `--depth N` limits recursion; `0` = unlimited |
| **Concurrent collection** | Semaphore worker pool — 35 repos queried in ~1 second |
| **Noise filtering** | Skips `node_modules`, `vendor`, `.venv`, `Library`, `.Trash`, and more by default |
| **Safe traversal** | Skips symlinks, bare repos, submodule `.git` files, and permission-denied paths gracefully |
| **Nested grouping** | Repos grouped by their actual folder hierarchy — not a flat list |

### 📊 Per-repo status
| Field | Description |
|---|---|
| **Branch** | Current branch name, or `detached @ <sha>` for detached HEAD |
| **Dirty / clean** | `✗ ~N` (red) = N changed files · `✓` (green) = clean working tree |
| **Ahead / behind** | `↑N` (yellow) commits ahead · `↓N` (orange) commits behind upstream |
| **Stash count** | `[$N]` (cyan) when stash entries exist |
| **Last commit** | Human-relative time: `just now` → `5m ago` → `2h ago` → `3d ago` → `1mo ago` → `2y ago` |

### 🎨 Visual design
| Feature | Details |
|---|---|
| **Catppuccin palette** | Mocha (dark) / Latte (light) — auto-detected from terminal background |
| **`NO_COLOR` support** | Detects `NO_COLOR` env var and non-TTY output — degrades to plain text for piping |
| **Terminal-width aware** | Long names truncated cleanly with `…` rather than wrapping |
| **Box-drawing tree** | `├──` / `└──` connectors, directory nodes in muted italic |
| **Column alignment** | All status columns pixel-aligned regardless of repo name length |
| **Animated spinner** | `⠙⠹⠸⠼⠴…` dot spinner while scanning |

### ⌨️ CLI flags
| Flag | Description |
|---|---|
| `[path]` | Scan path (default: current directory) |
| `-d, --depth N` | Maximum recursion depth (0 = unlimited) |
| `-s, --sort dirty\|recent\|name` | Sort order (default: `name`) |
| `--only-dirty` | Hide clean repositories |
| `--watch` | Live-refreshing view — re-scans every 5 seconds |
| `--no-color` | Disable color output |
| `-c, --config FILE` | Use a specific config file |
| `-v, --version` | Print version and build date |
| `-h, --help` | Full help text |

### 🔁 Watch mode
- Runs in **alt-screen** — doesn't pollute terminal scrollback
- Previous tree stays visible while **refreshing overlay** animates
- **`r`** — manual refresh immediately
- **`q` / `Esc`** — quit
- Status bar: `Updated 3s ago  •  r refresh  •  q quit  •  refreshes every 5s`

### ⚙️ Config file
Auto-created at `~/.config/gtree/config.yml` on first run with sane defaults.
Every value is overridable via CLI flag or `GTREE_<KEY>` env var.

### 🚀 Release tooling
| Tool | Details |
|---|---|
| **Makefile** | 20 targets: build, test, cover, lint, cross-compile, release, tag |
| **goreleaser** | Builds tarballs (Linux) + bare binaries (macOS), SHA-256 checksums, grouped semantic changelog |
| **GitHub Actions CI** | Go 1.21 + 1.22 × ubuntu + macOS matrix on every push/PR |
| **GitHub Actions Release** | Tag push → tests → goreleaser → GitHub Releases (automated) |
| **install.sh** | POSIX one-liner installer: auto-detects OS/arch, optional sudo, PATH hint |
| **golangci-lint** | errcheck · staticcheck · gocritic · misspell · prealloc · goimports |

---

## 📦 Installation

### One-liner (recommended)
```sh
curl -sSfL https://raw.githubusercontent.com/hamimlohani/gtree/main/install.sh | sh
```

Options:
```sh
# Install a specific version
GTREE_VERSION=v1.2.0 curl -sSfL .../install.sh | sh

# Install without sudo (forces ~/.local/bin)
GTREE_NO_SUDO=1 curl -sSfL .../install.sh | sh

# Install to a custom directory
GTREE_INSTALL_DIR=/opt/bin curl -sSfL .../install.sh | sh
```

### Homebrew (macOS / Linux)

```sh
# Install latest from source (builds with Go — works before a tagged release)
brew install --HEAD hamimlohani/tap/gtree
```

> **After the first `make tag VERSION=v1.0.0`**, goreleaser will update the tap formula
> with a stable tarball URL + SHA-256, and you can drop `--HEAD`:
> ```sh
> brew install hamimlohani/tap/gtree
> brew upgrade gtree   # upgrade to latest release
> ```

### Pre-built binary
Download from the [Releases page](https://github.com/hamimlohani/gtree/releases):

```sh
# macOS Apple Silicon
curl -L https://github.com/hamimlohani/gtree/releases/latest/download/gtree-darwin-arm64 \
  -o /usr/local/bin/gtree && chmod +x /usr/local/bin/gtree

# macOS Intel
curl -L https://github.com/hamimlohani/gtree/releases/latest/download/gtree-darwin-amd64 \
  -o /usr/local/bin/gtree && chmod +x /usr/local/bin/gtree

# Linux amd64
curl -L https://github.com/hamimlohani/gtree/releases/latest/download/gtree-<version>-linux-amd64.tar.gz \
  | tar xz -C /usr/local/bin
```

### From source (requires Go 1.21+)
```sh
git clone https://github.com/hamimlohani/gtree.git
cd gtree
make install          # builds and installs to ~/.local/bin/gtree
```

### go install
```sh
go install github.com/hamimlohani/gtree@latest
```

---

## 🖥️ Usage

```
Usage:
  gtree [path] [flags]
  gtree [command]

Available Commands:
  status      Scan a directory and display git repository status
  config      Show or edit the gtree config file
  version     Print version information
  help        Help about any command
  completion  Generate shell completion script
  man         Generate man pages into a directory or stdout
```

### Common examples

```sh
# Scan current directory
gtree

# Scan a specific path
gtree ~/Projects

# Explicit subcommand form
gtree status ~/Projects

# Only show repos with uncommitted changes, dirtiest first
gtree ~/Projects --only-dirty --sort dirty

# Sort by most recently committed
gtree ~/Projects --sort recent

# Limit scan to 3 levels deep
gtree ~/Projects --depth 3

# Live dashboard — great in a split pane
gtree ~/Projects --watch

# Pipe (auto-disables color)
gtree ~/Projects | grep "↑"

# Open config in $EDITOR
gtree config --edit

# Print config values and file path
gtree config
```

### Shell completions

```sh
# Zsh
gtree completion zsh > ~/.zsh/completions/_gtree

# Bash
gtree completion bash > /etc/bash_completion.d/gtree

# Fish
gtree completion fish > ~/.config/fish/completions/gtree.fish
```

### Manual pages

```sh
# View manual
man gtree

# Or view/generate directly with gtree
gtree man               # print roff man page to stdout
gtree man man/          # generate all man pages into man/
```

---

## ⚙️ Configuration

On first run, `~/.config/gtree/config.yml` is auto-created:

```yaml
# gtree configuration file
# https://github.com/hamimlohani/gtree

# Default path to scan (empty = current directory)
default_path: ""

# Maximum recursion depth (0 = unlimited)
default_depth: 0

# Default sort order: name | dirty | recent
default_sort: name

# Color theme: auto | dark | light
theme: auto

# Directories to skip during scan
ignore_dirs:
  - node_modules
  - .cache
  - vendor
  - target
  - dist
  - build
  - .venv
  - Library
  - .Trash
```

### Environment variable overrides

Every config key can be overridden with `GTREE_<KEY>`:

```sh
GTREE_DEFAULT_SORT=dirty gtree ~/Projects
GTREE_THEME=light        gtree
GTREE_DEFAULT_DEPTH=4    gtree
```

### Edit the config

```sh
gtree config           # show path + current values
gtree config --edit    # open in $EDITOR (falls back to vi)
```

---

## 🏗️ Building from source

```sh
# Show all available make targets
make help

# Build for the local machine
make build

# Run all tests (with race detector)
make test

# Run only fast unit tests (no git subprocess)
make test-short

# Generate HTML coverage report
make cover

# Run go vet + tests (CI-friendly)
make check

# Tidy dependencies
make tidy

# Cross-compile for all 4 platforms → dist/
make build-all

# Build for macOS only
make build-darwin

# Build for Linux only
make build-linux

# Install to ~/.local/bin
make install

# Remove installed binary
make uninstall

# Goreleaser snapshot (dry-run, no publish)
make snapshot

# Create and push a version tag → triggers GitHub Actions release
make tag VERSION=v1.0.0
```

---

## 🚀 Releasing

Releases are fully automated via GitHub Actions:

```sh
# 1. Commit and push your changes
git add .
git commit -m "feat: my new feature"
git push origin main

# 2. Tag and push — this triggers the release workflow
make tag VERSION=v1.2.3
```

The release workflow will:
1. Run all tests (aborts if any fail)
2. Build binaries for `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`
3. Create tarballs for Linux, bare binaries for macOS
4. Generate a `checksums.txt` (SHA-256)
5. Publish to GitHub Releases with a grouped semantic changelog

---

## 🧪 Tests

```sh
make test       # all tests with race detector
make test-short # fast unit tests only (no git subprocess calls)
make cover      # HTML coverage report
```

| Package | Tests |
|---|---|
| `internal/scanner` | `HumanDuration` · `ParsePorcelain` · `ParseAheadBehind` · `ScanFindsRepos` · `MaxDepth` · `IgnoresDir` · `OnlyDirty` · `SortByDirty` · `SortByRecent` |
| `internal/tree` | `BuildFlat` · `BuildNested` · `BuildRoot` |
| `internal/config` | `DefaultsPopulated` · `CreatesFile` · `ReadsExisting` · `IgnoreDirList` |

---

## 📁 Project structure

```
gtree/
├── main.go
├── go.mod / go.sum
├── Makefile                     20 targets, self-documented
├── README.md
├── LICENSE                      MIT
├── install.sh                   POSIX one-liner installer
├── .gitignore
├── .golangci.yml                Lint rules
├── .goreleaser.yml              Cross-platform release config
├── .github/
│   └── workflows/
│       ├── ci.yml               Tests on every push/PR (matrix)
│       └── release.yml          Tag push → test → publish
├── cmd/
│   ├── root.go                  Cobra root + all persistent flags
│   ├── status.go                Scan path resolution → TUI handoff
│   ├── version.go               --version flag + version subcommand
│   └── config_cmd.go            config show/edit subcommand
└── internal/
    ├── config/
    │   ├── config.go            YAML loader, auto-creates on first run
    │   └── config_test.go
    ├── scanner/
    │   ├── scanner.go           Concurrent walk + semaphore worker pool
    │   ├── gitinfo.go           Git shell-outs + pure parsing helpers
    │   ├── scanner_test.go      Integration tests (real temp git repos)
    │   └── gitinfo_test.go      Unit tests for parsers
    ├── theme/
    │   └── theme.go             Catppuccin dark/light palette + NO_COLOR
    ├── tree/
    │   ├── tree.go              Nested Node tree builder
    │   ├── styled_render.go     Lipgloss renderer
    │   ├── render.go            Plain-text fallback renderer
    │   ├── os_helpers.go        Testable os.UserHomeDir wrapper
    │   └── tree_test.go
    └── tui/
        └── model.go             Bubble Tea: spinner → tree → watch mode
```

---

## 🤝 Contributing

1. Fork the repo and clone your fork.
2. `make tidy` — download dependencies.
3. `make check` — run vet + tests.
4. Submit a PR — CI runs automatically on `ubuntu-latest` + `macos-latest` with Go 1.21 and 1.22.

Please follow [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `perf:`, etc.) so the auto-generated changelog stays clean.

---

## 📜 License

MIT © [Hamim Lohani](https://github.com/hamimlohani)
