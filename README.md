# gtree

> Scan your directory tree for git repositories and see their status at a glance — like `gh` meets `tree`.

```
gtree — ~/Projects  (12 repos found)
────────────────────────────────────────────────────────────
└── Projects/
    ├── go/
    │   ├── gtree                   [main]   ✗ ~3  ↑2  2h ago
    │   └── api-client              [main]   ✓       4d ago
    ├── iot/
    │   ├── firmware                [dev]    ✗ ~1       1w ago
    │   └── dashboard               [main]   ✓       3d ago
    └── web/
        └── portfolio               [main]   ✓       1mo ago
────────────────────────────────────────────────────────────
  12 repos  •  9 clean  •  3 dirty  •  2 unpushed
```

## Installation

### Homebrew (macOS / Linux)

```sh
brew install hamimlohani/tap/gtree
```

### Pre-built binaries

Download the latest binary for your platform from the
[Releases page](https://github.com/hamimlohani/gtree/releases).

```sh
# macOS arm64 (Apple Silicon)
curl -L https://github.com/hamimlohani/gtree/releases/latest/download/gtree-darwin-arm64 \
  -o /usr/local/bin/gtree && chmod +x /usr/local/bin/gtree
```

### From source (requires Go 1.21+)

```sh
git clone https://github.com/hamimlohani/gtree.git
cd gtree
make install   # installs to ~/.local/bin/gtree
```

## Usage

```
gtree [path]                     Scan current directory (or path)
gtree status ~/Projects          Explicit subcommand form

Flags:
  -d, --depth N        Maximum recursion depth (0 = unlimited)
  -s, --sort ORDER     Sort: dirty | recent | name  (default: name)
      --only-dirty     Hide clean repositories
      --watch          Live-refreshing view (polls every 5s)
      --no-color       Disable color output

Subcommands:
  gtree status [path]  Same as gtree [path]
  gtree config         Show / edit config file
  gtree version        Print version
  gtree help           Full help text
```

### Examples

```sh
# Scan Projects directory, show only repos with changes, sort by dirtiest
gtree ~/Projects --only-dirty --sort dirty

# Live-refresh view — great to leave in a split terminal pane
gtree --watch

# Set config defaults
gtree config --edit
```

## Configuration

On first run gtree creates `~/.config/gtree/config.yml`:

```yaml
# gtree configuration file
default_path: ""          # default scan path (empty = cwd)
default_depth: 0          # 0 = unlimited
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

## Demo

> _Drop a GIF here once you record one with [vhs](https://github.com/charmbracelet/vhs)_

## Contributing

1. Fork the repo and `git clone` your fork.
2. `make tidy` to download dependencies.
3. `make test` to run the test suite.
4. Submit a PR — CI will run tests + lint for you.

## License

MIT © Hamim Lohani
