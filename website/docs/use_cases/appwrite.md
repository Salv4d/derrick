---
title: Appwrite
---

# Appwrite Local Orchestration

**Appwrite** poses an enormous infrastructure footprint (Hard). Specifically, mapping pre-compiled **PHP Extensions** like Swoole is notoriously difficult for macOS and Windows/WSL users attempting a local boot.

### The Derrick Solution

Nix handles all C/C++ compiler toolchains underneath without user-input. By bounding Appwrite's PHP requirement and Swoole extensions inside the `dependencies` matrix, the user skips painful Makefiles entirely.

### The `derrick.yaml` Implementation

```yaml
---
name: "appwrite-core"
version: "1.4.0"

dependencies:
  nix_packages:
    - "php82"
    - "php82Extensions.swoole" # Critical pain-point resolved!
    - "php82Extensions.redis"
  docker_compose: "docker-compose.yml"
  docker_compose_profiles:
    - "core" # Spins up Redis and MariaDB via standard Docker

env:
  _APP_ENV:
    default: "development"
  _APP_DB_HOST:
    default: "mariadb"
  _APP_DB_USER:
    default: "root"

validations:
  - name: "Is Swoole Loaded?"
    command: "php -r \"if (!extension_loaded('swoole')) exit(1);\"" # Instantly verifies the Nix pipeline!

hooks:
  post_init:
    - "composer install --ignore-platform-reqs"
  post_start:
    - "php app/worker-http.php" # Spawns the Swoole Worker locally mapped to container DBs
```
