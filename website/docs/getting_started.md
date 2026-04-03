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

If you use `derrick` across multiple microservices (e.g., a Backend API and a separate Frontend Web app), **they instantly know how to communicate with zero configuration.**

Wait, what is the "easiest way to do it?" 
**Just boot them!** Run `derrick start` in Project A, and `derrick start` in Project B. Derrick automatically injects the `derrick-net` global Docker bridge into your projects behind the scenes without modifying your file states!

Here is how the magic works in practice:

### Example A: Container to Container (Across Projects)
Imagine your Backend API (`api/docker-compose.yml`) defines a service named `payment-worker`. 
Your separate Frontend Web app (`web/docker-compose.yml`) can immediately speak to it via standard DNS!
```javascript
// Inside your Frontend Web app
const response = await fetch("http://payment-worker:8080/charge");
```

### Example B: Container to Host Sandbox
What if your Frontend Web app isn't running in Docker, but is simply running natively in your `derrick shell` (Host OS) on `localhost:3000`? 
Historically, Docker containers struggle to route traffic backwards to the Host OS on Linux. Derrick solves this automatically:
```yaml
# Inside your Backend API docker-compose.yml
services:
  nginx-gateway:
    image: nginx:latest
```
```nginx
# Wait! How does the nginx gateway reach your native Host OS React App?
# Simply reference the globally injected host.docker.internal!
proxy_pass http://host.docker.internal:3000;
```
