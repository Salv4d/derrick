# Derrick — Gold Standard Roadmap

This is the living state log for elevating Derrick to a market-ready product.
Tasks are grouped by category and tackled sequentially. Each completed item
gets an atomic commit and a one-line technical note.

**Legend:** `[ ]` pending · `[~] IN PROGRESS` · `[x] DONE`

---

## Core Architecture

- [x] **CA1 — Restore a green build.** Unused import and typo introduced by pending hybrid work broke `go build`. Restore compilation and green tests before any new work.
    - Removed `path/filepath` from `internal/engine/shell.go` (now unused after the shell-init refactor) and corrected `"shell existed with status"` → `"shell exited with status"`. `go build ./...` and `go test ./...` are green.
- [x] **CA2 — Finalize HybridProvider.** Land the pending docker+nix composed provider with a clear `Status()` that reports both legs independently, wire it through `ResolveProvider`, and document it.
    - Introduced `providerLeg` as the narrow interface the hybrid composes, so tests can inject stubs without a running daemon. `IsAvailable` joins both-leg errors with `errors.Join`. `Status` no longer aborts when one leg errors — both are reported side by side. `NewHybridProvider()` is the public constructor; `ResolveProvider` now calls it instead of struct-literal-reaching into unexported fields.
- [x] **CA3 — Route `derrick shell` through `Provider.Shell`.** Today `cmd/derrick/shell.go` hard-codes `engine.NewShellEngine().EnterSandbox(".derrick")`, which breaks docker-only and hybrid projects. Delegate to `provider.Shell(cfg)` so each backend owns the interactive-shell behavior.
    - Widened `Provider.Shell(cfg, args)` so each backend handles one-shot command execution natively — nix via `nix develop --command`, docker via `compose exec <service> <args…>`. `cmd/derrick/shell.go` no longer reaches into the nix sandbox directly; it parses config, resolves the provider, and delegates. Docker-only projects finally get a working `derrick shell`.
- [x] **CA4 — Remove dead code in `internal/engine/docker.go`.** `StartContainers` / `StopContainers` are orphaned since the Provider refactor. Delete to prevent drift and duplicated logic.
    - Removed both orphans and the now-unused `bytes`, `strings`, and `ui` imports. All callers have moved to `DockerProvider.Start/Stop`; no external consumers.
- [x] **CA5 — Harden nil-safety for `state.Load` callers.** `cmd/derrick/stop.go` and `status.go` call `projectState.FlagsUsed` / `.Status` after a `state.Load(_, err) := ...` that returns `nil` on lock failure. Fall back to a zero-value state when Load errs.
    - Changed the contract: `state.Load` now always returns a non-nil `*EnvironmentState` (Status=StatusUnknown) even on error, so the `projectState, _ := state.Load(cwd)` idiom used by `stop`/`status` can no longer nil-deref. Dropped the now-dead `if projectState != nil` check in `status.go`. Added two tests that pin the contract: one for the corrupted-JSON error path, one that guards the non-nil guarantee directly.
- [ ] **CA6 — `NixProvider.Status` should reflect project reality.** Today it returns `Running:true` whenever the `nix` binary exists. Report `true` only when the project's flake has actually been built (i.e. `.derrick/flake.nix` exists).

## CLI UX

- [ ] **UX1 — `derrick completion` command.** Add cobra-native bash/zsh/fish/powershell completions. Low effort, major DX win.
- [ ] **UX2 — Version command stability when offline.** `RunVersion` prints only a single warning when GitHub is unreachable; keep that, but ensure the exit code stays 0 so CI using `derrick version` never fails due to a flaky network.
- [ ] **UX3 — Exit codes for `derrick doctor`.** Today doctor always exits 0. Exit with a non-zero code when `report.Issues > 0` so CI pipelines can gate on it.

## Testing

- [ ] **T1 — Tests for `ResolveProvider` dispatcher.** Cover docker, nix, hybrid, auto-detect, and unknown fallback paths. No external binaries; table-driven.
- [ ] **T2 — Tests for `HybridProvider` composition.** Inject stub providers behind the concrete struct (via an internal interface) and assert Start/Stop/Shell/Status route correctly and propagate errors.
- [ ] **T3 — Tests for hook `shouldRun` conditions.** always / first-setup / flag:* / unknown. Pure function, high value.
- [ ] **T4 — Tests for `GenerateNetworkOverride`.** Verify every service receives the `com.derrick.managed` label and `host.docker.internal` hosts entry.
- [ ] **T5 — Tests for state nil-safety in command callers.** Exercise the path where `state.Load` cannot acquire a lock or the file is malformed.

## CI/CD

- [ ] **CI1 — golangci-lint in PR CI.** Fail fast on vet, ineffassign, staticcheck, unused, gofmt.
- [ ] **CI2 — `go vet` as its own step.** Cheap, already covered by lint but explicit is clearer in logs.

## Documentation

- [ ] **D1 — CHANGELOG.md (Keep a Changelog format).** Seed with v0.1.0 → v0.2.0 so users can see semver history.
- [ ] **D2 — Document hybrid provider and multi-project behavior.** Cover when to use `provider: hybrid`, how `derrick shell` behaves, and what happens when multiple projects run concurrently (state lock, per-project docker network, port conflicts, shared /nix/store).
