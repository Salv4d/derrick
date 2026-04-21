---
title: Grafana
---

# Grafana

**The problem:** Grafana is a Go backend + React/TypeScript frontend + Yarn toolchain. Getting three language ecosystems on the same version across a team is painful. Frontend contributors shouldn't need Go installed; backend contributors shouldn't need to `nvm use` before every session.

**The Derrick solution:** `provider: hybrid` puts Go, Node, and Yarn in the Nix sandbox together. One `derrick start` and both ecosystems are on PATH, pinned.

```yaml
name: grafana-core
version: 10.4.0
provider: hybrid

nix:
  packages:
    - go_1_22
    - nodejs_20
    - yarn

docker:
  compose: devenv/docker-compose-devenv.yml
  # starts: postgres (for alerting state), prometheus (for self-monitoring)

env:
  GF_DATABASE_TYPE:
    description: "Database engine for Grafana's internal state"
    default: "sqlite3"
  GF_SECURITY_ADMIN_PASSWORD:
    description: "Admin user password"
    default: "admin"
  GF_SERVER_HTTP_PORT:
    default: "3000"

validations:
  - name: "Go 1.22"
    command: "go version | grep -q 'go1.22'"
  - name: "Yarn"
    command: "yarn --version"
  - name: "Port 3000 free"
    command: "! lsof -i :3000"
    auto_fix: "kill -9 $(lsof -t -i:3000) 2>/dev/null || true"

hooks:
  setup:
    - run: "go mod download"
      when: first-setup
    - run: "yarn install --immutable"
      when: first-setup
    - run: "go build -o bin/grafana ./pkg/cmd/grafana"
      when: first-setup
    - run: "yarn build"
      when: first-setup
  after_start:
    - "./bin/grafana server --config conf/custom.ini cfg:default.paths.data=data"
    - "echo 'Grafana running at http://localhost:3000 (admin / $GF_SECURITY_ADMIN_PASSWORD)'"

flags:
  rebuild:
    description: "Rebuild Go binary and frontend assets from scratch"
```

## What's happening

| Stage | Command | Why |
| :--- | :--- | :--- |
| `setup` (first time) | `go mod download` | Pre-fetches module cache — faster subsequent builds. |
| `setup` (first time) | `yarn install` | Installs frontend dependencies. |
| `setup` (first time) | `go build` | Produces the binary contributors actually run. |
| `setup` (first time) | `yarn build` | Compiles frontend assets into `public/build/`. |
| `after_start` | `./bin/grafana server` | Starts Grafana against the dev config. |

## Incremental rebuilds

Add a `rebuild` flag to avoid re-running the full first-time setup:

```yaml
hooks:
  setup:
    - run: "go build -o bin/grafana ./pkg/cmd/grafana && yarn build"
      when: flag:rebuild
```

```bash
derrick start --flag rebuild
```
