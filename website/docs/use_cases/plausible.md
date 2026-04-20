---
title: Plausible Analytics
---

# Plausible local Orchestration

**Plausible Analytics** acts as a difficult case study (Medium/Hard) due to its dependency on specific versions of **Elixir (Erlang/BEAM)**, alongside heavily concurrent database boundaries: **ClickHouse** running parallel with **PostgreSQL**.

### The Derrick Solution

Erlang bindings break severely between OS versions. By isolating Elixir/OTP strictly inside the `nix_packages` array, Derrick forces all team members to compile ClickHouse handlers on the exact same Erlang layer. We then use `after_start` to guarantee migrations hit the databases safely.

### The `derrick.yaml` Implementation

```yaml
---
name: "plausible-analytics"
version: "1.0.0"

dependencies:
  nix_packages:
    - "elixir_1_15" # Version locks OTP under the hood!
    - "nodejs_18"   # Frontend assets
  docker_compose: "docker-compose.yml"

env:
  CLICKHOUSE_DATABASE:
    description: "Analytics CH DB"
    default: "plausible_events"
  SECRET_KEY_BASE:
    description: "Elixir crypto key (Requires 64 bytes)"
    required: true

validations:
  - name: "Is mix installed?"
    command: "mix --version" # Enforces the Nix lock worked cleanly before booting.

hooks:
  setup:
    - "mix local.hex --force"
    - "mix local.rebar --force"
    - "mix deps.get"
  after_start:
    - "mix ecto.create"
    - "mix clickhouse.migrate"
    - "echo 'Visit http://localhost:8000'"
```
