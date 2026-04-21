---
title: Plausible Analytics
---

# Plausible Analytics

**The problem:** Plausible runs Elixir on the BEAM, ClickHouse for event storage, and Postgres for user data. Erlang bindings break across OS versions, and ClickHouse's HTTP interface clashes with local Postgres ports if both aren't coordinated.

**The Derrick solution:** `provider: hybrid` pins the Elixir/OTP version via Nix and starts ClickHouse + Postgres via Compose. The `after_start` hooks only run DB migrations once services are reachable — no more race conditions against a Postgres that isn't ready.

```yaml
name: plausible-analytics
version: 2.1.0
provider: hybrid

nix:
  packages:
    - elixir_1_16
    - nodejs_20

docker:
  compose: docker-compose.yml
  # starts: postgres:15, clickhouse/clickhouse-server

env:
  SECRET_KEY_BASE:
    description: "Phoenix secret key — generate with: openssl rand -hex 64"
    required: true
    validation: "[ ${#SECRET_KEY_BASE} -ge 64 ] || (echo 'SECRET_KEY_BASE must be ≥ 64 chars' && exit 1)"
  DATABASE_URL:
    description: "PostgreSQL connection string"
    default: "postgresql://postgres:postgres@localhost:5432/plausible_dev"
  CLICKHOUSE_DATABASE_URL:
    description: "ClickHouse HTTP endpoint"
    default: "http://localhost:8123/plausible_events_db"

env_management:
  base_file: .env.example
  prompt_missing: true

validations:
  - name: "Elixir 1.16"
    command: "elixir --version | grep -q 'Elixir 1.16'"

hooks:
  setup:
    - run: "mix local.hex --force && mix local.rebar --force"
      when: first-setup
    - run: "mix deps.get"
      when: first-setup
    - run: "mix compile"
      when: first-setup
  after_start:
    - run: "mix ecto.create && mix ecto.migrate"
      when: first-setup
    - "echo 'Plausible running at http://localhost:8000'"

flags:
  reset-db:
    description: "Drop and recreate both Postgres and ClickHouse schemas"
```

## What's happening

| Stage | Command | Why |
| :--- | :--- | :--- |
| `setup` (first time) | `mix local.hex / rebar` | Bootstraps the Hex package manager inside the nix sandbox. |
| `setup` (first time) | `mix deps.get` | Fetches all Elixir dependencies. |
| `setup` (first time) | `mix compile` | Compiles the application — catches errors before services start. |
| `after_start` (first time) | `mix ecto.create && mix ecto.migrate` | Safe to run only after Postgres is up. |

## Why not hooks.setup for migrations?

Migrations need a live Postgres. Running them in `setup` (before `provider.Start`) would race against the container — they go in `after_start` where services are guaranteed reachable.
