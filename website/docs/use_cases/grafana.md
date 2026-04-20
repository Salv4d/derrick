---
title: Grafana
---

# Grafana Local Orchestration

**Grafana** poses a unique polyglot challenge (Medium) requiring both backend and frontend binaries. A local developer typically needs **Go**, heavily specific **Node.js** versions, **Yarn**, and backend SQLite/PostgreSQL connectors configured correctly before they can test a dashboard feature.

### The Derrick Solution

Derrick excels at parallel abstractions. By declaring both runtimes inside `nix_packages`, we skip tedious local system installations. We then cleanly segregate the backend builds and the frontend builds utilizing consecutive bootstrap scripts in the `setup` stage.

### The `derrick.yaml` Implementation

```yaml
---
name: "grafana-core"
version: "10.0.0"

dependencies:
  nix_packages:
    - "go_1_21"
    - "nodejs_20"
    - "yarn"
  docker_compose: "docker-compose-dev.yml"

env:
  GF_DATABASE_TYPE:
    description: "Database Engine"
    default: "sqlite3"
  GF_SECURITY_ADMIN_PASSWORD:
    description: "Default Administrator Password"
    default: "admin"

hooks:
  setup:
    - "echo 'Bootstrapping Backend...'"
    - "go mod download"
    - "go build -o bin/grafana-server ./pkg/cmd/grafana-server"
    - "echo 'Bootstrapping Frontend...'"
    - "yarn install --immutable"
    - "yarn build"
  after_start:
    - "./bin/grafana-server web" # Runs binary native off the sandbox!
```
