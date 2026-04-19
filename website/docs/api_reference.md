---
layout: default
title: API & Config Reference
---

# CLI Reference & `derrick.yaml` Spec

This guide covers the complete Derrick CLI surface and every field in `derrick.yaml`.

---

## Command Line Interface

| Command | Description |
| :--- | :--- |
| `start [alias]` | Resolve the provider, run lifecycle hooks, and boot the environment. An optional alias clones and starts a Hub-registered project. |
| `stop` | Run stop hooks and tear down the environment gracefully. |
| `shell` | Open an interactive shell inside the active environment. |
| `run [packages...] [-- cmd]` | Spawn a throwaway Nix environment with ad-hoc packages. |
| `init` | Interactive wizard that generates `derrick.yaml` for a new project. |
| `doctor` | Audit the environment against `derrick.yaml` without booting it. |
| `dashboard` | Open the BubbleTea TUI dashboard. |
| `clean` | Garbage-collect orphaned Nix derivatives and unused Docker assets. |
| `update` | Replace the local binary with the latest GitHub release. |
| `version` | Print version and check for updates. |

### Global flags

| Flag | Description |
| :--- | :--- |
| `--debug` | Stream raw subprocess output and verbose diagnostics. |
| `-f, --file` | Custom config file path (default: `derrick.yaml`). |
| `-p, --profile` | Named profile to activate (see Profiles below). |

### `start` flags

| Flag | Description |
| :--- | :--- |
| `--reset` | Signal providers and hooks to rebuild from scratch. |
| `--flag <name>` | Activate a custom project flag (can be repeated). Enables `when: flag:<name>` hooks. |

---

## `derrick.yaml` Reference

### Top-level metadata

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `name` | `string` | **Yes** | Project name. Must be lowercase. |
| `version` | `string` | **Yes** | Project version string. |
| `provider` | `string` | No | Isolation backend: `docker`, `nix`, or `auto` (default). `auto` picks Docker when a compose file is present, otherwise Nix. |
| `requires` | `[]string` | No | Sibling project aliases that must be running. Derrick clones and starts them automatically. |

```yaml
name: "my-api"
version: "1.0.0"
provider: docker
```

---

### `docker` block

Used when `provider` is `docker` or `auto`.

| Field | Type | Description |
| :--- | :--- | :--- |
| `docker.compose` | `string` | Path to the Docker Compose file. |
| `docker.profiles` | `[]string` | Compose profiles to activate. |
| `docker.shell` | `string` | Service to exec into for `derrick shell`. Defaults to the first service in the compose file. |

```yaml
docker:
  compose: ./docker-compose.yml
  profiles: [dev, worker]
  shell: app
```

---

### `nix` block

Used when `provider` is `nix` or `auto`.

| Field | Type | Description |
| :--- | :--- | :--- |
| `nix.registry` | `string` | Nixpkgs flake input URL (default: `github:NixOS/nixpkgs/nixos-unstable`). |
| `nix.packages` | `[]string` or `[]object` | Packages to install. A plain string resolves from the default registry; `{package, registry}` allows per-package overrides. |

```yaml
nix:
  registry: "github:NixOS/nixpkgs/nixos-unstable"
  packages:
    - "go"
    - "nodejs_22"
    - package: "legacy_tool"
      registry: "github:NixOS/nixpkgs/nixos-22.11"
```

---

### `hooks` block

Lifecycle hooks are lists of commands run at each stage. Each entry is either a plain string (runs always) or a structured object with an optional `when:` condition.

| Hook stage | Timing |
| :--- | :--- |
| `hooks.start` | After the provider starts the environment. |
| `hooks.stop` | During `derrick stop`, before teardown. |
| `hooks.restart` | During a restart cycle (stop + start). |

#### `when:` conditions

| Value | Fires when... |
| :--- | :--- |
| `always` (default) | Every invocation. |
| `first-setup` | Only on the very first successful `derrick start`. Use for one-time setup: migrations, dependency installs, seed data. |
| `flag:<name>` | Only when `--flag <name>` is passed to `derrick start`. |

```yaml
hooks:
  start:
    - run: "echo 'Booting...'"
      when: always
    - run: "go mod download && make migrate"
      when: first-setup
    - run: "make seed-db"
      when: flag:seed-db
  stop:
    - run: "make cleanup"
      when: always
```

---

### `flags` block

Declare custom project flags. They appear in `derrick start --help` and gate `when: flag:<name>` hooks.

```yaml
flags:
  reset:
    description: "Rebuild the environment and replay migrations from scratch"
  seed-db:
    description: "Inject initial seed data after the environment boots"
```

Usage:
```bash
derrick start --flag seed-db
derrick start --flag reset --flag seed-db
```

---

### `env` block

Declare environment variables the project requires. Derrick validates them at startup and interactively prompts for missing ones.

| Field | Type | Description |
| :--- | :--- | :--- |
| `env.<KEY>.description` | `string` | Shown when prompting the developer. |
| `env.<KEY>.required` | `bool` | Fail-fast if the variable is missing and has no default. |
| `env.<KEY>.default` | `string` | Value injected when the variable is absent. |
| `env.<KEY>.validation` | `string` | Shell command. Non-zero exit triggers an interactive resolution flow. |

```yaml
env:
  DATABASE_URL:
    description: "PostgreSQL connection string"
    required: true
    default: "postgres://localhost:5432/myapp_dev"
  STRIPE_SECRET_KEY:
    required: true
    validation: "curl -sf -H 'Authorization: Bearer $STRIPE_SECRET_KEY' https://api.stripe.com/v1/balance"
```

---

### `env_management` block

| Field | Type | Description |
| :--- | :--- | :--- |
| `env_management.base_file` | `string` | Template file (e.g. `.env.example`) auto-copied to `.env` when absent. |
| `env_management.prompt_missing` | `bool` | Interactively prompt for any empty variables in the env file. |

```yaml
env_management:
  base_file: ".env.example"
  prompt_missing: true
```

---

### `validations` block

Arbitrary shell assertions run before the environment boots.

| Field | Type | Description |
| :--- | :--- | :--- |
| `validations[].name` | `string` | Human-readable check name. |
| `validations[].command` | `string` | Shell command. Non-zero exit = failure. |
| `validations[].auto_fix` | `string` | Shell command run automatically on failure, then re-checked. |

```yaml
validations:
  - name: "Port 8080 is free"
    command: "! lsof -i :8080"
    auto_fix: "kill -9 $(lsof -t -i:8080)"
  - name: "Go compiler"
    command: "go version"
```

---

### `profiles` block

Named overlays that extend or override the base config. Profiles can extend other profiles via `extend`.

```yaml
profiles:
  ci:
    nix:
      packages: ["golangci-lint"]
    hooks:
      start:
        - run: "go test ./..."
          when: always
  staging:
    extend: ci
    docker:
      profiles: [staging]
```

Activate with:
```bash
derrick start --profile staging
```

---

## Complete example

```yaml
name: "payment-service"
version: "2.1.0"
provider: docker

docker:
  compose: ./docker-compose.yml
  profiles: [dev]

nix:
  packages:
    - "go"
    - "golangci-lint"

hooks:
  start:
    - run: "echo 'Starting payment-service...'"
      when: always
    - run: "go mod download && make migrate"
      when: first-setup
    - run: "make seed-db"
      when: flag:seed-db
  stop:
    - run: "echo 'Goodbye!'"
      when: always

flags:
  seed-db:
    description: "Populate the database with seed data"
  reset:
    description: "Drop and recreate the database schema"

requires:
  - auth-service

env:
  DATABASE_URL:
    required: true
    default: "postgres://localhost:5432/payments_dev"
  JWT_SECRET:
    required: true

validations:
  - name: "Go compiler"
    command: "go version"
  - name: "Docker daemon"
    command: "docker info"
```
