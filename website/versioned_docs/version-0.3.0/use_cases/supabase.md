---
title: Supabase Architecture
---

# Supabase Local Orchestration

**Supabase** represents an extremely challenging local environment (Hard). It requires compiling **Go**, navigating **PostgreSQL (PostgREST)** bindings, establishing **Node.js** pipelines alongside an API gateway (**Kong**), and wiring up real-time websocket layers.

### The Derrick Solution

Derrick abstracts this complexity by focusing heavily on the `env` block and Profile overriding. By validating secrets interactively, a developer starting Supabase locally doesn't have to spend 4 hours debugging why Kong rejected their JWT secret token.

### The `derrick.yaml` Implementation

```yaml
---
name: "supabase-core"
version: "2.1.0"

dependencies:
  nix_registry: "github:NixOS/nixpkgs/nixos-unstable"
  nix_packages:
    - "go_1_21"
    - "nodejs_20"
  docker_compose: "docker/docker-compose.yml"
  docker_compose_profiles:
    - "analytics"

env:
  POSTGRES_PASSWORD:
    description: "Database root pass"
    required: true
    default: "postgres"
  JWT_SECRET:
    description: "Critical Kong Secret Route Key"
    required: true
    validation: "node -e 'if(process.env.JWT_SECRET.length < 32) process.exit(1)'" # Fails fast if secret is weak

validations:
  - name: "Port 5432 Open"
    command: "! lsof -i :5432"

hooks:
  post_start:
    - "echo 'Supabase Core Online: http://localhost:8000'"

profiles:
  auth-only:
    extend: "default"
    dependencies:
      docker_compose_profiles:
        - "auth-module" # Ignores heavy modules like storage or analytics
```
