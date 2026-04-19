# Derrick — Manual Testing Guide

This guide walks through every major feature with exact commands and expected outcomes,
then validates Derrick against six real public projects covering both providers and a
range of tech stacks.

---

## Prerequisites

```bash
# Build the binary from source
git clone https://github.com/Salv4d/derrick.git
cd derrick
go build -o derrick ./cmd/derrick
sudo mv derrick /usr/local/bin/

# Verify
derrick version
```

You will need at least one of:
- **Nix** — for any test marked `[nix]`
  ```bash
  curl --proto '=https' --tlsv1.2 -sSf -L https://install.determinate.systems/nix | sh -s -- install
  ```
- **Docker + Docker Compose** — for any test marked `[docker]`

---

## 1. Core Command Smoke Tests

### 1.1 Help & version

```bash
derrick --help
derrick version
derrick start --help     # should show --reset and --flag flags
derrick stop --help
derrick run --help
```

**What to check:** `start --help` must list `--reset`, `--flag strings`, and the `[alias]` argument.

---

### 1.2 `derrick init` wizard

```bash
mkdir /tmp/derrick-test && cd /tmp/derrick-test
derrick init
```

Walk through the wizard:
- Set project name to `test-project`
- Select a few Nix packages (e.g. `nodejs_22`, `git`)
- Skip Docker Compose
- Skip env management

**Expected result:** `derrick.yaml` generated in the current directory with the new
top-level `nix:` block (not `dependencies:`).

```bash
cat derrick.yaml
# Must contain:
#   name: test-project
#   nix:
#     packages:
#       - nodejs_22
#       - git
```

---

### 1.3 `derrick doctor` `[nix]`

```bash
cd /tmp/derrick-test
derrick doctor
```

**Expected:** Green checks for Nix and the listed packages. No errors.

---

### 1.4 `derrick start` with Nix `[nix]`

```bash
cd /tmp/derrick-test
derrick start
```

**Expected sequence:**
1. `Configuration` section loads the YAML
2. `nix` provider is selected
3. Nix packages resolve and lock
4. `Environment variables loaded`
5. Final `test-project is ready!` line

**Check state was persisted:**
```bash
cat .derrick/state.json
# Must contain:
#   "first_setup_completed": true
#   "provider": "nix"
#   "status": "running"
```

---

### 1.5 `derrick start` a second time — first-setup skipped

```bash
derrick start
```

**Expected:** Any hook marked `when: first-setup` does NOT fire. The state file already
has `first_setup_completed: true`.

---

### 1.6 `derrick shell` `[nix]`

```bash
derrick shell
# Inside the shell:
node --version     # resolves to the Nix-managed binary
which node         # path must be inside /nix/store/...
exit
```

**What to check:** `node` is available and its path is a Nix store path, not `/usr/bin/node`.

---

### 1.7 `derrick stop`

```bash
derrick stop
cat .derrick/state.json
# "status" must now be "stopped"
```

---

### 1.8 `derrick run` — ephemeral sandbox `[nix]`

```bash
# Interactive shell with ad-hoc packages
derrick run python3 jq

# Execute a command directly
derrick run python3 -- python3 -c "import sys; print(sys.version)"

# Save the environment to a directory
derrick run --save nodejs_22 yarn
```

**What to check:**
- The ephemeral shell has `python3` and `jq` available
- `--` separator correctly passes the command
- `--save` creates a `derrick-env-<timestamp>/` directory with a `derrick.yaml`

---

## 2. Lifecycle Hook Tests

Create this config and run through each scenario:

```bash
mkdir /tmp/hook-test && cd /tmp/hook-test
cat > derrick.yaml << 'EOF'
name: hook-test
version: "1.0.0"
provider: nix

nix:
  packages:
    - coreutils

hooks:
  start:
    - run: "echo '[hook] always fires'"
      when: always
    - run: "echo '[hook] first-setup only'"
      when: first-setup
    - run: "echo '[hook] seed flag active'"
      when: flag:seed
  stop:
    - run: "echo '[hook] stop hook'"
      when: always

flags:
  seed:
    description: "Run the seed hook"
EOF
```

### 2.1 First run — all three hooks fire

```bash
rm -f .derrick/state.json
derrick start
```

**Expected output (in order):**
```
[hook] always fires
[hook] first-setup only
```
> Note: `seed` flag is not active, so its hook is skipped.

---

### 2.2 Second run — first-setup hook is skipped

```bash
derrick start
```

**Expected output:**
```
[hook] always fires
```
`[hook] first-setup only` must NOT appear.

---

### 2.3 Custom flag activates its hook

```bash
derrick start --flag seed
```

**Expected output:**
```
[hook] always fires
[hook] seed flag active
```

---

### 2.4 Stop hooks fire

```bash
derrick stop
```

**Expected:**
```
[hook] stop hook
```

---

### 2.5 Reset flag

Add this to `derrick.yaml` under `flags:`:
```yaml
  reset:
    description: "Reset the environment"
```
And add to `hooks.start`:
```yaml
    - run: "echo '[hook] rebuilding from scratch'"
      when: flag:reset
```

```bash
derrick start --reset --flag reset
```

**Expected:** `[hook] rebuilding from scratch` appears.

---

## 3. Error Handling Tests

### 3.1 Docker socket permission `[docker]`

If your user is NOT in the `docker` group, run:

```bash
mkdir /tmp/docker-test && cd /tmp/docker-test
cat > derrick.yaml << 'EOF'
name: docker-test
version: "1.0.0"
provider: docker

docker:
  compose: docker-compose.yml
EOF

cat > docker-compose.yml << 'EOF'
services:
  app:
    image: nginx:alpine
    ports:
      - "8181:80"
EOF

sudo chmod 000 /var/run/docker.sock   # simulate denial
derrick start
sudo chmod 660 /var/run/docker.sock   # restore
```

**Expected:** A `DerrickError` with the message `Docker socket permission denied` and
the exact fix command `sudo usermod -aG docker $USER && newgrp docker`.

---

### 3.2 Missing YAML

```bash
cd /tmp && derrick start
```

**Expected:** A clear error that `derrick.yaml` was not found — not a raw Go panic.

---

### 3.3 Invalid YAML syntax

```bash
mkdir /tmp/bad-yaml && cd /tmp/bad-yaml
printf 'name: test\nversion: "1.0"\nnix:\n\tpackages: [go]\n' > derrick.yaml
derrick start
```

**Expected:** Syntax error with line number, the offending line printed, and a `^` indicator.

---

### 3.4 Nix package does not exist `[nix]`

```bash
mkdir /tmp/bad-pkg && cd /tmp/bad-pkg
cat > derrick.yaml << 'EOF'
name: bad-pkg
version: "1.0.0"
provider: nix
nix:
  packages:
    - this-package-does-not-exist-xyz
EOF
derrick start
```

**Expected:** Error mentions `attribute ... missing` and the fix points to
`search.nixos.org`.

---

## 4. Docker Compose Test `[docker]`

```bash
mkdir /tmp/compose-test && cd /tmp/compose-test

cat > docker-compose.yml << 'EOF'
services:
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_PASSWORD: secret
      POSTGRES_DB: testdb
    ports:
      - "5433:5432"
  redis:
    image: redis:7-alpine
    ports:
      - "6380:6379"
EOF

cat > derrick.yaml << 'EOF'
name: compose-test
version: "1.0.0"
provider: docker

docker:
  compose: docker-compose.yml

hooks:
  start:
    - run: "echo 'Waiting for Postgres...' && sleep 2 && docker exec $(docker ps -qf name=db) pg_isready"
      when: first-setup
  stop:
    - run: "echo 'Containers stopped.'"
      when: always

validations:
  - name: "Docker daemon"
    command: "docker info"
EOF

derrick start
```

**Expected:**
- `derrick-net` bridge network created
- Both containers start: `db` and `redis`
- `first-setup` hook fires once and waits for Postgres
- Check cross-project networking: `docker network inspect derrick-net`

```bash
derrick stop
# Verify containers are gone
docker ps | grep -E "db|redis" && echo "FAIL: still running" || echo "PASS: stopped"
```

---

## 5. Public Project Tests

> For each project, clone the repo first, drop the `derrick.yaml` into the root, then run `derrick start`.

---

### 5.1 Ghost CMS `[docker]`

**Repo:** https://github.com/TryGhost/Ghost

```bash
git clone --depth=1 https://github.com/TryGhost/Ghost.git /tmp/ghost
cd /tmp/ghost
```

**`derrick.yaml`:**
```yaml
name: "ghost"
version: "1.0.0"
provider: docker

docker:
  compose: docker-compose.yml

env:
  database__client:
    default: "mysql"
  database__connection__password:
    default: "ghostpassword"

hooks:
  start:
    - run: "echo 'Ghost is starting... visit http://localhost:2368 in a few seconds'"
      when: always
  stop:
    - run: "echo 'Ghost stopped.'"
      when: always
```

> Ghost ships a `docker-compose.yml` at the repo root for local development.

```bash
cat > derrick.yaml << 'EOF'
name: "ghost"
version: "1.0.0"
provider: docker

docker:
  compose: docker-compose.yml

hooks:
  start:
    - run: "echo 'Ghost is starting... visit http://localhost:2368 in ~30s'"
      when: always
    - run: "docker compose -f docker-compose.yml logs -f ghost &"
      when: flag:logs
  stop:
    - run: "echo 'Ghost stopped.'"
      when: always

flags:
  logs:
    description: "Tail Ghost container logs after boot"
EOF

derrick start
```

**What to check:**
- Containers `ghost` and `ghost-db` appear in `docker ps`
- After ~30s, `http://localhost:2368` returns the Ghost homepage
- `derrick start --flag logs` streams container logs

```bash
derrick stop
```

---

### 5.2 Plausible Analytics `[docker]`

**Repo:** https://github.com/plausible/community-edition

```bash
git clone --depth=1 https://github.com/plausible/community-edition.git /tmp/plausible
cd /tmp/plausible
```

**`derrick.yaml`:**
```yaml
name: "plausible"
version: "1.0.0"
provider: docker

docker:
  compose: docker-compose.yml

env_management:
  base_file: "plausible-conf.env"
  prompt_missing: false

env:
  SECRET_KEY_BASE:
    description: "64-char random secret. Generate with: openssl rand -base64 48"
    required: true
  BASE_URL:
    description: "Public URL of this Plausible instance"
    required: true
    default: "http://localhost:8000"

hooks:
  start:
    - run: "[ -f plausible-conf.env ] || cp plausible-conf.env.example plausible-conf.env 2>/dev/null || echo 'SECRET_KEY_BASE=\nBASE_URL=http://localhost:8000' > plausible-conf.env"
      when: first-setup
    - run: "echo 'Plausible is starting... visit http://localhost:8000'"
      when: always
  stop:
    - run: "echo 'Plausible stopped.'"
      when: always
```

```bash
cat > derrick.yaml << 'YAML'
name: "plausible"
version: "1.0.0"
provider: docker

docker:
  compose: docker-compose.yml

hooks:
  start:
    - run: "[ -f plausible-conf.env ] || (echo 'SECRET_KEY_BASE=$(openssl rand -base64 48)' > plausible-conf.env && echo 'BASE_URL=http://localhost:8000' >> plausible-conf.env)"
      when: first-setup
    - run: "echo 'Plausible starting... visit http://localhost:8000'"
      when: always
  stop:
    - run: "echo 'Plausible stopped.'"
      when: always
YAML

derrick start
```

**What to check:**
- `http://localhost:8000` returns the Plausible dashboard after ~20s
- Registration page is accessible at `http://localhost:8000/register`
- Run `derrick start` a second time — the `first-setup` env file creation hook does NOT fire again

```bash
derrick stop
```

---

### 5.3 Supabase `[docker]`

**Repo:** https://github.com/supabase/supabase

> Supabase is the hardest test: 10+ containers, custom networks, Kong API gateway. If Derrick handles this cleanly, it handles anything.

```bash
git clone --depth=1 https://github.com/supabase/supabase.git /tmp/supabase
cd /tmp/supabase/docker
```

**`derrick.yaml`** (place inside `docker/`):
```yaml
name: "supabase"
version: "1.0.0"
provider: docker

docker:
  compose: docker-compose.yml

env_management:
  base_file: ".env.example"
  prompt_missing: false

hooks:
  start:
    - run: "[ -f .env ] || cp .env.example .env"
      when: first-setup
    - run: "echo 'Supabase starting — this takes 60–90s on first pull'"
      when: first-setup
    - run: "echo 'Supabase ready. Studio: http://localhost:3000 | API: http://localhost:8000'"
      when: always
  stop:
    - run: "echo 'Supabase stopped. Data persisted in Docker volumes.'"
      when: always

flags:
  reset:
    description: "Wipe all volumes and start fresh (destructive)"

validations:
  - name: "Docker daemon"
    command: "docker info"
  - name: "Docker Compose v2"
    command: "docker compose version"
```

```bash
cat > derrick.yaml << 'YAML'
name: "supabase"
version: "1.0.0"
provider: docker

docker:
  compose: docker-compose.yml

hooks:
  start:
    - run: "[ -f .env ] || cp .env.example .env"
      when: first-setup
    - run: "echo 'Supabase starting (60-90s on first pull)...'"
      when: first-setup
    - run: "echo 'Studio → http://localhost:3000 | API → http://localhost:8000'"
      when: always
  stop:
    - run: "echo 'Supabase stopped. Data is safe in Docker volumes.'"
      when: always

flags:
  reset:
    description: "Wipe volumes and start fresh"

validations:
  - name: "Docker Compose v2"
    command: "docker compose version"
YAML

derrick start
```

**What to check:**
- `docker ps` shows ~10 containers (studio, kong, auth, rest, realtime, storage, imgproxy, meta, functions, analytics)
- `http://localhost:3000` — Supabase Studio UI
- `http://localhost:8000` — Kong API gateway
- Second `derrick start` skips the first-setup hooks and boots in seconds
- `derrick stop` shuts all containers cleanly

```bash
derrick stop
```

---

### 5.4 Gitea `[nix]`

**Repo:** https://github.com/go-gitea/gitea

> Pure Go project. Demonstrates the Nix provider replacing a full Go toolchain without `apt install golang`.

```bash
git clone --depth=1 https://github.com/go-gitea/gitea.git /tmp/gitea
cd /tmp/gitea
```

**`derrick.yaml`:**
```yaml
name: "gitea"
version: "1.0.0"
provider: nix

nix:
  packages:
    - "go"
    - "nodejs_22"
    - "gnumake"
    - "git"
    - "sqlite"

hooks:
  start:
    - run: "go mod download"
      when: first-setup
    - run: "make build"
      when: first-setup
    - run: "echo 'Gitea built. Run: ./gitea web'"
      when: always

validations:
  - name: "Go compiler"
    command: "go version"
  - name: "Make"
    command: "make --version"
```

```bash
cat > derrick.yaml << 'YAML'
name: "gitea"
version: "1.0.0"
provider: nix

nix:
  packages:
    - "go"
    - "nodejs_22"
    - "gnumake"
    - "git"
    - "sqlite"

hooks:
  start:
    - run: "go mod download"
      when: first-setup
    - run: "make build"
      when: first-setup
    - run: "echo 'Binary ready at ./gitea — run: ./gitea web'"
      when: always

validations:
  - name: "Go compiler"
    command: "go version"
YAML

derrick start
```

**What to check:**
- `go` and `make` resolve to Nix store paths (not system binaries):
  ```bash
  derrick shell
  which go     # must be /nix/store/...
  go version
  exit
  ```
- `./gitea` binary exists after the first-setup hooks
- Second `derrick start` skips `go mod download` and `make build`

```bash
derrick stop
```

---

### 5.5 n8n `[docker]`

**Repo:** https://github.com/n8n-io/n8n

> Workflow automation platform. Tests Docker Compose with persistent volume and port validation.

```bash
git clone --depth=1 https://github.com/n8n-io/n8n.git /tmp/n8n
cd /tmp/n8n
```

**`derrick.yaml`:**
```yaml
name: "n8n"
version: "1.0.0"
provider: docker

docker:
  compose: docker-compose.yml

ports:
  - 5678

hooks:
  start:
    - run: "echo 'n8n starting... visit http://localhost:5678'"
      when: always
    - run: "echo 'First run: create your owner account at http://localhost:5678/setup'"
      when: first-setup
  stop:
    - run: "echo 'n8n stopped. Workflows saved in Docker volume.'"
      when: always

validations:
  - name: "Port 5678 is free"
    command: "! lsof -i :5678 2>/dev/null | grep LISTEN"
```

```bash
cat > derrick.yaml << 'YAML'
name: "n8n"
version: "1.0.0"
provider: docker

docker:
  compose: docker-compose.yml

ports:
  - 5678

hooks:
  start:
    - run: "echo 'First run: set up your account at http://localhost:5678/setup'"
      when: first-setup
    - run: "echo 'n8n ready → http://localhost:5678'"
      when: always
  stop:
    - run: "echo 'n8n stopped. Workflows are safe in the Docker volume.'"
      when: always

validations:
  - name: "Port 5678 is free"
    command: "! lsof -i :5678 2>/dev/null | grep LISTEN"
YAML

derrick start
```

**What to check:**
- `http://localhost:5678` serves the n8n UI
- `http://localhost:5678/setup` prompts for owner account on first run
- The `first-setup` message does NOT appear on subsequent starts
- `derrick stop` leaves n8n data intact in the Docker volume
- `derrick start --flag reset` + manual `docker volume rm` workflow resets state

```bash
derrick stop
```

---

### 5.6 Grafana `[nix + docker]`

**Repo:** https://github.com/grafana/grafana

> Go + TypeScript monorepo. Demonstrates `provider: auto` — Nix for the toolchain, Docker for the data sources.

```bash
git clone --depth=1 https://github.com/grafana/grafana.git /tmp/grafana
cd /tmp/grafana
```

**`derrick.yaml`:**
```yaml
name: "grafana"
version: "1.0.0"
provider: nix

nix:
  packages:
    - "go"
    - "nodejs_22"
    - "yarn"
    - "gnumake"
    - "wire"

hooks:
  start:
    - run: "go mod download && yarn install --immutable"
      when: first-setup
    - run: "make gen-go"
      when: first-setup
    - run: "echo 'Grafana dev environment ready.'"
      when: always

flags:
  frontend:
    description: "Start the frontend dev server (yarn start)"
  backend:
    description: "Start the backend dev server (make run)"

validations:
  - name: "Go compiler"
    command: "go version"
  - name: "Node.js"
    command: "node --version"
  - name: "Yarn"
    command: "yarn --version"
```

```bash
cat > derrick.yaml << 'YAML'
name: "grafana"
version: "1.0.0"
provider: nix

nix:
  packages:
    - "go"
    - "nodejs_22"
    - "yarn"
    - "gnumake"

hooks:
  start:
    - run: "go mod download"
      when: first-setup
    - run: "yarn install --immutable"
      when: first-setup
    - run: "echo 'Grafana dev env ready. Enter shell: derrick shell'"
      when: always

validations:
  - name: "Go compiler"
    command: "go version"
  - name: "Node.js"
    command: "node --version"
YAML

derrick start
derrick shell

# Inside the Nix shell — all tools resolve to Nix store:
go version
node --version
yarn --version
make --version

exit
```

**What to check:**
- All binaries resolve to `/nix/store/...` paths — zero host pollution
- `go mod download` only ran once (first-setup)
- Re-entering `derrick shell` is instant (no re-download)
- `derrick doctor` shows all validations passing

```bash
derrick stop
```

---

## 6. Hub Alias Test

```bash
# Register an alias globally
mkdir -p ~/.derrick
cat > ~/.derrick/config.yaml << 'EOF'
projects:
  ghost: https://github.com/TryGhost/Ghost.git
EOF

# Clone and start in one command from any directory
cd /tmp
derrick start ghost
```

**Expected:** Derrick clones Ghost into `/tmp/Ghost/`, changes into it, and runs
`derrick start` — but since there is no `derrick.yaml` in Ghost's repo yet, it will
fail cleanly with a missing-file error. This validates the Hub resolution path works.

To make it boot fully, add the `derrick.yaml` from section 5.1 first.

---

## 7. Profile Test

```bash
mkdir /tmp/profile-test && cd /tmp/profile-test
cat > derrick.yaml << 'EOF'
name: "profile-test"
version: "1.0.0"
provider: nix

nix:
  packages:
    - "coreutils"

hooks:
  start:
    - run: "echo 'Base environment'"
      when: always

profiles:
  ci:
    nix:
      packages:
        - "git"
    hooks:
      start:
        - run: "echo 'CI profile active'"
          when: always
EOF

# Default start
derrick start
# Expected: "Base environment" only

# Profile start
derrick start --profile ci
# Expected: "Base environment" + "CI profile active", git available
```

---

## 8. Cleanup

```bash
# Remove all test directories
rm -rf /tmp/derrick-test /tmp/hook-test /tmp/compose-test \
       /tmp/ghost /tmp/plausible /tmp/supabase /tmp/gitea \
       /tmp/n8n /tmp/grafana /tmp/bad-yaml /tmp/bad-pkg \
       /tmp/profile-test /tmp/docker-test

# Remove the global Hub entry added in section 6
rm -f ~/.derrick/config.yaml

# Remove the shared Docker network
docker network rm derrick-net 2>/dev/null || true

# Nix garbage collection (optional — frees disk space)
nix-collect-garbage -d
```
