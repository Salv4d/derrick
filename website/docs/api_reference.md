---
layout: default
title: API & Config Reference
---

# ⚙️ CLI Reference & `derrick.yaml` Spec

This guide exhaustively covers the Derrick CLI surface and validates all possible schemas configurable in the `derrick.yaml` design state.

## 🛠 Command Line Interface (CLI)

The CLI acts as your portal to the sandbox. 

### Commands

| Command | Description | Behavior Context |
| :--- | :--- | :--- |
| `start` | Main Bootstrapper | Reads `derrick.yaml`, validates envs, boots Nix dependencies, and runs Docker Compose. |
| `run` | Ephemeral Sandbox | Spawns a temporary ad-hoc Nix environment based on requested packages. |
| `clean` | System Maintenance | Garbage collects orphaned Nix derivatives and unused Docker assets. |
| `shell` | Hermetic Terminal | Spawns an interactive bash terminal hermetically mapped to Nix definitions. |
| `stop` | Graceful Teardown | Executes `post_stop` scripts and gracefully halts Docker. |
| `update` | Binary Updater | Instantly replaces the local binary with the latest release from GitHub. |
| `doctor`| Diagnostic Tool | Runs Validation checks manually without bootstrapping. |
| `dashboard` | Interactive UI | Spawns a BubbleTea pane mapping Container Health via TUI. |

---

## 📄 The `derrick.yaml` Contract Reference

Your configuration defines the absolute state of the project. A complete YAML uses the `ProjectConfig` schema beneath the internal engine logic.

### 1. Metadata Schema

Identify your microservice cleanly across the host boundaries.

| Keyword | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `name` | `string` | **Yes** | `""` | The canonical name of the project. Must be lowercase. |
| `version` | `string` | **Yes** | `""` | Version mapping for the workspace. |
| `requires` | `array[string]` | No | `[]` | Used for project clustering to declare other services that must be booted. |

**Example Usage**:
```yaml
name: "my_service"
version: "1.2.0"
requires: 
  - "database_service" 
```

### 2. Dependencies Schema

The core pillar. Determines what global packages and container topologies must exist.

| Keyword | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `dependencies.nix_registry` | `string` | No | `"github:NixOS/nixpkgs/nixos-unstable"` | The default Nixpkgs flake or channel to lock. |
| `dependencies.nix_packages` | `array[string|object]` | **Yes** | `[]` | List of system dependencies. Can be an object `{package, registry}` for mixed versions. |
| `dependencies.docker_compose` | `string` | No | `""` | Filepath to docker-compose matrix. |
| `dependencies.docker_compose_profiles` | `array[string]`| No | `[]` | Subset of docker profiles to boot natively. |

**Example Usage**:
```yaml
dependencies:
  nix_registry: "github:NixOS/nixpkgs/nixos-unstable" 
  nix_packages:
    - "nodejs_20"
    - package: "legacy_database"
      registry: "github:NixOS/nixpkgs/nixos-20.09"
  docker_compose: "docker-compose.yml" 
```

### 3. Environment Constraints (`env`)

Avoid missing `.env` miscommunication scenarios and assure developers possess tokens securely.

| Keyword | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `env.<KEY>.description` | `string` | No | `""` | Helper text when prompting developers. |
| `env.<KEY>.required` | `boolean`| No | `false` | Prompts fail-fast mechanism if missing. |
| `env.<KEY>.default` | `string` | No | `""` | Injects baseline values if `required` is false. |
| `env.<KEY>.validation` | `string` | No | `""` | Shell command. If it yields exit code `> 0`, fails validation. |

**Example Usage**:
```yaml
env:
  STRIPE_SECRET_KEY:
    description: "Production Stripe Key for Sandbox Testing"
    required: true
    validation: "curl -s --fail -H \"Authorization: Bearer $STRIPE_SECRET_KEY\" https://api.stripe.com/v1/balance"
```

### 4. Direct Validations

Custom assertions ran systematically before Nix or Docker attempt to pull assets.

| Keyword | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `validations.name` | `string` | **Yes** | `""` | Contextual name of the Validation step. |
| `validations.command` | `string` | **Yes** | `""` | Evaluated Bash script. |
| `validations.auto_fix`| `string` | No | `""` | Shell script triggered natively if `command` fails. |

**Example Usage**:
```yaml
validations:
  - name: "Is Port 8080 Available?"
    command: "! lsof -i :8080" 
    auto_fix: "kill -9 $(lsof -t -i:8080)" 
```

### 5. Lifecycle Hooks

Bash scripts fired across bootstrap stages. Extremely efficient for executing seed generators or local development commands.

| Keyword | Description | Timing Logic |
| :--- | :--- | :--- |
| `pre_init` | Triggers before Nix resolves dependencies. | Used for validation or directory structures. |
| `post_init` | Triggers directly after Sandbox lock established. | E.g., `npm install` (utilizing the cached Nix node bin). |
| `pre_start` | Triggers before Docker Compose UP. | Database setups or network proxies. |
| `post_start`| Triggers when Environment is functionally ready. | Displaying ASCII art, or running auto-migrations. |
| `post_stop` | Triggers after terminal exits and environment tears down.| Cleaning up temporary PID files. |

**Example Usage**:
```yaml
hooks:
  post_init:
    - "npm install" 
  pre_start:
    - "go run scripts/migrate.go"
```

### 6. Profile Overrides (`profiles`)

Derrick allows environment inheritance to scale up into multiple profiles. By targeting a profile via the CLI (e.g., `-p prod`), you can extend core attributes.

| Keyword | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `profiles.<KEY>.extend` | `string` | No | `""` | Inherit from `default` or another profile snippet. |
| `profiles.<KEY>.<SCHEMA>` | `object`| No | `{}` | Re-implements any core `ProjectConfig` key to overwrite. |

**Example Usage**:
```yaml
profiles:
  testing:
    extend: "default"
    dependencies:
      docker_compose_profiles:
        - "integration-broker" 
```
