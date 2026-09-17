# 🌳 gtree

> Scan your directory tree for git repositories and see their live status at a glance — like `gh` meets `tree`.

```
 🌳 gtree  ~/Developer  35 repos

├── Arduino-Project/
│   ├── ESP32_SmartBus_Complete   [main]  ✓          6mo ago
│   └── SmartHome-esp32-code      [main]  ✗ ~1       2mo ago
├── Shell/
│   └── Project/
│       ├── gtree                 [main]  ✗ ~7  ↑2   just now
│       └── run-cplex             [main]  ✓          8mo ago
└── Websites/
    └── portfolio                 [main]  ✓     ↑1   5d ago

────────────────────────────────────────────────────────────────────────────────
  35 repos  •  24 clean  •  11 dirty  •  2 unpushed ↑
```

> **Demo GIF** — _drop a recording here (e.g. made with [vhs](https://github.com/charmbracelet/vhs))_

---

## Features

| Feature | Details |
|---|---|
| **Concurrent scanning** | Worker pool of goroutines — 35 repos in ~1 second |
| **Nested tree view** | Repos grouped by their actual folder hierarchy |
| **Live status** | Branch · dirty/clean · ahead/behind · stash count · last commit time |
| **Spinner** | Animated indicator while scanning |
| **Watch mode** | `--watch` live-refreshes every 5 seconds in alt-screen |
| **Smart filtering** | `--only-dirty` · `--sort dirty\|recent\|name` |
| **Catppuccin palette** | Auto light/dark detection; respects `NO_COLOR` and piped output |
| **Config file** | `~/.config/gtree/config.yml` — auto-created with sane defaults |
| **Cross-platform** | macOS (amd64/arm64) · Linux (amd64/arm64) |

---

## Installation

### Homebrew (macOS / Linux)

```sh
brew install hamimlohani/tap/gtree
```

### Pre-built binary (fastest)

Download from the [Releases page](https://github.com/hamimlohani/gtree/releases):

```sh
# macOS Apple Silicon
curl -L https://github.com/hamimlohani/gtree/releases/latest/download/gtree-darwin-arm64 \
  -o /usr/local/bin/gtree && chmod +x /usr/local/bin/gtree

# macOS Intel
curl -L https://github.com/hamimlohani/gtree/releases/latest/download/gtree-darwin-amd64 \
  -o /usr/local/bin/gtree && chmod +x /usr/local/bin/gtree

# Linux amd64
curl -L https://github.com/hamimlohani/gtree/releases/latest/download/gtree-linux-amd64 \
  -o /usr/local/bin/gtree && chmod +x /usr/local/bin/gtree
```

### From source (requires Go 1.21+)

```sh
git clone https://github.com/hamimlohani/gtree.git
cd gtree
make install        # builds and copies to ~/.local/bin/gtree
```

### go install

```sh
go install github.com/hamimlohani/gtree@latest
```

---

## Usage

```
Usage:
  gtree [path] [flags]
  gtree [command]

Examples:
  gtree                        Scan current directory
  gtree ~/Projects             Scan a specific path
  gtree status ~/Projects      Explicit subcommand form

Available Commands:
  status      Scan a directory and display git repository status
  config      Show or edit the gtree config file
  version     Print version information

Flags:
  -d, --depth int       Maximum recursion depth (0 = unlimited)
  -s, --sort string     Sort order: dirty | recent | name
      --only-dirty      Hide clean repositories
      --watch           Live-refreshing view (refreshes every 5s)
      --no-color        Disable color output
  -c, --config string   Config file path
  -h, --help            Help
  -v, --version         Version
```

### Examples

```sh
# Scan Projects, show only repos with changes, sort dirtiest first
gtree ~/Projects --only-dirty --sort dirty

# Sort by most recently committed
gtree ~/Projects --sort recent

# Live-refresh dashboard — great in a split terminal pane
gtree ~/Projects --watch

# Limit scan to 3 levels deep
gtree ~/Projects --depth 3

# Plain text output (piping auto-disables color)
gtree ~/Projects | grep dirty

# Open the config file in $EDITOR
gtree config --edit

# Print config location and current values
gtree config
```

### Watch mode key bindings

| Key | Action |
|---|---|
| `r` | Manual refresh now |
| `q` / `Esc` | Quit |

---

## Configuration

On first run gtree creates `~/.config/gtree/config.yml` with these defaults:

```yaml
# gtree configuration file
# https://github.com/hamimlohani/gtree

default_path: ""          # default scan path (empty = cwd)
default_depth: 0          # 0 = unlimited recursion
default_sort: name        # name | dirty | recent
theme: auto               # auto | dark | light
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

Edit it with:

```sh
gtree config --edit          # opens in $EDITOR
```

Any config value can be overridden per-run with a flag or an env var:

```sh
GTREE_DEFAULT_SORT=dirty gtree ~/Projects
```

---

## Building from source

```sh
# Build for current platform
make build

# Build for all platforms (output in dist/)
make build-all

# Run tests
make test

# Run linter
make lint

# Install globally
make install
```

### Cross-compile manually

```sh
GOOS=linux  GOARCH=arm64 go build -o gtree-linux-arm64  .
GOOS=darwin GOARCH=arm64 go build -o gtree-darwin-arm64 .
```

---

## Releases

Releases are built automatically by GitHub Actions on tag push:

```sh
git tag v1.0.0
git push origin v1.0.0
```

See [`.github/workflows/release.yml`](.github/workflows/release.yml) and [`.goreleaser.yml`](.goreleaser.yml).

---

## Contributing

1. Fork and clone your fork.
2. `make tidy` — download dependencies.
3. `make test` — run the test suite.
4. Submit a PR — CI runs tests + lint automatically.

---

## License

MIT © Hamim Lohani
