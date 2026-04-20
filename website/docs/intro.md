---
slug: /
sidebar_position: 1
title: Introduction
---

# Derrick

Derrick is a local dev-environment orchestrator. You describe what your project needs in a `derrick.yaml` — language toolchain, containerized services, setup hooks — and a single `derrick start` boots the whole thing, the same way, on every teammate's laptop.

## Why it exists

Most dev teams live with one of two problems:

- **Host pollution.** `nvm`, `pyenv`, `asdf`, global `go install`, `brew` drift. Three engineers, three subtly different toolchains, one "works on my machine" PR review.
- **Heavy container setups.** Full dev-containers boot a VM for every project. Language servers lag. File watching is flaky. Your IDE fights the sandbox.

Derrick takes a middle path: reproducible language toolchains via **Nix** (host stays clean, but no VM), paired with **Docker Compose** for the services that genuinely need to run as servers (Postgres, Redis, ClickHouse, Grafana). You pick one or compose both.

## How to think about it

Three concepts carry almost everything:

1. **Providers** are the backends. `nix` for toolchains, `docker` for services, `hybrid` for both. The CLI layer is provider-agnostic — `derrick start` delegates to whichever one your `derrick.yaml` names. → [Architecture: Provider interface](./architecture.md#the-provider-interface)
2. **State** lives in `.derrick/state.json` and is locked per-project. It tracks whether first-setup has run, which flags were active, and which provider owns the environment. → [Architecture: State management](./architecture.md#state-management)
3. **Hooks** are lifecycle scripts split across five stages (`before_start`, `setup`, `after_start`, `before_stop`, `after_stop`), each with a `when:` condition (`always`, `first-setup`, `flag:<name>`). Setup-style commands like `npm install` live in `setup` (sandbox ready, services not yet up); DB seeding lives in `after_start`. The same YAML encodes "run once on clone" and "run every boot" without separate config sections. → [Architecture: Hook executor](./architecture.md#the-hook-executor)

Read those three architecture links in order and you have the full mental model.

## Start from your use case

<div class="personas">

### I'm onboarding to a microservices backend
Your team has a compose file and a handful of setup scripts. Wrap them in a `derrick.yaml` so new hires run one command instead of chasing a wiki.
→ [Getting Started tutorial](./getting_started.md)

### I want a reproducible polyglot shell
Go + Node + Python pinned to exact versions, without asdf/pyenv. Nix alone handles this cleanly. No Docker needed.
→ [Getting Started → Nix-only project](./getting_started.md#3-your-first-project)

### I run many projects at once
Multiple services, each with its own Postgres, all on the same laptop. Derrick isolates them per-project — no shared networks, no port-remapping magic.
→ [Architecture → Multi-project behavior](./architecture.md#multi-project-behavior)

### I'm evaluating vs. mise / devcontainers / devenv.sh
Honest comparison with trade-offs.
→ [Why Derrick?](./why_derrick.md)

### I want to copy a real config
Working `derrick.yaml` files for Supabase, Grafana, Ghost, Plausible, Appwrite.
→ [Recipes](./use_cases/)

</div>

## Next steps

- **[Install Derrick](./installation.md)** — one line on Linux/macOS.
- **[Getting Started tutorial](./getting_started.md)** — zero to working in 3 minutes.
- **[CLI & `derrick.yaml` Reference](./api_reference.md)** — every command and config field.
- **[Troubleshooting](./troubleshooting.md)** — when something breaks.
