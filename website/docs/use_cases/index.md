---
sidebar_position: 5
title: Recipes
---

# Recipes

Copy-paste `derrick.yaml` files for real open-source projects. Each recipe is a complete, working configuration you can adapt to your own codebase. Pick the one closest to your stack.

## By topology

### Polyglot backend (multi-service)
**[Supabase](./supabase.md)** — Postgres, Auth, Realtime, Storage, Kong. Env validation catches weak JWT secrets before containers start.

**[Sentry](./sentry.md)** — Python + Django + Kafka + ClickHouse. `after_start` migrations run only after Postgres is healthy.

### Strict boot-order pipeline
**[Plausible Analytics](./plausible.md)** — Elixir + Phoenix + ClickHouse + Postgres with deterministic startup sequencing.

**[Plane](./plane.md)** — Django API + Next.js frontend + Celery + Redis. Migrations gated behind service readiness.

### Frontend + backend polyglot
**[Grafana](./grafana.md)** — Go backend + React frontend + Yarn toolchain cleanly isolated in a single sandbox.

**[n8n](./n8n.md)** — Node.js automation platform: simple Docker mode for runners, hybrid mode for custom node contributors.

### Language-pinned single-runtime
**[Ghost CMS](./ghost.md)** — Pinning Node.js LTS without `nvm` or `asdf` drift.

**[Keycloak](./keycloak.md)** — Docker-only: realm config imported from JSON so every developer starts from the same auth state.

**[Meilisearch](./meilisearch.md)** — Docker-only: seed data loaded on first start so every developer searches the same dataset.

### Complex compiled extensions
**[Appwrite](./appwrite.md)** — PHP + Compose stack. Nix provides the toolchain; containers run the workers.

## By feature demonstrated

| Want to see how to… | Recipe |
| :--- | :--- |
| Validate required `.env` keys before boot | [Supabase](./supabase.md), [Sentry](./sentry.md) |
| Orchestrate dependent services with boot ordering | [Plausible](./plausible.md), [Plane](./plane.md) |
| Mix native toolchain with containerized services (`provider: hybrid`) | [Grafana](./grafana.md), [n8n](./n8n.md) |
| Pin a specific language version | [Ghost](./ghost.md), [Grafana](./grafana.md) |
| Import config/data on first setup | [Keycloak](./keycloak.md), [Meilisearch](./meilisearch.md) |
| Use `flags` for on-demand operations (rebuild, reseed, migrate) | [Grafana](./grafana.md), [Meilisearch](./meilisearch.md), [Plane](./plane.md) |
| Use `profiles` for lightweight vs full-stack modes | [Supabase](./supabase.md) |
| Use `requires` to boot a dependency automatically | [Meilisearch](./meilisearch.md) |

## Using a recipe

1. Copy the `derrick.yaml` from the recipe page into your project root.
2. Adjust the `name`, `version`, and any paths to match your repo layout.
3. Run `derrick start`.
4. If first-setup needs project-specific seeding, add a `when: first-setup` hook.

All recipes use the current `derrick.yaml` schema. If a recipe breaks against a new upstream release, please [open an issue](https://github.com/Salv4d/derrick/issues) or PR the fix.
