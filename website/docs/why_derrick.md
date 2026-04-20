---
sidebar_position: 4
title: Why Derrick?
---

# Why Derrick?

Derrick isn't the only way to manage a local dev environment. This page is an honest comparison against the alternatives you're probably also considering. If another tool fits your project better, you should use it.

## The short answer

Pick Derrick when you need **both** reproducible language toolchains *and* containerized services, and you want them described in one file that a new hire runs with one command. If you only need one half of that, a simpler tool might suit you better.

## Comparison matrix

| Tool | Language toolchain | Services / DB | Host pollution | Setup for a new hire | Single declarative file |
| :--- | :---: | :---: | :---: | :---: | :---: |
| **Derrick** | Nix | Docker Compose | None (Nix sealed) | `derrick start` | `derrick.yaml` |
| [mise](https://mise.jdx.dev/) | asdf-style version shims | Out of scope | Shims on PATH | `mise install && mise x` | `.mise.toml` |
| [Dev Containers](https://containers.dev/) | Inside a container | Inside containers | None (all in VM) | Reopen in container | `devcontainer.json` |
| [devenv.sh](https://devenv.sh/) | Nix | Nix-managed processes | None (Nix sealed) | `devenv up` | `devenv.nix` |
| [direnv](https://direnv.net/) + compose | Your choice | Docker Compose | Depends on `.envrc` | Multiple commands | Two+ files |
| Raw Docker Compose | Inside containers | Compose services | None | `docker compose up` | `docker-compose.yml` |

The columns that matter most in practice:

- **Language toolchain** — where `go`, `node`, `python` actually come from.
- **Services / DB** — where Postgres / Redis / Kafka run.
- **Host pollution** — whether your `~/.zshrc` and global binaries drift.
- **Setup for a new hire** — how many commands (and docs pages) they need to go through.
- **Single declarative file** — whether one file describes the whole environment or you need to chain two or three tools.

## Tool-by-tool honest take

### vs. mise

[mise](https://mise.jdx.dev/) (ex-rtx) is excellent if your whole problem is "pin my language versions and activate them per-directory." It's shim-based like `asdf`, fast, and has great version coverage.

**Choose mise if:** your services run in production-like environments you don't need locally, or you're fine running Postgres via Homebrew. No containerized services needed.

**Choose Derrick if:** you already have a `docker-compose.yml` and also need language version pinning. mise has no story for orchestrating Compose, hooks, per-project state, or cross-service lifecycle.

### vs. Dev Containers (VS Code / `devcontainer.json`)

Dev Containers put *everything* — language tools, services, your editor's language servers — inside a container. Excellent reproducibility, painful developer ergonomics.

**Choose Dev Containers if:** your team is entirely in VS Code, you're comfortable with remote container workflows, and absolute isolation beats local speed.

**Choose Derrick if:** you want language tools running **natively** on your host (fast file watching, no remote-LSP weirdness, your editor doesn't care which editor), while services run in containers. Hybrid mode is the sweet spot for most backend teams.

### vs. devenv.sh

[devenv.sh](https://devenv.sh/) is the closest conceptual match — Nix for toolchains, optional processes, declarative config. Powerful, with a strong ecosystem (built on [devshell](https://numtide.dev/) + Cachix).

**Choose devenv.sh if:** your team is already Nix-fluent and you want to write `devenv.nix` in the Nix language directly. More flexibility, steeper learning curve.

**Choose Derrick if:** you want YAML instead of Nix expressions, and you want first-class Docker Compose integration (devenv treats Compose as a secondary citizen). Derrick's `derrick.yaml` is approachable by engineers who have never heard of Nix.

### vs. direnv + Docker Compose (DIY)

Many teams stitch `direnv` (for PATH), `.env` files (for secrets), and `docker compose up` (for services) into something that works. It does — until the next hire joins and tries to figure out the order of operations from a README.

**Choose DIY if:** your team is small, the setup is static, and a 15-line README is enough.

**Choose Derrick if:** the onboarding doc keeps growing, hooks keep getting added, and someone keeps forgetting to `source .envrc` before `docker compose up`. `derrick start` replaces all of that with one command.

### vs. raw Docker Compose

Docker Compose alone handles services fine, but it puts your language tools inside containers too. Language servers over `docker exec` are slow and fragile; file watchers miss changes on bind-mounted volumes on macOS.

**Choose raw Compose if:** you're already happy with container-side development and don't mind the IDE friction.

**Choose Derrick if:** you want to `go test ./...` as fast as a native build while Postgres runs in a container.

## When **not** to use Derrick

Be honest about fit:

- **You don't want to install Nix.** Derrick can run docker-only projects, but the killer feature is hermetic toolchains via Nix. If that's a no-go, mise or Dev Containers are better.
- **Your team is single-language, single-platform, single-service.** `docker compose up` or `mise install` is simpler. Don't import complexity you won't use.
- **You need Windows-native support today.** Derrick targets Linux, macOS, and WSL2. Native Windows is not on the roadmap; Dev Containers handle Windows better.
- **You need remote / cloud dev environments.** Derrick is local-first. Look at GitHub Codespaces, Gitpod, or Coder.

## Where Derrick is genuinely different

Three design choices make Derrick distinct from everything above:

- **`provider: hybrid` is a first-class mode**, not an afterthought. The Provider interface composes Docker and Nix as two equal legs with explicit responsibility split — Compose owns services, Nix owns the shell.
- **Multi-project isolation is the default.** The old global `derrick-net` was removed in v0.1.0. Each project has its own compose network, its own state lock, and its own label-scoped `derrick clean`. Run five derrick projects concurrently with no crosstalk.
- **Hooks have conditions.** `when: always`, `when: first-setup`, `when: flag:<name>` lets one YAML encode setup, every-boot tasks, and on-demand operations without three separate config sections.
