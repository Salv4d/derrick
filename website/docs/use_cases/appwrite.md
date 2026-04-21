---
title: Appwrite
---

# Appwrite

**The problem:** Appwrite's PHP worker (Swoole) requires native extensions that break on macOS and WSL when installed outside a controlled environment. `pecl install swoole` fails on half of developer machines due to compiler version mismatches.

**The Derrick solution:** `provider: hybrid` puts PHP 8.2 and the Swoole extension inside the Nix sandbox — no Makefiles, no `pecl`, no host compiler involved. Docker Compose runs Redis and MariaDB. The validation check confirms Swoole loaded before the worker starts.

```yaml
name: appwrite-core
version: 1.5.0
provider: hybrid

nix:
  packages:
    - php82
    - php82Packages.composer

docker:
  compose: docker-compose.yml
  profiles:
    - core

env:
  _APP_ENV:
    default: "development"
  _APP_DB_HOST:
    description: "MariaDB host (matches compose service name)"
    default: "mariadb"
  _APP_DB_USER:
    default: "root"
  _APP_DB_PASS:
    required: true
    default: "password"
  _APP_REDIS_HOST:
    default: "redis"
  _APP_OPENSSL_KEY_V1:
    description: "Encryption key — generate with: openssl rand -hex 32"
    required: true

validations:
  - name: "PHP 8.2"
    command: "php --version | grep -q 'PHP 8.2'"

hooks:
  setup:
    - run: "composer install --ignore-platform-reqs"
      when: first-setup
  after_start:
    - "echo 'Appwrite API: http://localhost/v1/health'"
```

## What's happening

| Stage | Command | Why |
| :--- | :--- | :--- |
| `setup` (first time) | `composer install` | Installs PHP dependencies inside the nix sandbox. `--ignore-platform-reqs` skips host extension checks — Nix already satisfies them. |
| `after_start` | echo | Confirms the API is reachable once containers are up. |

## Why no Swoole in Nix packages?

Nixpkgs ships `php82Packages.composer` and the PHP interpreter with common extensions compiled in. Appwrite's HTTP worker runs as a Docker service (`appwrite` container), not a local PHP process — so the host nix sandbox only needs PHP for CLI tooling (Composer, artisan-style commands). The container handles Swoole internally.

If you need Swoole for local CLI testing, add it via a custom Nix overlay or use the container exec approach:

```bash
docker compose exec appwrite php your-script.php
```
