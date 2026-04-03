---
layout: default
title: API & Config Reference
---

# ⚙️ CLI Reference & `derrick.yaml` Spec

This guide exhaustively covers the Derrick CLI surface and validates all possible schemas configurable in the `derrick.yaml` design state.

## 🛠 Command Line Interface (CLI)

The CLI acts as your portal to the sandbox. 

* `derrick start`
  * **Behavior**: Reads the `derrick.yaml`, executes `pre_init` hooks, validates environment constraints, downloads Nix derivations, boosts Docker Compose profiles, and executes `pre_start` hooks.
* `derrick shell`
  * **Behavior**: Bypasses host binaries. Spawns an interactive bash terminal hermetically mapped to the dependencies defined inside the `derrick.yaml`. 
* `derrick stop`
  * **Behavior**: Hooks `post_stop` scripts and gracefully orchestrates Docker down algorithms.
* `derrick doctor`
  * **Behavior**: Diagnostics only. Runs `Validations` mapping network/port bindings. Does not bootstrap anything. Useful for troubleshooting why a teammate's environment refuses to start.
* `derrick dashboard`
  * **Behavior**: Spawns an interactive BubbleTea application pane visually mapping Docker Container Health logs across sidecars.

---

## 📄 The `derrick.yaml` Contract Reference

Your configuration defines the absolute state of the project. A complete YAML uses the `ProjectConfig` schema beneath the internal engine logic.

### 1. Metadata
```yaml
name: "my_service" # Must be lowercase as per validation tags
version: "1.2.0"
requires: 
  - "database_service" # For project clustering hints across workspaces
```

### 2. Dependencies
Orchestrate global dependencies seamlessly.
```yaml
dependencies:
  nix_registry: "github:NixOS/nixpkgs/nixos-unstable" # Highly recommended to keep unstable for modern packages
  nix_packages:
    - "nodejs_20"
    - "go"
    - "postgresql_15"
  docker_compose: "docker-compose.yml" # Target filepath for container services
  docker_compose_profiles: # E.g., spawn only 'backend' profiles from a multi-service yaml
    - "cache"
```

### 3. Environment Constraints (`env`)
Avoid missing `.env` miscommunication scenarios and assure developers possess tokens they need securely.
```yaml
env:
  STRIPE_SECRET_KEY:
    description: "Production Stripe Key for Sandbox Testing"
    required: true
    # Evaluated at runtime! If it exits > 0, the environment fails fast.
    validation: "curl -s --fail -H \"Authorization: Bearer $STRIPE_SECRET_KEY\" https://api.stripe.com/v1/balance"
  DB_PASSWORD:
    description: "Local Postgres bypass"
    default: "postgres"
```

### 4. Direct Validations
Custom assertions ran systematically before Nix or Docker attempt to pull assets.
```yaml
validations:
  - name: "Is Port 8080 Context Availability?"
    command: "! lsof -i :8080" # Fails fast if port is busy
    auto_fix: "kill -9 $(lsof -t -i:8080)" # Automatically executes if the above check fails
```

### 5. Lifecycle Hooks
Bash scripts fired across bootstrap stages. Extremely efficient for executing seed generators or local development commands.
```yaml
hooks:
  pre_init:
    - "echo 'Validating directory structures...'"
  post_init:
    - "npm install" # Happens safely *after* Nix successfully caches the nodejs_20 binary!
  pre_start:
    - "go run scripts/migrate.go" # Runs locally against the network footprint before compose fully yields.
```

### 6. Profile Overrrides
Derrick fundamentally allows environment inheritance to scale up into multiple profiles easily. By targeting a profile via the CLI (e.g., `-p prod`), you can extend or replace core attributes.
```yaml
profiles:
  testing:
    extend: "default"
    dependencies:
      docker_compose_profiles:
        - "cache" 
        - "integration-broker" # In the testing profile, we append the integration-broker container!
```
