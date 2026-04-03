---
layout: default
title: Getting Started
---

# 🚀 Getting Started

## 1. Prerequisites

You must have the following installed on your host OS:
* [Nix](https://nixos.org/download) (The package manager itself)
* (Optional) [Docker](https://docs.docker.com/engine/install/) & Docker Compose

## 2. Installation

**Option A: Download Pre-compiled Binary (Recommended)**
```bash
curl -L -o derrick https://github.com/Salv4d/derrick/releases/latest/download/derrick-linux-amd64
chmod +x derrick
sudo mv derrick /usr/local/bin/
```

**Option B: Build from Source**
If you have Go installed, you can compile it manually:

```bash
git clone https://github.com/Salv4d/derrick.git
cd derrick
go build -o derrick ./cmd/derrick
sudo mv derrick /usr/local/bin/
```

## 3. The "Hello World" Project

Create a new directory for your microservice and run the initialization wizard:

```bash
mkdir my-service && cd my-service
derrick init
```

The smart wizard will auto-detect your project’s language natively (Node, Go, Python, etc.) and seamlessly prompt you to optionally attach environment files or a Docker Compose stack. 

Here is an example structure it could generate:

```yaml
name: "hello-world"
version: "0.1.0"

dependencies:
  nix_packages:
    - "nodejs_20"
  # docker_compose is entirely optional!
```

Start the Derrick engine:
```bash
derrick start
```

Drop into the hermetic sandbox. Notice how `node` is available even if uninstalled on the Host OS:
```bash
derrick shell

# inside the sandbox prompt:
node -v
```

## 4. IDE Integration & AI Coding Agents ✨

To access your dependencies (like Language Servers, Linters, or Compilers) in your favorite IDE without polluting your host OS, simply launch your IDE directly using the `derrick code` command.

**Using your default Editor:**
```bash
# Automatically detects your $VISUAL or $EDITOR OS variable, or prompts you!
derrick code 
```

**Using an explicit IDE (e.g. Neovim, Cursor, VSCode, Emacs):**
```bash
derrick code neovim
```

If you don't want the IDE lockfile to persist cleanly, simply pass the `--rm` flag for an ephemeral evaluation:
```bash
derrick code --rm helix
```

*The IDE will launch securely infused with the Nix PATH! Any extension requiring `node`, `go`, or `python` will automatically resolve to the isolated sandbox dependencies while simultaneously preserving your global user settings (like `~/.config/nvim/` or `~/.vscode/`).*

**Ephemeral AI Agents:**
You can also securely run terminal-based AI coding agents inside verified sandboxes using the ephemeral `run` tool:
```bash
derrick run claude-code
```

## 5. Cross-Project Clustering 🌐

If you use `derrick` mapped to multiple microservices (e.g., `api/derrick.yaml` and `web/derrick.yaml`), **they instantly know how to communicate.**
You do not need to hardcode manual IP addresses or rewrite your `docker-compose` bridges. 

Derrick automatically injects the `derrick-net` global bridge into your compose evaluations behind the scenes. Furthermore, every Container spawned can flawlessly speak backwards to any local ports open in your `derrick shell` Host sandbox using the URL `http://host.docker.internal:<PORT>`!
