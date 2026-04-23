# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.5.5] — 2026-04-21

### Added
- Derrick Hub management: new `hub` command suite with `add`, `list`, and `remove` subcommands to manage project aliases and their Git URLs in `~/.derrick/config.yaml`.
- Remote Hub Subscriptions: added `hub subscribe` and `hub unsubscribe` to point Derrick to external YAML indices. `derrick start [alias]` now automatically checks all subscribed remotes if an alias isn't found locally.
- Automatic registration: `derrick start --register` now automatically detects the project's Git remote URL and adds it to the local Hub.

## [0.5.4] — 2026-04-21

### Added
- Cycle detection: `derrick start` now detects circular dependencies in `requires:` chains and aborts early to prevent infinite recursion.

### Fixed
- Robust binary resolution: replaced `os.Args[0]` with `os.Executable()` for all recursive project calls, ensuring reliability in CI and when running via `go run`.
- Linting & Formatting: fixed all `gofmt` issues and removed several unused functions and types to satisfy strict CI checks.

## [0.5.3] — 2026-04-21

### Added
- Recursive dependency stop: `derrick stop` now recursively shuts down all project dependencies defined in `requires:` in parallel. Includes cycle detection to prevent infinite recursion.
- Standardized core logging: all `start` and `stop` operations now consistently mirror output to `.derrick/last.log` for future debugging and dashboard integration.

### Fixed
- Hardened error handling: addressed several instances of ignored errors from `os.Getwd()` and `json.MarshalIndent` across the CLI.
- `derrick shell` feedback: improved error message when running shell/command before the project is initialized, suggesting `derrick start` first.
- `derrick doctor` precision: now correctly handles empty projects and avoids reporting irrelevant failures for projects that don't use Nix.
- Test suite regressions: fixed build and assertion errors in `hooks_test.go`, `nix_test.go`, and `config_test.go` following recent schema updates.
- `.env` file management: fixed a bug where missing `.env` files were not correctly initialized during the validation phase.

## [0.5.2] — 2026-04-21

### Fixed
- `derrick shell` (Docker): added `-T` flag for non-interactive commands to prevent "the input device is not a TTY" errors in CI or scripts.
- `derrick shell` (Docker): improved TTY handling for interactive sessions by ensuring direct inheritance of standard input/output/error streams.

## [0.5.1] — 2026-04-21

### Added
- Multi-condition hooks: the `when:` field now accepts an array of triggers (e.g. `when: ["first-setup", "flag:seed"]`). The hook executes if any condition is met.

### Fixed
- `derrick stop` reliability: the command now explicitly includes the `.derrick/docker-compose.override.yml` during teardown, ensuring that auto-wired shared networks and labels are correctly cleaned up.

## [0.5.0] — 2026-04-21

### Added
- `docker.networks` field: list external Docker networks every service in the project joins. Derrick creates any missing ones on start (labelled `com.derrick.managed=true`) so `derrick clean` still prunes them safely. Opt-in escape hatch from per-project network isolation.
- `requires` entries now accept a struct form with a `connect` option (plain string form shorthand = `connect: true`). When `connect: true`, Derrick creates a shared network `derrick-<parent>` and auto-wires both sides via `DERRICK_JOIN_NETWORK` in the dependency subprocess — no config required in the dependency.
- Recursive parallel booting: `derrick start` now boots all project dependencies in parallel using `errgroup`, significantly reducing startup time for large project trees.
- `derrick run` now supports inline command execution via the `--` separator (e.g. `derrick run nodejs -- npm test`).
- Improved `derrick init` wizard: now explicitly asks for the isolation backend (Hybrid, Docker Only, or Nix Only) when both Compose and a language toolchain are detected.
- Expanded `derrick doctor` audit: now verifies Docker daemon connectivity, Docker Compose file syntax, existence of environment base files, and presence of all required projects.
- Five new recipe docs: Sentry, Keycloak, Plane, n8n, Meilisearch.
- `TestRecipes_Parse` — every full `derrick.yaml` block embedded in `website/docs/use_cases/*.md` is parsed against the current schema on every CI run.

### Changed
- Nix evaluation optimization: `EnsureNixEnvironment` now skips writing `flake.nix` if the content is unchanged, and `BootEnvironment` skips the verification phase if a `flake.lock` exists and the flake is up-to-date.
- Strict `derrick.yaml` validation: the project contract is now strictly validated on load using `validator/v10` tags, catching invalid provider names or project aliases immediately.

### Fixed
- JSON Error Feedback: Commands using `--json` now return a properly formatted JSON error object on failure instead of decorative text.
- `derrick run` reliability: fixed `os.Chdir` timing so that cleanup hooks still fire correctly on early failures.
- `docker compose` was invoked without `-p cfg.Name`, so containers inherited their project name from the cwd basename. Now `-p cfg.Name` is passed on all compose operations.
- Existing recipes (ghost, grafana, plausible, supabase, appwrite) rewritten to match the current `derrick.yaml` schema.

## [0.4.1] — 2026-04-20

### Fixed
- CI: bumped golangci-lint from v2.1.6 to v2.11.4 and lowered `go.mod` to `go 1.25.0` so the linter can load the config (v2.1.6 was built with go1.24 and rejects go1.25+ targets).

## [0.4.0] — 2026-04-20

### Added
- `Provider.Provision(cfg)` splits environment materialization (writing `.derrick/flake.nix`, resolving the nix registry, writing the compose override) away from `Start()`, which now only boots long-running services. This is the fix for `flake.nix does not exist` errors when setup hooks ran before the provider had a chance to generate the flake.
- Five-stage lifecycle hooks: `before_start` (host), `setup` (sandbox, services down), `after_start` (sandbox + services up), `before_stop` (sandbox + services up), `after_stop` (host). Replaces the single `hooks.start` / `hooks.stop` pair.
- Architecture docs: Lifecycle Stages section with start/stop flowchart and a per-provider shell matrix.

### Changed
- BREAKING: `hooks.start` / `hooks.stop` / `hooks.restart` are gone. Move `npm install`/`go mod download`-style work to `hooks.setup`; move DB seeding / warmup to `hooks.after_start`; move graceful drain / DB dumps to `hooks.before_stop`. This is a pre-1.0 breaking change — no migration shim.
- `HybridProvider.Start` no longer calls nix — nix has no long-running services. Provisioning still runs both legs with nix first so a bad package name aborts before docker is touched.
- `NixProvider.Start` is now a no-op; all heavy lifting moved to `Provision`.

### Fixed
- Setup-style hooks (`npm install`, `go mod download`) no longer fail with `path '.derrick/flake.nix' does not exist` on the first boot after deleting `.derrick/`. Provision now runs before any sandbox hook fires.

## [0.3.0] — 2026-04-20

### Added
- Scoped `.golangci.yml` (govet, ineffassign, staticcheck, unused, gofmt) wired into PR CI via `golangci-lint-action@v8`, plus an explicit `go vet` step in the test job.
- `derrick completion [bash|zsh|fish|powershell]` built on cobra's native generators, with activation docs in the long help.
- Table-driven unit tests for `ResolveProvider`, `HybridProvider` composition, hook `when:` conditions, `GenerateNetworkOverride` service labelling, and the `state.Load` non-nil contract.
- Dedicated docs for the Hybrid provider and multi-project behavior (state lock, per-project networks, port-conflict stance, shared `/nix/store`, cycle detection).

### Changed
- `Provider.Shell(cfg, args)` now takes trailing arguments; docker backend forwards them to `compose exec`, nix backend forwards them to `nix develop --command`. `derrick shell` delegates to the resolved provider instead of hard-coding nix — docker-only projects finally get a working interactive shell.
- `HybridProvider` is now composed of two narrow `providerLeg`s (testable without a daemon). `IsAvailable` joins both leg errors with `errors.Join`; `Status` aggregates both legs without short-circuiting on a single-leg failure; `NewHybridProvider()` is the public constructor used by `ResolveProvider`.
- `state.Load` always returns a non-nil `*EnvironmentState` (zeroed with `Status: StatusUnknown`) even on error, so the `projectState, _ := state.Load(cwd)` idiom used by `stop`/`status` can no longer nil-deref.

### Fixed
- `NixProvider.Status` now requires `.derrick/flake.nix` on disk before reporting the environment as ready. Previously it returned `Running: true` whenever the `nix` binary existed, even before `derrick start`.
- `derrick doctor` exits non-zero when `report.Issues > 0`, in both text and `--json` modes, so CI scripts can gate on exit code.

### Removed
- Orphaned `StartContainers` / `StopContainers` helpers in `internal/engine/docker.go` (superseded by `DockerProvider.Start/Stop` since the provider refactor).

## [0.2.0] — 2026-04-19

### Added
- `--json` output for `derrick status`, `derrick doctor`, and `derrick version` so the CLI can be driven programmatically.
- `--dry-run` on `derrick start` and `derrick clean` to preview provider actions without mutating state.
- Nix flake at `packaging/nix/` so derrick itself can be installed via `nix profile install`.
- One-line installer script that downloads the correct platform binary from a GitHub release.
- Schema versioning on `derrick.yaml` (`CurrentSchema = 1`), surfaced by doctor when the config is from a future version.
- `derrick update` command that downloads the latest release and atomically replaces the running binary.
- `derrick status` command reporting environment state, active provider, and hook flags.
- Per-entry pinning for EOL nix packages, replacing the old whole-project pin hack.
- Nightly integration CI covering docker start/stop on real compose projects.

### Changed
- Release pipeline now builds darwin `amd64` and `arm64` binaries in addition to linux `amd64` / `arm64`.
- Removed the unfinished TUI dashboard — `derrick status` covers the same need without the maintenance weight.

### Fixed
- `derrick doctor` no longer mutates `.derrick/` during its audit — reports are pure observations.
- Nix flake generation now detects the host OS/arch instead of hard-coding `x86_64-linux`, unbreaking macOS and ARM.
- `derrick clean` scopes docker prune to resources labelled `com.derrick.managed=true` instead of blowing away every dangling container on the host.
- `derrick stop` runs stop hooks before provider teardown so hooks can still reach the running services.
- `NixPackage` marshals as a plain YAML scalar when `Registry` is empty, restoring round-trip fidelity for hand-written `derrick.yaml` files.
- Nix legacy package resolution and the EOL-package index now handle a wider range of nixpkgs history.

## [0.1.1] — 2026-04-19

### Added
- `.envrc` emission on first `derrick start` for direnv integration.
- Per-project shell command history persisted under `.derrick/`.

### Fixed
- `NixPackage` YAML marshalling (also shipped in 0.2.0 via fix forward).
- Nix legacy package resolution and an expanded EOL index.

### Docs
- Fixed stale `docker.network` reference in the API docs; documented `docker.shell`; removed the obsolete `ports` field.

## [0.1.0] — 2026-04-19

First public release.

### Added
- Provider interface with independent Docker and Nix backends, plus a Hybrid provider that composes both.
- Conditional hook execution via `when:` (`always`, `first-setup`, `flag:<name>`).
- Environment state persistence at `.derrick/state.json` with file-locking against concurrent writes.
- Redesigned `derrick.yaml` schema with top-level `provider`, `docker`, and `nix` keys.
- `derrick start` / `derrick stop` rewritten on top of the Provider interface.

### Security
- Replaced `bash -c` executor with a safe shlex-based dispatcher to prevent command injection through hook strings.
- Removed the global `derrick-net` to enforce per-project network isolation.
- Added cycle detection via `DERRICK_START_CHAIN`, atomic `.env` writes, and fail-fast behaviour on compose overlay errors.

### Fixed
- `derrick shell` no longer hardcodes a service name; `docker.shell` is now configurable.
- Hook flags are restored on stop so `first-setup` stays honest across restarts.

[Unreleased]: https://github.com/Salv4d/derrick/compare/v0.5.5...HEAD
[0.5.5]: https://github.com/Salv4d/derrick/compare/v0.5.4...v0.5.5
[0.5.4]: https://github.com/Salv4d/derrick/compare/v0.5.3...v0.5.4
[0.5.3]: https://github.com/Salv4d/derrick/compare/v0.5.2...v0.5.3
[0.5.0]: https://github.com/Salv4d/derrick/compare/v0.4.1...v0.5.0
[0.4.1]: https://github.com/Salv4d/derrick/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/Salv4d/derrick/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/Salv4d/derrick/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/Salv4d/derrick/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/Salv4d/derrick/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/Salv4d/derrick/releases/tag/v0.1.0
