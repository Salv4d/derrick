<p align="center">
  <img src="assets/logo.jpg" alt="Derrick CLI Logo" width="200">
</p>

<h1 align="center">Derrick CLI</h1>

<p align="center">
  <strong>A professional-grade local development orchestrator for complex microservice environments.</strong><br>
  <em>Unite the absolute reproducibility of Nix with the optional containerization of Docker Compose.</em>
</p>

<p align="center">
  <a href="https://salv4d.github.io/derrick/"><img src="https://img.shields.io/badge/📖_Documentation-salv4d.github.io/derrick-blue.svg?style=for-the-badge" alt="Documentation"></a>
  <a href="https://github.com/Salv4d/derrick/actions/workflows/deploy-docs.yml"><img src="https://img.shields.io/github/actions/workflow/status/Salv4d/derrick/deploy-docs.yml?style=for-the-badge&label=Docs%20CI" alt="Docs CI"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-purple.svg?style=for-the-badge" alt="License: MIT"></a>
  <a href="https://goreportcard.com/report/github.com/Salv4d/derrick"><img src="https://goreportcard.com/badge/github.com/Salv4d/derrick?style=for-the-badge" alt="Go Report Card"></a>
  <img src="https://img.shields.io/badge/Stability-Alpha-orange.svg?style=for-the-badge" alt="Stability: Alpha">
</p>

---

## 💡 The Problem Derrick Solves

Unlike generic task runners, Derrick ensures that every contributor's machine is a bit-for-bit clone of the production-grade toolchain. 

1. **Zero Host OS Pollution:** No `nvm`, `pyenv`, or global `go` required. All your project dependencies live strictly inside an isolated Nix sandbox.
2. **Declarative Contracts:** Define your environment in `derrick.yaml`. If the contract says you need Go 1.21 and Postgres, Derrick guarantees that state.
3. **Fail-Fast Validation:** Audits local constraints like `.env` secrets or active ports in milliseconds before booting the environment, prompting self-healing fixes.

---

## ⚡ Quick Start

Experience a fully hermetic environment in seconds. 

> *Note: Requires [Nix](https://nixos.org/download) to be previously installed on your machine. Docker is optional!*

### 1. Installation

**Download Pre-compiled Binary (Recommended)**
```bash
curl -L -o derrick https://github.com/Salv4d/derrick/releases/latest/download/derrick-linux-amd64
chmod +x derrick
sudo mv derrick /usr/local/bin/
```

**Using Go**
```bash
go install github.com/Salv4d/derrick/cmd/derrick@latest
```

**Build from Source**
```bash
git clone https://github.com/Salv4d/derrick.git
cd derrick
go build -o derrick ./cmd/derrick
sudo mv derrick /usr/local/bin/
```

### 2. Enter the Sandbox

Navigate to your project folder and run the smart wizard to generate your `derrick.yaml`:

```bash
# Auto-detects your project language and optional external databases
derrick init

# Formats and starts Nix binaries and Docker containers
derrick start

# Drops you into the strictly sealed bash terminal (your sandbox)
derrick shell

# Need an ad-hoc throwaway sandbox with specific tools?
derrick run python3 jq

# Audits if your project meets the strict derrick.yaml constraints
derrick doctor

# Free up disk space with the universal garbage collector
derrick clean
```

### 3. IDE Integration & AI Coding Agents ✨

After running `derrick start`, you can open your project in any editor from within the hermetic sandbox. All dependencies (Language Servers, Linters, Compilers) are available on the PATH without polluting your Host OS.

**Open your project from the sandbox:**
```bash
# Enter the sandbox, then launch your editor
derrick shell
$EDITOR .    # or e.g. code ., nvim ., etc.
```

---

## 📖 Deep-Dive Documentation

For advanced features, lifecycle hooks, complex `derrick.yaml` directives, and the architecture design, visit our official documentation portal:

👉 **[https://salv4d.github.io/derrick/](https://salv4d.github.io/derrick/)**

*(Or browse the raw Markdown files locally in the [`/website/docs`](./website/docs/getting_started.md) folder).*

---

## 🚧 Status & Roadmap

Derrick is currently in **Alpha**. It is stable for Linux/WSL environments.

- [x] Nix + Docker Compose Orchestration
- [x] Interactive Environment Validation & `.env` Setup
- [x] Custom Config YAML Support (`-f` flag)
- [x] **TUI Dashboard:** A live BubbleTea-powered container and lifecycle log viewer (`derrick dashboard`).
- [x] **Project Clustering:** Docker & Host-native global network bridging.
- [x] **Project Clustering:** Docker & Host-native global network bridging.
- [ ] **Remote Config Extensions:** Inherit base YAML settings from remote URLs securely.
- [ ] **Cloud Workspace Provisioning:** Sync your local sandbox state directly to cloud VMs.

---

## 🤝 Contributing

We heavily welcome contributions and improvements!
Read our fully detailed **[Contributing Guide](./website/docs/contributing.md)** to learn how to properly set up your environment, write tests, and open Conventional pull requests.

*We also maintain a set of [Benchmark Projects](.derrick/benchmark_projects.md) to test Derrick against strict, enterprise-style scenarios.*

---

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.
