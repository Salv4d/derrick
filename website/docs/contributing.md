---
layout: default
title: Contributing Guide
---

# 🤝 Contributing Guide

Welcome to the Derrick project! We are building the smartest local control plane.

## Setting Up Your Development Environment

You **DO NOT** need to install Go, linters, or formatting tools on your host machine to contribute. We use Derrick to build Derrick!

1. Ensure you have [Nix](https://nixos.org/download) installed.
2. Install the latest **Derrick CLI** on your system.
3. Clone the repository: `git clone https://github.com/Salv4d/derrick.git`
4. Navigate inside the project root and run `derrick shell`.

*(Derrick will parse our `derrick.yaml`, download the exact `go` compiler structure, `gofumpt`, and `golangci-lint` binaries transparently into your active sandbox).*

You can now run `go test ./...` safely!

## Branching & Commit Strategy

We strictly enforce **Conventional Commits** for generating reliable changelogs.

* `feat: ...` for a new feature (e.g. `feat: add TUI dashboard view`).
* `fix: ...` for resolving bugs.
* `docs: ...` for documentation modifications in `/docs` or `/assets`.
* `chore/refactor: ...` for refactoring or minor changes.

Branch names should follow the format: `type/context-description` (e.g., `feat/nix-shell-flags`).

## Testing Flow

Run tests across the business logic to ensure regressions are avoided:

```bash
go test ./... -v
```

Tests for core sandbox orchestration specifically live in `internal/engine/`. Because they deal with OS boundaries (Docker/Nix context parsing), verify test output carefully if refactoring `docker.go` or `nix.go`.
