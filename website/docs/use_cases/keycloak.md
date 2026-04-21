---
title: Keycloak
---

# Keycloak

**The problem:** Keycloak's dev mode is fine for one person. The moment you need a real realm config, client secrets, and a Postgres backend checked into version control, every team member ends up with a different realm state — and "it works on my machine" becomes the default answer.

**The Derrick solution:** `provider: docker` runs Keycloak and Postgres via Compose. Env vars control the admin credentials and DB connection. A `setup` hook imports a realm JSON on first setup, so every developer starts from the same baseline configuration.

```yaml
name: keycloak-dev
version: 24.0.0
provider: docker

docker:
  compose: docker-compose.yml

env:
  KEYCLOAK_ADMIN:
    description: "Admin username"
    default: "admin"
  KEYCLOAK_ADMIN_PASSWORD:
    description: "Admin password"
    default: "admin"
    required: true
  KC_DB_URL:
    description: "Postgres JDBC URL"
    default: "jdbc:postgresql://postgres:5432/keycloak"
  KC_DB_USERNAME:
    default: "keycloak"
  KC_DB_PASSWORD:
    required: true
    default: "keycloak"
  KC_HOSTNAME:
    description: "Public hostname Keycloak serves on"
    default: "localhost"

validations:
  - name: "Port 8080 free"
    command: "! lsof -i :8080"

hooks:
  after_start:
    - run: |
        until curl -sf http://localhost:8080/realms/master > /dev/null; do sleep 2; done
        /opt/keycloak/bin/kcadm.sh config credentials \
          --server http://localhost:8080 --realm master \
          --user $KEYCLOAK_ADMIN --password $KEYCLOAK_ADMIN_PASSWORD
        /opt/keycloak/bin/kcadm.sh create realms -f config/realm-export.json 2>/dev/null || true
      when: first-setup
    - "echo 'Keycloak admin: http://localhost:8080/admin  ($KEYCLOAK_ADMIN / $KEYCLOAK_ADMIN_PASSWORD)'"

flags:
  reimport-realm:
    description: "Re-import realm config from config/realm-export.json"
```

## What's happening

| Stage | Command | Why |
| :--- | :--- | :--- |
| `after_start` (first time) | wait + kcadm | Polls until Keycloak is healthy, then imports the realm JSON via the admin CLI. |

## Exporting realm config

After making changes in the admin UI, export the realm to keep it in version control:

```bash
docker compose exec keycloak \
  /opt/keycloak/bin/kc.sh export --realm my-realm --dir /tmp/export
docker compose cp keycloak:/tmp/export/my-realm-realm.json config/realm-export.json
```

Commit `config/realm-export.json` — the next `derrick start` picks it up automatically.

## Re-importing after changes

```bash
derrick start --flag reimport-realm
```

Add to `hooks.setup`:

```yaml
hooks:
  setup:
    - run: |
        /opt/keycloak/bin/kcadm.sh config credentials \
          --server http://localhost:8080 --realm master \
          --user $KEYCLOAK_ADMIN --password $KEYCLOAK_ADMIN_PASSWORD
        /opt/keycloak/bin/kcadm.sh update realms/my-realm -f config/realm-export.json
      when: flag:reimport-realm
```
