---
title: n8n
---

# n8n

**The problem:** n8n is a Node.js automation platform that stores workflow state in either SQLite (fine for one developer) or Postgres (required for teams). The official Docker image works, but contributors who need to modify n8n nodes have to build from source — which requires the right Node and pnpm versions.

**The Derrick solution:** Two patterns — a pure `docker` setup for teams just running n8n, and a `hybrid` setup for contributors building custom nodes with a pinned Node + pnpm version.

## Running n8n (no customisation)

For teams that run n8n as-is with Postgres persistence:

```yaml
name: n8n
version: 1.40.0
provider: docker

docker:
  compose: docker-compose.yml

env:
  N8N_BASIC_AUTH_USER:
    description: "Basic auth username for n8n UI"
    default: "admin"
  N8N_BASIC_AUTH_PASSWORD:
    description: "Basic auth password"
    required: true
    default: "admin"
  N8N_ENCRYPTION_KEY:
    description: "Key for encrypting stored credentials — generate with: openssl rand -hex 24"
    required: true
  DB_POSTGRESDB_PASSWORD:
    required: true
    default: "n8n"
  WEBHOOK_URL:
    description: "Public URL for incoming webhooks (use ngrok or similar for local testing)"
    default: "http://localhost:5678"

validations:
  - name: "Port 5678 free"
    command: "! lsof -i :5678"

hooks:
  after_start:
    - "echo 'n8n running at http://localhost:5678 ($N8N_BASIC_AUTH_USER / $N8N_BASIC_AUTH_PASSWORD)'"
```

## Building custom nodes

For contributors developing custom n8n nodes who need the full source toolchain:

```yaml
name: n8n-dev
version: 1.40.0
provider: hybrid

nix:
  packages:
    - nodejs_20

docker:
  compose: docker-compose.dev.yml

env:
  N8N_ENCRYPTION_KEY:
    description: "Credential encryption key"
    required: true
  DB_TYPE:
    default: "sqlite"

hooks:
  setup:
    - run: "corepack enable && pnpm install --frozen-lockfile"
      when: first-setup
    - run: "pnpm build"
      when: first-setup
  after_start:
    - "echo 'n8n dev server: http://localhost:5678'"

flags:
  rebuild:
    description: "Rebuild all packages from source"
```

## What's happening

| Stage | Command | Why |
| :--- | :--- | :--- |
| `setup` (first time) | `corepack enable` | Activates pnpm via the Node.js corepack shim — no global pnpm install needed. |
| `setup` (first time) | `pnpm install` | Installs the monorepo workspace dependencies. |
| `setup` (first time) | `pnpm build` | Compiles TypeScript across all packages. |

## Incremental rebuilds

```bash
derrick start --flag rebuild
```

```yaml
hooks:
  setup:
    - run: "pnpm build"
      when: flag:rebuild
```
