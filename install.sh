#!/usr/bin/env sh
# install.sh — one-liner installer for gtree
#
# Usage:
#   curl -sSfL https://raw.githubusercontent.com/hamimlohani/gtree/main/install.sh | sh
#
# Options (env vars):
#   GTREE_VERSION   — specific version to install, e.g. "v1.2.3" (default: latest)
#   GTREE_INSTALL_DIR — install directory (default: /usr/local/bin, fallback ~/.local/bin)
#   GTREE_NO_SUDO   — set to "1" to never use sudo (forces ~/.local/bin)

set -e

REPO="hamimlohani/gtree"
BINARY="gtree"
VERSION="${GTREE_VERSION:-}"
NO_SUDO="${GTREE_NO_SUDO:-0}"

# ── Color helpers ─────────────────────────────────────────────────────────────
if [ -t 1 ] && [ "${NO_COLOR:-}" = "" ]; then
  BOLD='\033[1m'
  GREEN='\033[32m'
  YELLOW='\033[33m'
  RED='\033[31m'
  CYAN='\033[36m'
  RESET='\033[0m'
else
  BOLD=''; GREEN=''; YELLOW=''; RED=''; CYAN=''; RESET=''
fi

info()  { printf "${CYAN}  →${RESET} %s\n" "$*"; }
ok()    { printf "${GREEN}  ✓${RESET} %s\n" "$*"; }
warn()  { printf "${YELLOW}  ⚠${RESET} %s\n" "$*"; }
error() { printf "${RED}  ✗${RESET} %s\n" "$*" >&2; exit 1; }

# ── Detect OS / arch ──────────────────────────────────────────────────────────
detect_platform() {
  OS="$(uname -s)"
  ARCH="$(uname -m)"

  case "$OS" in
    Linux)  OS="linux"  ;;
    Darwin) OS="darwin" ;;
    *)      error "Unsupported OS: $OS" ;;
  esac

  case "$ARCH" in
    x86_64 | amd64)  ARCH="amd64" ;;
    aarch64 | arm64) ARCH="arm64" ;;
    *)               error "Unsupported architecture: $ARCH" ;;
  esac
}

# ── Resolve latest version from GitHub API ────────────────────────────────────
resolve_version() {
  if [ -n "$VERSION" ]; then
    return
  fi
  info "Fetching latest release..."
  if command -v curl >/dev/null 2>&1; then
    VERSION=$(curl -sSfL "https://api.github.com/repos/${REPO}/releases/latest" \
      | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\(.*\)".*/\1/')
  elif command -v wget >/dev/null 2>&1; then
    VERSION=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" \
      | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\(.*\)".*/\1/')
  else
    error "curl or wget is required"
  fi
  if [ -z "$VERSION" ]; then
    error "Could not determine latest version. Set GTREE_VERSION manually."
  fi
}

# ── Choose install directory ──────────────────────────────────────────────────
choose_install_dir() {
  if [ -n "${GTREE_INSTALL_DIR:-}" ]; then
    INSTALL_DIR="$GTREE_INSTALL_DIR"
    return
  fi

  # Prefer /usr/local/bin (needs sudo on most systems).
  if [ "$NO_SUDO" = "0" ] && [ -w /usr/local/bin ]; then
    INSTALL_DIR="/usr/local/bin"
  elif [ "$NO_SUDO" = "0" ] && command -v sudo >/dev/null 2>&1; then
    INSTALL_DIR="/usr/local/bin"
    USE_SUDO=1
  else
    INSTALL_DIR="$HOME/.local/bin"
    warn "/usr/local/bin not writable; installing to $INSTALL_DIR"
    warn "Make sure \$HOME/.local/bin is in your PATH."
  fi
}

# ── Download and install ──────────────────────────────────────────────────────
download_and_install() {
  # Build the download URL based on OS and archive format.
  if [ "$OS" = "darwin" ]; then
    # macOS: bare binary
    FILENAME="${BINARY}-${OS}-${ARCH}"
    URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"
  else
    # Linux: tarball
    FILENAME="${BINARY}-${VERSION#v}-${OS}-${ARCH}.tar.gz"
    URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"
  fi

  TMP_DIR="$(mktemp -d)"
  trap 'rm -rf "$TMP_DIR"' EXIT

  info "Downloading ${BOLD}${BINARY} ${VERSION}${RESET} (${OS}/${ARCH})..."

  if command -v curl >/dev/null 2>&1; then
    curl -sSfL "$URL" -o "$TMP_DIR/$FILENAME" || error "Download failed: $URL"
  else
    wget -qO "$TMP_DIR/$FILENAME" "$URL" || error "Download failed: $URL"
  fi

  # Extract if tarball.
  if echo "$FILENAME" | grep -q '\.tar\.gz$'; then
    tar -xzf "$TMP_DIR/$FILENAME" -C "$TMP_DIR"
    SRC="$TMP_DIR/${BINARY}"
  else
    SRC="$TMP_DIR/$FILENAME"
  fi

  chmod +x "$SRC"

  mkdir -p "$INSTALL_DIR"
  if [ "${USE_SUDO:-0}" = "1" ]; then
    sudo mv "$SRC" "${INSTALL_DIR}/${BINARY}"
  else
    mv "$SRC" "${INSTALL_DIR}/${BINARY}"
  fi
}

# ── Verify ────────────────────────────────────────────────────────────────────
verify() {
  if "${INSTALL_DIR}/${BINARY}" --version >/dev/null 2>&1; then
    INSTALLED_VER=$("${INSTALL_DIR}/${BINARY}" --version 2>&1 | head -1)
    ok "Installed: $INSTALLED_VER"
    ok "Location:  ${INSTALL_DIR}/${BINARY}"
  else
    warn "Binary installed but --version check failed."
  fi
}

# ── PATH hint ─────────────────────────────────────────────────────────────────
print_path_hint() {
  case ":$PATH:" in
    *":${INSTALL_DIR}:"*) return ;;  # already in PATH
  esac

  printf "\n${YELLOW}  ⚠  ${INSTALL_DIR} is not in your PATH.${RESET}\n"
  printf "  Add this to your shell profile (~/.zshrc or ~/.bashrc):\n\n"
  printf "    ${CYAN}export PATH=\"\$PATH:${INSTALL_DIR}\"${RESET}\n\n"
}

# ── Main ──────────────────────────────────────────────────────────────────────
main() {
  printf "\n${BOLD}Installing gtree${RESET}\n\n"

  detect_platform
  resolve_version
  choose_install_dir

  info "Version:     ${VERSION}"
  info "Platform:    ${OS}/${ARCH}"
  info "Install dir: ${INSTALL_DIR}"
  printf "\n"

  download_and_install
  verify
  print_path_hint

  printf "\n${BOLD}${GREEN}Done!${RESET} Run ${CYAN}gtree --help${RESET} to get started.\n\n"
}

main "$@"
