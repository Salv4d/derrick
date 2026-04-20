---
sidebar_position: 2
title: Installation
---

# Installation

Derrick ships a single static binary. Pick whichever install method fits your workflow.

## Supported platforms

| OS / arch | Supported | Notes |
| :--- | :--- | :--- |
| Linux amd64 | Yes | Primary development target |
| Linux arm64 | Yes | Tested on Raspberry Pi OS, Asahi |
| macOS amd64 (Intel) | Yes | Released binary |
| macOS arm64 (Apple Silicon) | Yes | Released binary |
| WSL2 | Yes | Treat as Linux |
| Windows native | No | Use WSL2 |

Prerequisites (install whichever your projects will use — Derrick itself needs neither at install time):

- [Nix](https://nixos.org/download) — for `provider: nix` or `provider: hybrid` projects
- [Docker](https://docs.docker.com/engine/install/) and Docker Compose — for `provider: docker` or `hybrid`

## Recommended: one-line install

Downloads the right binary for your OS + arch and drops it in `/usr/local/bin`:

```bash
curl -fsSL https://raw.githubusercontent.com/Salv4d/derrick/main/install.sh | bash
```

## Other install methods

### Nix flake

```bash
# One-shot (no install)
nix run github:Salv4d/derrick -- --help

# Install into your profile
nix profile install github:Salv4d/derrick
```

### Go install

Builds from source into `$GOBIN` (or `$GOPATH/bin`):

```bash
go install github.com/Salv4d/derrick/cmd/derrick@latest
```

### Pre-built binary

```bash
# Pick the right asset from https://github.com/Salv4d/derrick/releases/latest
curl -L -o derrick https://github.com/Salv4d/derrick/releases/latest/download/derrick-linux-amd64
chmod +x derrick
sudo mv derrick /usr/local/bin/
```

Asset naming: `derrick-<os>-<arch>` where `os` is `linux` or `darwin` and `arch` is `amd64` or `arm64`.

### Build from source

```bash
git clone https://github.com/Salv4d/derrick.git
cd derrick
go build -o derrick ./cmd/derrick
sudo mv derrick /usr/local/bin/
```

Go 1.26+ required (the repo's `go.mod` is the source of truth).

## Verify the install

```bash
derrick version
```

This also pings the GitHub releases API to check for updates. If you're offline, it prints a single warning and exits 0 — safe to use in CI.

## Shell completion

Derrick generates completions for bash, zsh, fish, and powershell via cobra's native support:

```bash
# Bash (Linux)
derrick completion bash | sudo tee /etc/bash_completion.d/derrick > /dev/null

# Zsh
derrick completion zsh > "${fpath[1]}/_derrick"

# Fish
derrick completion fish > ~/.config/fish/completions/derrick.fish

# PowerShell
derrick completion powershell >> $PROFILE
```

Full activation instructions for each shell are in `derrick completion --help`.

## Updating

Derrick can update itself:

```bash
derrick update
```

This downloads the latest release, verifies it runs, and atomically replaces the binary in place. No package manager required.

## Uninstall

```bash
sudo rm /usr/local/bin/derrick
rm -rf ~/.derrick          # hub config and logs
# In each project: rm -rf .derrick
```
