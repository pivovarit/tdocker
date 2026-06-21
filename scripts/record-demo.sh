#!/usr/bin/env bash
#
# Records the tdocker README demo GIF, fully scripted and reproducible.
# Stages a Compose project ("shop": web/api/db) plus one standalone container,
# drives tdocker through the headline keystrokes inside a tmux pane that
# asciinema records, then renders to a GIF with agg. Re-run on every release
# to keep docs/tdocker.gif current.
#
# Requires: docker (running) + compose, asciinema, agg, tmux, go.
#   brew install asciinema agg tmux
#
# Usage: ./scripts/record-demo.sh [output.gif]   # default: docs/tdocker.gif
set -euo pipefail
cd "$(dirname "$0")/.."

OUT="${1:-docs/tdocker.gif}"
CAST="$(mktemp -t tdocker-demo).cast"
BIN="$(mktemp -t tdocker-bin)"
WORK="$(mktemp -d -t tdocker-compose)"
COMPOSE="$WORK/docker-compose.yml"

cleanup() {
  tmux kill-session -t tdrec 2>/dev/null || true
  docker compose -f "$COMPOSE" down >/dev/null 2>&1 || true
  docker rm -f cache >/dev/null 2>&1 || true
  rm -rf "$WORK" "$BIN"
}
trap cleanup EXIT

echo "==> building tdocker"
go build -o "$BIN" .

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
# generate a few access-log lines so the logs view has content
for p in / /health /api/users /api/orders /favicon.ico /metrics; do
  curl -s "localhost:8080$p" >/dev/null 2>&1 || true
done

echo "==> recording (asciinema driven via tmux)"
tmux kill-session -t tdrec 2>/dev/null || true
tmux new-session -d -s tdrec -x 120 -y 22        # 120 cols so expanded detail values fit
K()   { tmux send-keys -t tdrec -l "$1"; }       # literal key(s)
ENT() { tmux send-keys -t tdrec Enter; }
tmux send-keys -t tdrec -l 'export DOCKER_CLI_HINTS=false'; ENT   # suppress Docker's "What's next" promo
tmux send-keys -t tdrec -l 'clear'; ENT
tmux send-keys -t tdrec -l "asciinema rec --overwrite -c \"$BIN\" \"$CAST\""; ENT
sleep 4.2                                  # startup + dwell on the grouped list
tmux send-keys -t tdrec Left;  sleep 1.5   # collapse Compose group -> "▸ shop (3 running)"
tmux send-keys -t tdrec Right; sleep 1.2   # expand it again
K j; sleep 0.6                             # -> shop/db
tmux send-keys -t tdrec Right; sleep 3.8   # expand inline details (async inspect load, then port bindings + network)
tmux send-keys -t tdrec Left;  sleep 0.8   # collapse details
K j; sleep 0.8                             # -> shop/web (nginx)
K l; sleep 2.4                             # logs (access-log lines)
K q; sleep 0.7                             # back
K e; sleep 1.9                             # exec a shell into the container
K 'ls /'; ENT; sleep 1.4                   # run a command in the shell
K 'exit'; ENT; sleep 1.1                   # leave shell -> back to tdocker
K c; sleep 1.6                             # copy container ID (toast)
K q; sleep 1.5                             # quit -> asciinema writes the cast
tmux kill-session -t tdrec 2>/dev/null || true

echo "==> rendering $OUT"
agg "$CAST" "$OUT" --idle-time-limit 1.2 --font-size 16
cp "$CAST" docs/tdocker.cast

echo "done: $OUT"
