---
layout: default
title: Glossary
---

# 📚 Glossary

To avoid semantic ambiguity when discussing Derrick's core mechanics, we use strict definitions for specific terms.

* **Sandbox**: Refers specifically to the ephemeral `nix-shell`/`nix develop` environment spawned by Derrick, strictly isolating dependencies from the Host OS.
* **Engine**: The core Go logic layer (`internal/engine`) responsible for transitioning between orchestration states (Validating -> Bootstrapping Nix -> Bootstrapping Docker -> Ready).
* **Host Pollution**: The anti-pattern of installing global binaries (e.g., `apt get install nodejs`, `nvm`) natively on the developer's operating system. Derrick fundamentally battles Host Pollution.
* **Smart Control Plane**: Derrick acts as a middleware orchestration layer deciding *when* and *if* lower-level orchestrators (Nix/Docker) should execute depending on contract validations.
* **State Contract**: The `derrick.yaml` file, which is the undeniable truth of what constitutes a "valid" local environment.
