---
layout: default
title: Getting Started
---

# Getting Started

## 1. Prerequisites

You need at least one of the following on your host OS:

* [Nix](https://nixos.org/download) — for `provider: nix` projects
* [Docker](https://docs.docker.com/engine/install/) & Docker Compose — for `provider: docker` projects

Derrick itself requires neither at install time: `derrick init` generates a config, then `derrick start` pulls in whichever tool the project declares.

## 2. Installation

**Download the pre-compiled binary (recommended)**
```bash
curl -L -o derrick https://github.com/Salv4d/derrick/releases/latest/download/derrick-linux-amd64
chmod +x derrick
sudo mv derrick /usr/local/bin/
```

**Using Go**
```bash
go install github.com/Salv4d/derrick/cmd/derrick@latest
```

**Build from source**
```bash
git clone https://github.com/Salv4d/derrick.git
cd derrick
go build -o derrick ./cmd/derrick
sudo mv derrick /usr/local/bin/
```

## 3. Your First Project

Create a new directory and run the initialization wizard:

```bash
mkdir my-service && cd my-service
derrick init
```

The wizard auto-detects your project language and generates a `derrick.yaml`. Here is a typical result:

```yaml
name: "my-service"
version: "0.1.0"
provider: nix

nix:
  packages:
    - "nodejs_22"
```

Boot the environment:
```bash
derrick start
```

Drop into the isolated shell — `node` is available even without being installed on the host:
```bash
derrick shell

# inside the sandbox:
node -v
```

## 4. Lifecycle Hooks & Custom Flags

Hooks let you automate setup tasks that run at specific moments. The `when:` condition controls when each hook fires:

```yaml
hooks:
  start:
    - run: "npm install"
      when: first-setup    # only on the very first boot
    - run: "npm run build"
      when: always
    - run: "npm run seed"
      when: flag:seed-db   # only when --flag seed-db is passed

flags:
  seed-db:
    description: "Populate the database with development seed data"
```

```bash
# Normal start — runs npm install (first time only) and npm run build
derrick start

# Start with seed data
derrick start --flag seed-db
```

Derrick persists state in `.derrick/state.json` to track whether first-setup has already completed, so `when: first-setup` hooks never run twice accidentally.

## 5. Provider Selection

Derrick selects the isolation backend automatically with `provider: auto` (the default):

| Config | Backend chosen |
| :--- | :--- |
| `docker.compose` is set | Docker Compose |
| Only `nix.packages` set | Nix dev shell |
| Both set, `provider: auto` | Docker |

Override it explicitly when you need to be unambiguous:

```yaml
provider: docker   # always Docker
provider: nix      # always Nix
provider: auto     # Derrick decides (default)
```

## 6. Starting a Hub Project

If a project alias is registered in `~/.derrick/config.yaml`, you can clone and boot it in one command:

```bash
# Clones the repo and runs `derrick start` inside it
derrick start auth-service
```

Register aliases globally:
```yaml
# ~/.derrick/config.yaml
projects:
  auth-service: https://github.com/your-org/auth-service.git
  payment-api: https://github.com/your-org/payment-api.git
```

## 7. Cross-Project Clustering

Run `derrick start` in multiple microservice directories. Derrick automatically attaches all Docker Compose projects to a shared `derrick-net` bridge network, so containers across projects can resolve each other by service name:

```javascript
// Frontend container talking to a backend service in another project
const response = await fetch("http://payment-worker:8080/charge");
```

Containers can also reach host-native processes (running in your Nix shell) via `host.docker.internal`, which Derrick injects automatically.

## 8. IDE & Tool Integration

### Automatic (recommended): direnv

When `derrick start` runs a Nix project for the first time, it writes a `.envrc` at the project root. With [direnv](https://direnv.net) and [nix-direnv](https://github.com/nix-community/nix-direnv) installed, the Nix environment activates automatically whenever any shell — your editor, Claude Code, a terminal — enters the project directory. No `derrick shell` wrapper needed.

```bash
# One-time setup per machine
# Install direnv: https://direnv.net/docs/installation.html
# Install nix-direnv: https://github.com/nix-community/nix-direnv#installation

# One-time per project, after derrick start writes .envrc:
direnv allow
```

After that, tools like Claude Code's Bash tool, VS Code's integrated terminal, and language server processes all see the project's binaries on `PATH` without any manual wrapping. The `.envrc` file is safe to commit so your teammates get the same behaviour.

### Manual: launch from inside the sandbox

If you prefer not to use direnv, launch your editor or AI tool from inside the active environment:

```bash
derrick shell
code .       # VS Code
claude       # Claude Code
nvim .
```

Language servers, linters, and compilers resolve to the sandboxed versions without polluting the host OS.

### Quick throwaway sandboxes

No `derrick.yaml` needed:
```bash
# Interactive shell with jq and python
derrick run jq python3

# Run a command directly without entering a shell
derrick run python3 -- python -c "print('hello')"
```
