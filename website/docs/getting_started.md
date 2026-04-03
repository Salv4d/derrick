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

Clone the repository and build the binary:

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

## 4. IDE Integration

To access your dependencies (like Language Servers, Linters, or Compilers) in your favorite IDE without polluting your host OS, simply launch your IDE directly from the active sandbox:

**Using VSCode:**
```bash
derrick shell
code .
```

*The IDE will launch securely infused with the Nix PATH! Any extension requiring `node`, `go`, or `python` will automatically resolve to the isolated sandbox dependencies.*
