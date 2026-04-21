---
sidebar_position: 8
title: Troubleshooting
---

# Troubleshooting

Common problems and their fixes. If you hit something not listed here, try `derrick doctor` — it audits the environment against `derrick.yaml` without booting it and often surfaces the issue directly.

## First stop: `derrick doctor`

```bash
derrick doctor
```

Reports stale state, missing binaries, drifted compose files, and broken hook references. In v0.3.0+ it exits non-zero when issues are found, so you can gate CI on it:

```bash
derrick doctor --json | jq '.issues'
```

## Docker errors

### `Cannot connect to the Docker daemon`

The daemon isn't running or your user can't reach the socket.

- **Linux:** `sudo systemctl start docker` (and enable it for reboots: `sudo systemctl enable docker`).
- **macOS / Windows:** Start Docker Desktop from the applications folder.
- **WSL2:** Make sure Docker Desktop's WSL integration is enabled for your distro.

### `permission denied while trying to connect to the Docker daemon socket`

Your user isn't in the `docker` group.

```bash
sudo usermod -aG docker $USER
newgrp docker   # or log out / back in
```

### `bind: address already in use`

Another process is already on the host port. Derrick does **not** auto-remap ports — that's by design, so you notice the conflict.

```bash
# Find the offender (example: port 5432)
sudo lsof -i :5432
# Either stop it, or change the port mapping in docker-compose.yml
```

When running multiple derrick projects concurrently, use distinct host ports per project or bind to `127.0.0.1` only.

### `pull access denied for <image>`

Either the image name is wrong or the registry needs auth.

```bash
docker login                         # for Docker Hub
docker login ghcr.io                 # for GHCR
docker login registry.gitlab.com     # for GitLab
```

Double-check the image name and tag in `docker-compose.yml`.

## Nix errors

### `attribute '<name>' missing`

The nix package you named doesn't exist in the pinned nixpkgs revision. Search the real name at [search.nixos.org](https://search.nixos.org/packages).

Common gotchas:

- Attribute names use underscores in versions (`nodejs_22`, not `nodejs-22`).
- Some packages moved: `yarn` → `yarn-berry`, `nodePackages.X` → `nodejs.pkgs.X`.
- EOL versions need explicit pinning — see the [`nix.packages` reference](./api_reference.md).

### `error: flake '<path>' does not provide attribute 'devShells.<system>.default'`

Usually means `.derrick/flake.nix` was generated for a different OS/arch than you're running on. Fix:

```bash
derrick stop
rm -rf .derrick
derrick start   # regenerates the flake for the current host
```

### `/nix/store: No space left on device`

The Nix store accumulates old derivations over time.

```bash
derrick clean           # prunes derrick-managed resources only
nix-collect-garbage -d  # host-wide Nix GC (more aggressive)
```

### Nix shell activation is slow on first boot

First boot downloads and evaluates the full toolchain. This is one-time per pinned nixpkgs revision; subsequent boots hit the shared `/nix/store` cache and are near-instant. If you have many projects pinning the same nixpkgs revision, they share derivations on disk.

## State errors

### `state.Load: cannot acquire lock`

Another `derrick` process is operating on the same project. Wait for it, or if you believe it's a stale lock (previous process crashed):

```bash
rm .derrick/state.lock
```

### `state: malformed JSON in .derrick/state.json`

Usually means the state file was edited by hand or a prior crash corrupted it. v0.3.0+ handles this gracefully — `state.Load` returns a non-nil zero-state so `derrick status` / `stop` still work. Reset cleanly:

```bash
derrick stop
rm .derrick/state.json
derrick start
```

## Hook errors

### A `when: first-setup` hook ran twice

It shouldn't — but it can happen if `derrick stop` didn't persist cleanly after the first start. Check:

```bash
cat .derrick/state.json | jq .first_setup_completed
```

If `false` but you know setup completed, the hook will re-run. To force the flag without re-running the hook, edit the file and set `"first_setup_completed": true`.

### A hook referencing a flag never fires

Flags must be declared in `derrick.yaml` *and* passed at the CLI:

```yaml
# derrick.yaml
flags:
  seed-db:
    description: "Populate the database with seed data"

hooks:
  after_start:
    - run: "./scripts/seed.sh"
      when: flag:seed-db
```

```bash
derrick start --flag seed-db
```

Flag names are exact-match — `seed_db` and `seed-db` are different flags.

## `derrick shell` errors

### On a docker-only project, `derrick shell` exits with a compose error

Derrick exec's into a service. By default it picks the first service in the compose file; override explicitly if that's not what you want:

```yaml
docker:
  compose: ./docker-compose.yml
  shell: api   # service name
```

### `derrick shell -- <cmd>` runs the wrong tool

When your project is `provider: hybrid`, `shell` routes to the **nix** leg (that's where your language tools are, not inside containers). If you need to run a command inside a container, use `docker compose exec` directly or set `provider: docker`.

## Multi-project issues

### Containers from project A can't resolve project B's service names

By default, projects are isolated — each gets its own compose network. You have three ways to let them talk:

**1. Through the host (no config required).** Every derrick-managed container gets `host.docker.internal:host-gateway` injected, so A can reach B's published ports:

```
http://host.docker.internal:<port>
```

**2. Declare a shared external network in both projects.** Both declare the same name under `docker.networks`; Derrick creates it on first start (labelled `com.derrick.managed=true`) and attaches every service:

```yaml
# In both projects' derrick.yaml
docker:
  networks:
    - shared-infra
```

Now A's services can reach B's by service name, and vice-versa.

**3. Use `requires` with `connect: true`.** When A declares B as a requirement, Derrick boots B first and auto-wires a shared network `derrick-A` between them — no config needed on B:

```yaml
# Project A only
requires:
  - name: project-b
    connect: true   # default — can be omitted
```

Set `connect: false` when B only needs to be booted (e.g. a CLI helper reachable via the host).

### `derrick clean` in one project removed another project's containers

It shouldn't, and in v0.3.0+ it doesn't. Clean filters by the `com.derrick.managed=true` label. If you see cross-project deletion, you're either on an old version (`derrick update`) or the containers in question were created before the label landed — restart them with `derrick stop && derrick start` to re-apply.

## Still stuck?

- Run with `--debug` to see raw subprocess output: `derrick --debug start`.
- Check the [glossary](./glossary.md) for term definitions.
- Open an issue with `derrick doctor --json` output attached: [github.com/Salv4d/derrick/issues](https://github.com/Salv4d/derrick/issues).
