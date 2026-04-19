#!/usr/bin/env bash
# Derrick installer.
#
#   curl -fsSL https://raw.githubusercontent.com/Salv4d/derrick/main/install.sh | bash
#
# Env overrides:
#   DERRICK_VERSION   pin a specific tag (e.g. v0.3.0). Defaults to "latest".
#   PREFIX            install dir. Defaults to /usr/local/bin (or ~/.local/bin
#                     if /usr/local/bin is not writable).
set -euo pipefail

REPO="Salv4d/derrick"
VERSION="${DERRICK_VERSION:-latest}"

err() { printf '\033[1;31merror:\033[0m %s\n' "$*" >&2; exit 1; }
info() { printf '\033[1;34m::\033[0m %s\n' "$*"; }

uname_s=$(uname -s)
uname_m=$(uname -m)

case "$uname_s" in
    Linux)  os=linux ;;
    Darwin) os=darwin ;;
    *) err "unsupported OS: $uname_s" ;;
esac

case "$uname_m" in
    x86_64|amd64) arch=amd64 ;;
    arm64|aarch64) arch=arm64 ;;
    *) err "unsupported arch: $uname_m" ;;
esac

asset="derrick-${os}-${arch}"

if [ "$VERSION" = "latest" ]; then
    url="https://github.com/${REPO}/releases/latest/download/${asset}"
else
    url="https://github.com/${REPO}/releases/download/${VERSION}/${asset}"
fi

# Pick install dir.
if [ -n "${PREFIX:-}" ]; then
    dest_dir="$PREFIX"
elif [ -w /usr/local/bin ] 2>/dev/null; then
    dest_dir=/usr/local/bin
elif [ "$(id -u)" -eq 0 ]; then
    dest_dir=/usr/local/bin
else
    dest_dir="$HOME/.local/bin"
    mkdir -p "$dest_dir"
fi
dest="$dest_dir/derrick"

info "Downloading $asset from $url"
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT

if command -v curl >/dev/null 2>&1; then
    curl -fL --progress-bar -o "$tmp" "$url"
elif command -v wget >/dev/null 2>&1; then
    wget -O "$tmp" "$url"
else
    err "need curl or wget"
fi

chmod +x "$tmp"

if [ -w "$dest_dir" ]; then
    mv "$tmp" "$dest"
else
    info "Need sudo to write $dest"
    sudo mv "$tmp" "$dest"
fi
trap - EXIT

info "Installed: $dest"

case ":$PATH:" in
    *":$dest_dir:"*) ;;
    *) info "Note: $dest_dir is not in your PATH. Add it to your shell rc to run 'derrick'." ;;
esac

"$dest" --version 2>/dev/null || "$dest" version || true
