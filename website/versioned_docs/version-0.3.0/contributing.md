---
sidebar_position: 9
title: Contributing
---

# Contributing

Derrick is MIT-licensed and open to contributions. This guide covers everything you need to land a PR — from bootstrapping a dev environment to running tests to adding a new provider backend.

## Setting up a dev environment

Derrick builds Derrick: you don't need to install Go, golangci-lint, or any other toolchain on your host OS.

```bash
# One-time
git clone https://github.com/Salv4d/derrick.git
cd derrick
derrick shell

# Inside the sandbox:
go build ./cmd/derrick
go test ./...
```

The repo's own `derrick.yaml` pins the exact Go compiler, `gofumpt`, and `golangci-lint` versions via Nix. Your host `$PATH` is untouched.

If you don't have Derrick installed yet, bootstrap with `nix develop` directly:

```bash
nix develop .derrick/#default   # or run derrick init + derrick start on a fresh clone
```

## Running tests

```bash
# Unit + integration, race detector, no cache
go test -race -count=1 ./...

# Just one package
go test -race ./internal/engine/...

# Verbose
go test -v ./internal/engine/...
```

CI runs:

- `go vet ./...`
- `go test -v -race -count=1 ./...`
- `go build -o /dev/null ./cmd/derrick`
- `golangci-lint run ./...` (config in `.golangci.yml`)

Run `golangci-lint run ./...` locally before pushing — CI will reject PRs that fail lint.

## Project layout

```
cmd/derrick/              # CLI entry points (Cobra commands, one file per subcommand)
internal/config/          # derrick.yaml parser + schema
internal/engine/          # Provider interface, Docker/Nix/Hybrid backends,
                          # hook executor, error translation
internal/state/           # .derrick/state.json persistence with file locking
internal/discovery/       # project auto-detection heuristics for `derrick init`
internal/ui/              # lipgloss/huh output helpers
tests/                    # integration tests
website/                  # Docusaurus docs site
```

Start in `internal/engine/provider.go` to understand the core abstraction — everything else sits around the Provider interface.

## Commit conventions

We enforce [Conventional Commits](https://www.conventionalcommits.org/) so `CHANGELOG.md` can be generated mechanically:

| Prefix | Use for |
| :--- | :--- |
| `feat:` | New user-facing feature |
| `fix:` | Bug fix |
| `refactor:` | Code restructure without behavior change |
| `docs:` | Documentation only |
| `test:` | Test-only changes |
| `chore:` | Tooling, dependencies, release bumps |
| `ci:` | CI pipeline changes |

Scope the message with parens when it helps: `feat(nix): ...`, `fix(state): ...`.

Branch naming: `type/short-description` (e.g., `feat/podman-provider`, `fix/state-lock-race`).

One atomic commit per logical change. If a PR contains feat + refactor + test, split them.

## Pull requests

Before opening a PR:

1. **Rebase on `main`** — no merge commits.
2. **Run the full test + lint suite** locally.
3. **Add tests** for new behavior. Pure functions with table-driven tests are preferred; see `internal/engine/hybrid_provider_test.go` for the stub-based injection pattern we use to test providers without spawning daemons.
4. **Update `CHANGELOG.md`** under `[Unreleased]` if the change is user-visible.
5. **Update docs** in `website/docs/` if you changed CLI behavior or `derrick.yaml` schema.

PR titles should be valid Conventional Commit messages — the squash-merge uses the title as the commit message.

## Adding a new Provider backend

Providers are the core abstraction. Adding one (Podman, nerdctl, DevContainers) is the most welcomed contribution type.

1. **Implement the `Provider` interface** in `internal/engine/<your_backend>_provider.go`:

    ```go
    type Provider interface {
        Name() string
        IsAvailable() error
        Start(cfg *config.ProjectConfig, flags Flags) error
        Stop(cfg *config.ProjectConfig) error
        Shell(cfg *config.ProjectConfig, args []string) error
        Status(cfg *config.ProjectConfig) (EnvironmentStatus, error)
    }
    ```

    When `args` is empty, `Shell` drops into an interactive shell; when non-empty, it runs `args` as a one-shot command in the environment.

2. **Register it in `ResolveProvider`** (`internal/engine/provider.go`) with an explicit case for its name.

3. **Add auto-detection** if applicable — e.g., "this backend is chosen when `provider: auto` and the user has `podman` but no `docker`."

4. **Add tests.** At minimum: a provider-level test with stubs for the external tool, and an entry in `TestResolveProvider` for the dispatch case.

5. **Update docs:**
   - `api_reference.md` → add the value to the `provider` enum and any backend-specific config block.
   - `architecture.md` → if the backend has interesting composition behavior.
   - `why_derrick.md` → if it's worth comparing to an existing alternative.

6. **If the backend can compose with another** (like Hybrid = Docker + Nix), see `internal/engine/hybrid_provider.go` for the `providerLeg` pattern. Composing providers is a feature, not a special case.

## Releasing

Maintainers only. The release process:

1. Land all target PRs on `main`, green CI.
2. Move `[Unreleased]` entries in `CHANGELOG.md` to a new `[X.Y.Z]` section, dated.
3. Bump `Version` in `cmd/derrick/version.go`.
4. Commit: `chore(release): bump version to X.Y.Z`.
5. Tag: `git tag -a vX.Y.Z -m "vX.Y.Z — <one-line summary>"`.
6. Push: `git push origin main && git push origin vX.Y.Z`.
7. The release workflow builds cross-platform binaries and publishes the GitHub release.

Semver rules we follow:

- **Major** — breaking CLI flag, `derrick.yaml` schema, or state-file format changes.
- **Minor** — new features, new config fields, new commands (additive).
- **Patch** — bug fixes, perf, docs.

## Architecture principles

When proposing larger changes, these principles are load-bearing:

- **Supreme Orchestrator.** Derrick wraps proven tools (Docker CLI, Nix) by `exec`. It does not re-implement what they already do. If your change reimplements a Compose feature, reconsider.
- **CLI-layer agnostic.** `cmd/derrick/` must never branch on "is this Docker or Nix." All backend-specific logic lives behind `Provider`.
- **Error translation, not suppression.** Subprocess failures go through `internal/engine/executor.go`'s `translateError` table. Add patterns there; never swallow errors silently.
- **Fail-fast validation at boot.** `derrick doctor` and the first-pass checks in `derrick start` should catch broken config before any provider action runs.

## Where to ask questions

- **Feature ideas / "is this in scope?"** — open a GitHub Discussion or draft RFC-style issue.
- **Bug reports** — open an issue with `derrick doctor --json` output attached.
- **Security issues** — email the maintainer directly, not a public issue.
