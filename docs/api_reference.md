---
layout: default
title: API & Config Reference
---

# ⚙️ API & Configuration Reference

## CLI Usage

Derrick uses standard command-line verbs. Run `derrick --help` in the terminal for standard flags and documentation.

* `derrick start`: Triggers the engine lifecycle, validates configurations, boots Nix, and starts Docker services.
* `derrick shell`: Drops you into an interactive bash terminal hermetically sealed with the dependencies defined in your project contract.
* `derrick stop`: Elegantly tears down the project infrastructure (e.g., gracefully halts Docker Compose containers).
* `derrick doctor`: An interactive CLI that runs validations and diagnoses your current environment against the `derrick.yaml` requirements.
* `derrick dashboard`: Launches an interactive BubbleTea monitoring pane.

---

## The `derrick.yaml` Contract

The `derrick.yaml` file supports several complex schemas managed by `internal/config`:

### Env Directives
Allows strict enforcement of `.env` limits ensuring no developer misses a token:
```yaml
env:
  GITHUB_TOKEN:
    description: "Token for private module access"
    required: true
    validation: "curl -s --fail -H \"Authorization: token $GITHUB_TOKEN\" https://api.github.com/notifications"
```

### Hooks
Arbitrary shell scripts executed at specific lifecycle intervals, providing escape hatches:
```yaml
hooks:
  pre_init:
    - "echo '🔍 Checking custom structures before booting Docker...'"
```

### Validations
Fail-fast rules that must cleanly execute prior to entering the sandbox:
```yaml
validations:
  - name: "Port 8080 Available"
    command: "! lsof -i :8080"
```
