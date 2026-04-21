---
title: Meilisearch
---

# Meilisearch

**The problem:** Meilisearch is a single Rust binary — easy to run. The friction is everything around it: keeping the master key in sync across team members, managing index configuration, and loading seed data so every developer searches the same dataset.

**The Derrick solution:** `provider: docker` runs Meilisearch with a Compose file. Env vars manage the master key. A `setup` hook configures indexes and loads seed data once — subsequent starts just boot the container.

```yaml
name: meilisearch-dev
version: 1.8.0
provider: docker

docker:
  compose: docker-compose.yml

env:
  MEILI_MASTER_KEY:
    description: "Master API key — generate with: openssl rand -hex 16"
    required: true
    validation: "[ ${#MEILI_MASTER_KEY} -ge 16 ] || (echo 'MEILI_MASTER_KEY must be ≥ 16 chars' && exit 1)"
  MEILI_ENV:
    description: "Meilisearch environment (development or production)"
    default: "development"
  MEILI_HTTP_ADDR:
    default: "localhost:7700"

validations:
  - name: "Port 7700 free"
    command: "! lsof -i :7700"
    auto_fix: "kill -9 $(lsof -t -i:7700) 2>/dev/null || true"

hooks:
  after_start:
    - run: |
        until curl -sf http://$MEILI_HTTP_ADDR/health > /dev/null; do sleep 1; done
        curl -sf -X POST "http://$MEILI_HTTP_ADDR/indexes" \
          -H "Authorization: Bearer $MEILI_MASTER_KEY" \
          -H "Content-Type: application/json" \
          -d '{"uid":"products","primaryKey":"id"}' > /dev/null || true
      when: first-setup
    - run: |
        curl -sf -X POST "http://$MEILI_HTTP_ADDR/indexes/products/documents" \
          -H "Authorization: Bearer $MEILI_MASTER_KEY" \
          -H "Content-Type: application/json" \
          --data-binary @seed/products.json > /dev/null
      when: first-setup
    - "echo 'Meilisearch: http://$MEILI_HTTP_ADDR (key: $MEILI_MASTER_KEY)'"

flags:
  reseed:
    description: "Re-load seed data into all indexes"
```

## What's happening

| Stage | Command | Why |
| :--- | :--- | :--- |
| `after_start` (first time) | wait + create index | Polls until healthy, then creates the `products` index with the correct primary key. |
| `after_start` (first time) | load documents | Bulk-loads `seed/products.json` so every developer has the same searchable dataset. |

## Seed data format

`seed/products.json` should be a JSON array of objects, each with an `id` field:

```json
[
  { "id": 1, "name": "Wireless Keyboard", "category": "peripherals", "price": 79.99 },
  { "id": 2, "name": "USB-C Hub", "category": "peripherals", "price": 49.99 }
]
```

## Reseeding

```bash
derrick start --flag reseed
```

```yaml
hooks:
  after_start:
    - run: |
        curl -sf -X DELETE "http://$MEILI_HTTP_ADDR/indexes/products/documents" \
          -H "Authorization: Bearer $MEILI_MASTER_KEY" > /dev/null
        curl -sf -X POST "http://$MEILI_HTTP_ADDR/indexes/products/documents" \
          -H "Authorization: Bearer $MEILI_MASTER_KEY" \
          -H "Content-Type: application/json" \
          --data-binary @seed/products.json > /dev/null
      when: flag:reseed
```

## Using as a dependency

If your main project needs Meilisearch as a service, use `requires`:

```yaml
# In your main project's derrick.yaml
requires:
  - name: meilisearch-dev
    connect: false   # no shared Docker network needed — HTTP is enough
```
