---
layout: default
title: Contributing Guide
---

# 🤝 Contributing Guide

Welcome to the Derrick project! We are building the smartest local control plane.

## Setting Up Your Development Environment

You will need the host prerequisites to develop Derrick itself.

1. Ensure you have **Go 1.26.1** installed (see `go.mod`).
2. Clone the repository and download dependencies using `go mod download`.

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
