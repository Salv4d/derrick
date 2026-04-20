---
title: Ghost CMS
---

# Ghost CMS Local Orchestration

**Ghost CMS** represents the classic friction of "Node Version Drift" (Easy/Medium). Ghost specifically demands an exact LTS cycle of **Node.js**, while developers constantly have conflicting versions installed locally via tools like `nvm` or global repositories.

### The Derrick Solution

Zero host OS pollution. By passing `nodejs_18` directly to Derrick, Ghost is boxed beautifully within the exact binaries it wants, entirely isolating developers from managing their OS's node. We also map strict `env` overrides for local mail setups securely.

### The `derrick.yaml` Implementation

```yaml
---
name: "ghost-blog"
version: "5.0.0"

dependencies:
  nix_packages:
    - "nodejs_18" # Strict LTS bounds required by Ghost
    - "ghost-cli"
  docker_compose: "docker-compose.yml" # Assumes a MySQL db template

env:
  url:
    description: "Localhost bound URL"
    default: "http://localhost:2368"
  database__client:
    default: "mysql"
  database__connection__password:
    description: "Root dev password"
    default: "root"

validations:
  - name: "Port Check"
    command: "! lsof -i :2368"
    auto_fix: "kill -9 $(lsof -t -i:2368)"

hooks:
  setup:
    - run: "ghost install local"
      when: first-setup
  after_start:
    - "ghost start"
```
