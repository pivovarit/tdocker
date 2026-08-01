#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

OUT="${1:-docs/tdocker.gif}"
TAPE="scripts/demo.tape"
BINDIR="$(mktemp -d -t tdocker-bin)"
WORK="$(mktemp -d -t tdocker-compose)"
COMPOSE="$WORK/docker-compose.yml"

cleanup() {
  docker compose -f "$COMPOSE" down >/dev/null 2>&1 || true
  docker rm -f cache >/dev/null 2>&1 || true
  rm -rf "$WORK" "$BINDIR"
}
trap cleanup EXIT

echo "==> building tdocker onto PATH"
go build -o "$BINDIR/tdocker" .
export PATH="$BINDIR:$PATH"
export DOCKER_CLI_HINTS=false             # suppress Docker's "What's next" promo

echo "==> staging Compose project 'shop' + standalone 'cache'"
cat > "$COMPOSE" <<'YAML'
name: shop
services:
  web:
    image: nginx:alpine
    ports: ["8080:80"]
  api:
    image: redis:alpine
  db:
    image: postgres:alpine
    environment:
      POSTGRES_USER: app
      POSTGRES_PASSWORD: demo
      POSTGRES_DB: appdb
    ports: ["5432:5432"]
YAML
docker compose -f "$COMPOSE" down >/dev/null 2>&1 || true
docker rm -f cache >/dev/null 2>&1 || true
docker compose -f "$COMPOSE" up -d >/dev/null
docker run -d --name cache -p 6390:6379 redis:alpine >/dev/null
sleep 2
for p in / /health /api/users /api/orders /favicon.ico /metrics; do
  curl -s "localhost:8080$p" >/dev/null 2>&1 || true
done

echo "==> recording with VHS ($TAPE)"
vhs "$TAPE" -o "$OUT"

echo "done: $OUT"
