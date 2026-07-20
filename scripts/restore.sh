#!/usr/bin/env bash
# Bring a WC-Tournament backup (from backup.sh) up on a fresh host — one command.
# Prereqs on the target: docker, docker compose, git, and (for the AI overlay)
# sudo to place /etc/gemini/cred.json.
#
#   ./restore.sh wc2026-backup-<ts>.tar.gz [target-dir]
#
# It clones the repo from the bundle, restores secrets + the database, then builds
# and starts the stack (with the Gemini AI overlay if a cred was in the backup).
set -euo pipefail

ARCHIVE="${1:?usage: restore.sh <backup.tar.gz> [target-dir]}"
DIR="${2:-WC-Tournament}"
ARCHIVE="$(cd "$(dirname "$ARCHIVE")" && pwd)/$(basename "$ARCHIVE")"
STAGE="$(mktemp -d)"
trap 'rm -rf "$STAGE"' EXIT

echo "==> [1/5] unpack $ARCHIVE"
tar -C "$STAGE" -xzf "$ARCHIVE"
cat "$STAGE/MANIFEST.txt" 2>/dev/null || true

echo "==> [2/5] clone repo from bundle -> $DIR"
[ -e "$DIR" ] && { echo "target '$DIR' already exists — aborting"; exit 1; }
git clone -q "$STAGE/repo.bundle" "$DIR"
cd "$DIR"

echo "==> [3/5] restore secrets"
cp "$STAGE/env.app" .env
OVERLAY=()
if [ -f "$STAGE/gemini-cred.json" ]; then
  sudo mkdir -p /etc/gemini && sudo cp "$STAGE/gemini-cred.json" /etc/gemini/cred.json
  OVERLAY=(-f docker-compose.yml -f docker-compose.gemini.yml)
  echo "    AI overlay enabled (NOTE: Vertex WIF is host-bound — may need re-provisioning)"
fi

echo "==> [4/5] start db + restore dump"
docker compose up -d db
until docker compose exec -T db pg_isready -U wc -d wc2026 >/dev/null 2>&1; do sleep 1; done
docker compose exec -T db pg_restore -U wc -d wc2026 --clean --if-exists --no-owner < "$STAGE/db.dump"

echo "==> [5/5] build + start the stack"
docker compose "${OVERLAY[@]}" up -d --build

echo "==> DONE. Frontend on the mapped web port, backend on :8081 (see docker-compose.yml)."
echo "    Verify: docker compose ps ; curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8080/"
