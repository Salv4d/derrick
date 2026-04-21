---
title: Sentry
---

# Sentry

**The problem:** Sentry self-hosted requires Python (Django backend), Node (frontend build), and a stack of Postgres, Redis, Kafka, ClickHouse, and Zookeeper containers. The official `install.sh` script takes 20+ minutes and writes to the host — every contributor runs a different variant.

**The Derrick solution:** `provider: hybrid` pins Python and Node via Nix. Compose runs the entire service stack from Sentry's official `docker-compose.yml`. The `setup` hook runs `sentry upgrade` once — which handles DB migrations and creates the admin user — so the team never races against an uninitialized schema.

```yaml
name: sentry-self-hosted
version: 24.4.0
provider: hybrid

nix:
  packages:
    - python311
    - nodejs_20

docker:
  compose: docker-compose.yml

env:
  SENTRY_SECRET_KEY:
    description: "Django secret key — generate with: openssl rand -base64 50"
    required: true
  SENTRY_DB_PASSWORD:
    description: "Postgres password for Sentry schema"
    default: "sentry"
  SENTRY_REDIS_HOST:
    default: "redis"
  SENTRY_EMAIL_HOST:
    description: "SMTP relay for email alerts (leave blank to disable)"
    default: ""

env_management:
  base_file: .env.example
  prompt_missing: true

validations:
  - name: "Python 3.11"
    command: "python3 --version | grep -q 'Python 3.11'"
  - name: "Port 9000 free"
    command: "! lsof -i :9000"

hooks:
  setup:
    - run: "pip install -r requirements-base.txt"
      when: first-setup
  after_start:
    - run: "sentry upgrade --noinput"
      when: first-setup
    - "echo 'Sentry running at http://localhost:9000'"

flags:
  reset-db:
    description: "Drop and recreate the Sentry database (loses all data)"
```

## What's happening

| Stage | Command | Why |
| :--- | :--- | :--- |
| `setup` (first time) | `pip install` | Installs Sentry Python deps in the nix Python environment. |
| `after_start` (first time) | `sentry upgrade` | Runs Django migrations and seeds the DB — safe only after Postgres is up. |

## Generating SENTRY_SECRET_KEY

```bash
openssl rand -base64 50
```

Add the output to `.env` (or `.env.local` if you keep `.env` in version control).

## Resetting the database

```bash
derrick start --flag reset-db
```

Wire the flag in your `derrick.yaml`:

```yaml
hooks:
  setup:
    - run: "sentry django flush --noinput && sentry upgrade --noinput"
      when: flag:reset-db
```
