---
layout: default
title: Getting Started
---

# 🚀 Getting Started

## 1. Prerequisites

You must have the following installed on your host OS:
* [Nix](https://nixos.org/download) (The package manager itself)
* [Docker](https://docs.docker.com/engine/install/) & Docker Compose

## 2. Installation

Clone the repository and build the binary:

```bash
git clone https://github.com/Salv4d/derrick.git
cd derrick
go build -o derrick ./cmd/derrick
sudo mv derrick /usr/local/bin/
```

## 3. The "Hello World" Project

Create a new directory for your microservice and generate a `derrick.yaml`:

```yaml
name: "hello-world"
version: "0.1.0"

dependencies:
  nix_packages:
    - "nodejs_20"
  docker_compose: "docker-compose.yml"
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
