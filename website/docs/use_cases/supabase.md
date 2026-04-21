---
title: Supabase
---

# Supabase Local Orchestration

**The problem:** Supabase local dev requires Go (CLI tooling), Node.js (Studio), and a Docker stack of ~8 services (Postgres, PostgREST, Kong, GoTrue, Realtime, Storage, Meta, Studio). Getting secrets like `JWT_SECRET` wrong silently breaks Kong routing — errors surface minutes later, not at startup.

**The Derrick solution:** `provider: hybrid` pins Go and Node via Nix while Compose runs the full service stack. Env `validation` catches weak secrets before any container starts. A `profiles` overlay lets lightweight contributors skip analytics and storage.

```yaml
name: supabase-core
version: 2.1.0
provider: hybrid

nix:
  packages:
    - go_1_21
    - nodejs_20

docker:
  compose: docker/docker-compose.yml
  profiles:
    - analytics

env:
  POSTGRES_PASSWORD:
    description: "Database root password"
    required: true
    default: "postgres"
  JWT_SECRET:
    description: "Kong JWT signing secret — must be ≥ 32 characters"
    required: true
    validation: "[ ${#JWT_SECRET} -ge 32 ] || (echo 'JWT_SECRET must be ≥ 32 chars' && exit 1)"
  ANON_KEY:
    description: "Supabase anon public key (JWT signed with JWT_SECRET)"
    required: true
  SERVICE_ROLE_KEY:
    description: "Supabase service role key (JWT signed with JWT_SECRET)"
    required: true
  SITE_URL:
    description: "Public URL for redirects"
    default: "http://localhost:3000"

env_management:
  base_file: .env.example
  prompt_missing: true

validations:
  - name: "Port 5432 free"
    command: "! lsof -i :5432"
  - name: "Port 8000 free"
    command: "! lsof -i :8000"

hooks:
  setup:
    - run: "npm install --prefix supabase/studio"
      when: first-setup
  after_start:
    - "echo 'Supabase Studio: http://localhost:3000'"
    - "echo 'API gateway:     http://localhost:8000'"

profiles:
  auth-only:
    docker:
      compose: docker/docker-compose.yml
      profiles:
        - auth-module
```

## What's happening

| Stage | Command | Why |
| :--- | :--- | :--- |
| `setup` (first time) | `npm install --prefix supabase/studio` | Installs Studio frontend deps inside the nix sandbox. |
| `after_start` | echo messages | Reports endpoints once Kong and Studio are reachable. |

## Profiles

`auth-only` swaps the Compose profiles list to start only the auth stack (GoTrue, Kong) — skipping Storage, Realtime, and analytics containers. Useful for contributors working on auth flows who don't need the full footprint:

```bash
derrick start --profile auth-only
```

## JWT secrets

Supabase requires three coordinated secrets. Generate them once and commit to `.env.local`:

```bash
# Generate JWT_SECRET
openssl rand -hex 32

# Generate ANON_KEY and SERVICE_ROLE_KEY using the Supabase CLI
npx supabase gen keys --project-ref local
```
