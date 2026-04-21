---
title: Ghost CMS
---

# Ghost CMS

**The problem:** Ghost requires a specific Node.js LTS. Without a toolchain manager every contributor ends up on a different version — Ghost's CLI rejects mismatches with cryptic errors.

**The Derrick solution:** `nix.packages: [nodejs_18]` pins the exact LTS across every machine. No `nvm`, no `.nvmrc` that people forget to source, no host pollution.

```yaml
name: ghost-blog
version: 5.82.0
provider: nix

nix:
  packages:
    - nodejs_18

env:
  url:
    description: "Public URL Ghost serves on"
    default: "http://localhost:2368"
  database__client:
    description: "Database adapter — sqlite3 for local, mysql for staging"
    default: "sqlite3"

validations:
  - name: "Node 18"
    command: "node --version | grep -qE '^v18'"

hooks:
  setup:
    - run: "npm install -g ghost-cli@latest"
      when: first-setup
    - run: "ghost install local --no-prompt"
      when: first-setup
  after_start:
    - "ghost start"
  before_stop:
    - "ghost stop"
```

## What's happening

| Stage | Command | Why |
| :--- | :--- | :--- |
| `setup` (first time) | `npm install -g ghost-cli` | Installs the CLI inside the nix sandbox — not globally on the host. |
| `setup` (first time) | `ghost install local` | Initialises the Ghost directory, downloads content, creates SQLite DB. |
| `after_start` | `ghost start` | Launches the Ghost server process. |
| `before_stop` | `ghost stop` | Graceful shutdown — ensures SQLite is flushed before the sandbox exits. |

## MySQL variant

Swap SQLite for a containerised MySQL by switching to `provider: hybrid` and adding a compose file:

```yaml
name: ghost-blog
version: 5.82.0
provider: hybrid

nix:
  packages:
    - nodejs_18

docker:
  compose: docker-compose.yml   # starts MySQL 8

env:
  url:
    default: "http://localhost:2368"
  database__client:
    default: "mysql"
  database__connection__host:
    default: "localhost"
  database__connection__user:
    default: "ghost"
  database__connection__password:
    required: true
    default: "ghost"
  database__connection__database:
    default: "ghost_dev"

hooks:
  setup:
    - run: "npm install -g ghost-cli@latest"
      when: first-setup
    - run: "ghost install local --no-prompt --db mysql"
      when: first-setup
  after_start:
    - "ghost start"
  before_stop:
    - "ghost stop"
```
