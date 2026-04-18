# Derrick Architecture Notes

*Generated from Phase 1 research into `mise` (Rust) and `devcontainers-cli` (TypeScript/Node.js).*

---

## Findings from Reference Tools

### mise (Rust)

**Key pattern — the `Backend` trait:**
mise defines a single `trait Backend` (in `src/backend/mod.rs`) that every tool provider
(npm, cargo, go, asdf, aqua, etc.) implements. The trait exposes a uniform lifecycle:
`install_version_`, `list_versions_matching`, `exec_env`, `get_dependencies`.

Lessons:
- **Interface-first design** lets the orchestrator core be completely agnostic of the
  underlying technology. Adding a new backend requires zero changes to the CLI layer.
- **State is filesystem-local** — installed tool versions live under
  `~/.local/share/mise/<tool>/<version>/`. The CLI simply checks directory existence to
  determine install state.
- **PATH manipulation over shims** — `hook_env.rs` modifies `PATH` at shell-hook time
  rather than maintaining shim binaries. This keeps the host clean and reduces
  install-time side-effects.
- **CacheManager** wraps any expensive computation (version list fetching, plugin
  resolution) with a TTL-based disk cache, avoiding redundant network calls.

### devcontainers-cli (TypeScript)

**Key pattern — CLI wrapping over Docker API:**
devcontainers-cli does NOT use the Docker Engine API SDK directly. It wraps the `docker`
and `docker compose` binaries via `exec`/`spawn` (`dockerCLI`, `dockerPtyCLI` in
`spec-shutdown/dockerUtils.ts`). This is a deliberate choice for portability — it
supports podman as a drop-in by just changing the binary path.

Lessons:
- **Wrap the CLI, not the API.** Wrapping the Docker binary is simpler, requires no
  version pinning of Docker SDK binaries, and works with Docker-compatible runtimes
  (Podman, nerdctl) for free.
- **State via container labels** — devcontainers attaches metadata labels to running
  containers (`devcontainer.local_folder`, `devcontainer.config_file`). This makes
  state inspectable with a plain `docker inspect`.
- **`ProvisionOptions` → `launch()` pipeline** — a typed options struct is built from
  CLI flags and config, then threaded through a single `launch()` function that
  delegates to the right provider (single-container vs. compose). Clean separation
  between flag parsing and execution.
- **Feature composition** — each dev environment "feature" is a self-contained
  directory with `install.sh` + `devcontainer-feature.json`. Features are layered as
  Dockerfile stages. This is how zero-to-environment happens without users writing
  Dockerfiles.

---

## Derrick's Architecture: The Supreme Orchestrator Pattern

### Guiding Principle

Derrick is NOT a package manager. It is a **lifecycle orchestrator** that translates a
human-readable `derrick.yml` contract into calls to proven underlying technologies
(Docker, Nix, DevContainers). The user never sees a Dockerfile, a flake.nix, or a
docker-compose override. Derrick hides all of that.

---

### Layer Model

```
┌─────────────────────────────────────────────┐
│              CLI (Cobra commands)            │  ← User-facing layer
│  start / stop / shell / run / doctor / init  │
└────────────────────┬────────────────────────┘
                     │ typed config + flags
┌────────────────────▼────────────────────────┐
│           Orchestrator Core                  │  ← Business logic
│  hook executor · state manager · resolver    │
└──────┬─────────────────────┬────────────────┘
       │                     │
┌──────▼──────┐    ┌─────────▼──────────┐
│  Provider   │    │    Provider         │
│  (Docker)   │    │    (Nix)            │   ← Swappable backends
└──────┬──────┘    └─────────┬──────────┘
       │                     │
  docker CLI              nix CLI            ← OS-level tools (never re-implemented)
```

---

### The `Provider` Interface (Go)

Every environment backend implements one interface:

```go
// Provider abstracts an environment backend (Docker, Nix, DevContainer, etc.).
type Provider interface {
    // Name returns the human-readable backend name.
    Name() string

    // IsAvailable checks whether the backend's required tooling is installed.
    IsAvailable() error

    // Setup performs first-time environment setup (builds images, resolves flake, etc.).
    Setup(ctx context.Context, cfg *config.ProjectConfig, flags Flags) error

    // Start activates the environment (starts containers, enters nix shell, etc.).
    Start(ctx context.Context, cfg *config.ProjectConfig, flags Flags) error

    // Stop tears down the environment gracefully.
    Stop(ctx context.Context, cfg *config.ProjectConfig) error

    // Shell opens an interactive shell inside the managed environment.
    Shell(ctx context.Context, cfg *config.ProjectConfig) error

    // Status returns the current operational state.
    Status(ctx context.Context, cfg *config.ProjectConfig) (EnvironmentStatus, error)

    // Logs streams runtime logs for the active environment.
    Logs(ctx context.Context, cfg *config.ProjectConfig) error
}
```

This interface means the CLI layer never branches on "is this Docker or Nix" — it just
calls `provider.Start(...)` and the backend does the right thing.

---

### State Management

Derrick tracks per-project state in `.derrick/state.json`:

```json
{
  "project": "my-api",
  "provider": "docker",
  "status": "running",
  "first_setup_completed": true,
  "started_at": "2026-04-18T12:00:00Z",
  "container_ids": ["abc123"],
  "flags_used": ["--seed-db"]
}
```

This enables:
- **Conditional hook execution** — hooks marked `when: first-setup` only fire when
  `first_setup_completed` is false in state.
- **`derrick doctor`** — can inspect state and validate it against what Docker/Nix
  reports, surfacing drift.
- **Future web dashboard** — an HTTP server reads `state.json` to render the
  management UI without re-querying Docker/Nix every time.

State is stored within the project directory (never globally) to stay sandbox-safe.

---

### Declarative Config (`derrick.yml`) — Redesigned Schema

The config focuses on **what the project needs**, not **how to achieve it**:

```yaml
name: my-api
version: "1.0.0"

# Which isolation backend to use. "auto" picks Docker if available, falls back to Nix.
provider: docker   # docker | nix | devcontainer | auto

# Docker-specific config (only read when provider: docker or auto)
docker:
  compose: ./docker-compose.yml
  profiles: [dev]
  network: derrick-net    # Derrick creates and manages this network

# Nix-specific config (only read when provider: nix or auto)
nix:
  registry: github:NixOS/nixpkgs/nixos-unstable
  packages:
    - go
    - nodejs_22
    - postgresql_16

# Ports this project exposes (used by future dashboard + doctor checks)
ports:
  - 3000    # API
  - 5432    # PostgreSQL

# Lifecycle hooks with execution conditions
hooks:
  start:
    - run: "echo 'Booting…'"
      when: always
    - run: "go mod download"
      when: first-setup          # Only on first `derrick start`
    - run: "make seed-db"
      when: flag:seed-db         # Only when `derrick start --seed-db` is passed
  stop:
    - run: "echo 'Bye!'"
      when: always
  restart:
    - run: "make reset-migrations"
      when: flag:reset

# Custom flags exposed by this project (shown in `derrick start --help`)
flags:
  reset:
    description: "Rebuild the environment and replay migrations from scratch"
  seed-db:
    description: "Inject initial seed data after environment boots"

# Required sibling projects (Derrick clones + starts them automatically)
requires:
  - alias: auth-service

# Environment variables this project expects
env:
  DATABASE_URL:
    description: "PostgreSQL connection string"
    required: true
    default: "postgres://localhost:5432/myapi_dev"
  API_SECRET:
    required: true
```

---

### Error Translation Layer

Because Derrick wraps external processes, all `exec.Command` invocations go through
`internal/engine/executor.go`, which:
1. Captures `stderr` in a `bytes.Buffer`.
2. On non-zero exit, classifies the error with a `matchError()` function that tests
   against a table of known patterns (Docker socket denied, port conflict, image not
   found, Nix flake eval error, etc.).
3. Returns a `DerrickError` struct with `Message`, `Fix`, and optionally `DocsURL`.
4. The UI layer renders this with lipgloss as a rich diagnostic panel.

```
✖  CRITICAL ERROR
   Docker socket permission denied.

   Your user does not have access to the Docker daemon.

   Fix:  sudo usermod -aG docker $USER && newgrp docker

   Docs: https://docs.derrick.dev/errors/docker-socket
```

---

### API Surface for Future Web Dashboard

The orchestrator core is written as a pure library (`internal/engine/`) with no
stdout coupling. Commands in `cmd/derrick/` call the library and pipe results to the
UI renderer. This means a future `derrick serve` command can expose the same library
over an HTTP API (JSON responses instead of lipgloss output) without rewriting
business logic.

Proposed API shape:
- `GET  /api/projects`        — list known projects (from state files)
- `POST /api/projects/:id/start` — start an environment
- `POST /api/projects/:id/stop`  — stop an environment
- `GET  /api/projects/:id/status` — current status + container info
- `GET  /api/projects/:id/logs`   — SSE stream of logs

---

### Implementation Roadmap

| Phase | Deliverable |
|-------|-------------|
| 1 (now) | `Provider` interface + `DockerProvider` + `NixProvider` |
| 2 | State manager (`internal/state/`) |
| 3 | Advanced hook executor with `when:` conditions + custom flags |
| 4 | Error translation layer (`internal/engine/executor.go`) |
| 5 | `derrick start [alias]` with Hub resolution |
| 6 | `derrick serve` — HTTP API for web dashboard |
