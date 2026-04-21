---
title: Plane
---

# Plane

**The problem:** Plane (the open-source Linear alternative) runs a Django API, a Next.js frontend, Celery workers, Redis, and Postgres. Standing it up locally means coordinating five processes, their environment variables, and their startup order — and the official docs recommend Docker Compose but don't address the day-two developer experience.

**The Derrick solution:** `provider: hybrid` pins Python and Node via Nix. Compose runs the data layer (Postgres, Redis, MinIO). The `setup` hook runs migrations once; subsequent `derrick start` calls skip them and go straight to the workers.

```yaml
name: plane
version: 0.22.0
provider: hybrid

nix:
  packages:
    - python311
    - nodejs_20

docker:
  compose: docker-compose.dev.yml

env:
  SECRET_KEY:
    description: "Django secret key — generate with: openssl rand -hex 32"
    required: true
  DATABASE_URL:
    description: "Postgres connection string"
    default: "postgresql://plane:plane@localhost:5432/plane"
  REDIS_URL:
    default: "redis://localhost:6379/"
  AWS_S3_ENDPOINT_URL:
    description: "MinIO endpoint for file storage"
    default: "http://localhost:9000"
  AWS_ACCESS_KEY_ID:
    default: "plane-access-key"
  AWS_SECRET_ACCESS_KEY:
    default: "plane-secret-key"
  AWS_S3_BUCKET_NAME:
    default: "uploads"
  WEB_URL:
    description: "Public URL for the frontend"
    default: "http://localhost:3000"

env_management:
  base_file: .env.example
  prompt_missing: true

validations:
  - name: "Python 3.11"
    command: "python3 --version | grep -q 'Python 3.11'"
  - name: "Port 3000 free"
    command: "! lsof -i :3000"
  - name: "Port 8000 free"
    command: "! lsof -i :8000"

hooks:
  setup:
    - run: "pip install -r requirements/dev.txt"
      when: first-setup
    - run: "npm install --prefix web"
      when: first-setup
  after_start:
    - run: "python manage.py migrate --noinput"
      when: first-setup
    - "echo 'Plane frontend: http://localhost:3000'"
    - "echo 'Plane API:      http://localhost:8000'"

flags:
  migrate:
    description: "Run Django migrations (use after pulling schema changes)"
  seed:
    description: "Seed the database with sample workspace and issues"
```

## What's happening

| Stage | Command | Why |
| :--- | :--- | :--- |
| `setup` (first time) | `pip install` | Installs Django + Celery deps in the nix Python environment. |
| `setup` (first time) | `npm install` | Installs Next.js deps for the web frontend. |
| `after_start` (first time) | `python manage.py migrate` | Runs DB migrations once Postgres is up. |

## Running migrations after a pull

```bash
derrick start --flag migrate
```

```yaml
hooks:
  setup:
    - run: "python manage.py migrate --noinput"
      when: flag:migrate
```

## Architecture note

Plane's `docker-compose.dev.yml` typically only runs the data layer (Postgres, Redis, MinIO). The API and frontend run as local processes. If you want the full containerized stack instead, switch to `docker-compose.yml` and drop the `nix` block.
