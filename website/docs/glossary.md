---
layout: default
title: Glossary
---

# Glossary

Precise definitions used throughout Derrick's codebase and documentation.

---

**Provider** — A Go interface (`internal/engine/provider.go`) that every isolation backend implements. Current implementations: `DockerProvider` (wraps `docker compose`), `NixProvider` (wraps `nix develop`), and `HybridProvider` (composes both — containers for services, nix for the language toolchain). The CLI layer is completely agnostic of which provider is active.

**Supreme Orchestrator** — Derrick's core design pattern: act as a high-level coordinator that delegates to proven tools (Docker, Nix) rather than reimplementing their capabilities. Derrick's value is developer experience and abstraction, not package management.

**Sandbox** — The ephemeral `nix develop` environment spawned by the Nix provider. Dependencies are strictly isolated from the host OS PATH.

**State Contract** — The `derrick.yaml` file. The source of truth for what constitutes a valid environment.

**State File** — `.derrick/state.json`. Persisted runtime metadata: which provider is active, whether first-setup completed, container IDs, and which custom flags were used on the last start. Powers conditional hooks and the future dashboard.

**Engine** — The core Go logic layer (`internal/engine/`). Contains the Provider interface, provider implementations, the hook executor, the error translation layer, Nix flake generation, and Docker network orchestration.

**Hook Condition** — The `when:` field on a lifecycle hook entry. Possible values: `always` (default, every run), `first-setup` (only before `first_setup_completed` is persisted in state), `flag:<name>` (only when `--flag <name>` is passed to `derrick start`).

**Hub** — The global project registry at `~/.derrick/config.yaml`. Maps short aliases (`auth-service`) to Git URLs. `derrick start <alias>` clones and boots Hub-registered projects in one command.

**Host Pollution** — The anti-pattern of installing global tool versions directly on the developer's OS (e.g., `nvm`, global `go install`, `pyenv`). Derrick eliminates host pollution by isolating all project dependencies inside the Provider's environment.

**`com.derrick.managed`** — Docker label applied to every service, network, and volume Derrick creates (via the generated compose override). Scopes `derrick clean` so prune operations only touch derrick-managed resources and never another project's containers.

**Error Translation** — The process in `internal/engine/executor.go` that catches raw stderr from wrapped subprocesses, matches it against known patterns, and returns a structured `DerrickError` with a human-readable `Fix` message instead of a raw exit code.
