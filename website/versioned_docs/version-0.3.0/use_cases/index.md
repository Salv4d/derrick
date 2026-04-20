---
sidebar_position: 5
title: Recipes
---

# Recipes

Copy-paste `derrick.yaml` files for real open-source projects. Each recipe is a complete, working configuration you can adapt to your own codebase. Pick the one closest to your stack.

## By topology

### Polyglot backend (10+ services)
**[Supabase](./supabase.md)** — Postgres, Auth, Realtime, Storage, Edge Functions, Studio. Aggressive `.env` validation trees and tight boot ordering.

### Strict boot-order pipeline (Elixir OTP)
**[Plausible Analytics](./plausible.md)** — Elixir + Phoenix + ClickHouse + Postgres with deterministic startup sequencing.

### Frontend + backend polyglot
**[Grafana](./grafana.md)** — Go backend + React frontend + Yarn toolchain cleanly isolated in a single sandbox.

### Language-pinned single-runtime
**[Ghost CMS](./ghost.md)** — Pinning Node.js LTS without `nvm` or `asdf` drift.

### Complex compiled extensions
**[Appwrite](./appwrite.md)** — C++ PHP extensions (Swoole) pre-built inside the Nix sandbox so contributors never compile them by hand.

## By feature demonstrated

| Want to see how to… | Recipe |
| :--- | :--- |
| Validate required `.env` keys before boot | [Supabase](./supabase.md) |
| Orchestrate dependent services with boot ordering | [Plausible](./plausible.md) |
| Mix native toolchain with containerized services (`provider: hybrid`) | [Grafana](./grafana.md) |
| Pin an EOL Node version | [Ghost](./ghost.md) |
| Ship pre-built binaries for C/C++ dependencies | [Appwrite](./appwrite.md) |

## Using a recipe

1. Copy the `derrick.yaml` from the recipe page into your project root.
2. Adjust the `name`, `version`, and any paths to match your repo layout.
3. Run `derrick start`.
4. If first-setup needs project-specific seeding, add a `when: first-setup` hook.

All recipes are tested against real upstream repos in `benchmark_projects.md`. If a recipe breaks against a new upstream release, please [open an issue](https://github.com/Salv4d/derrick/issues) or PR the fix.
