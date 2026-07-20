#!/usr/bin/env bash
# Full, portable backup of a running WC-Tournament deployment into ONE tarball you
# can carry to any host and bring up with restore.sh. Captures everything that is
# NOT reproducible from a plain `git clone`:
#   - db.dump          the Postgres database (pg_dump -Fc: all tournament data)
#   - env.app          the app .env (JWT/OAuth/Telegram/admin secrets) — SENSITIVE
#   - gemini-cred.json  the Vertex WIF credential, if mounted — SENSITIVE
#   - repo.bundle      a git bundle of the repo (so restore needs no GitHub)
#   - MANIFEST.txt     what's inside + the exact commit
#
# Run it on the current server from the repo checkout (e.g. /srv/WC-Tournament):
#   ./scripts/backup.sh            # -> ./wc2026-backup-<ts>.tar.gz
#   ./scripts/backup.sh /srv/out   # -> /srv/out/wc2026-backup-<ts>.tar.gz
set -euo pipefail

cd "$(cd "$(dirname "$0")/.." && pwd)"   # repo root
OUT_DIR="${1:-$PWD}"
TS="$(date +%Y%m%d-%H%M%S)"
NAME="wc2026-backup-$TS"
STAGE="$(mktemp -d)"
trap 'rm -rf "$STAGE"' EXIT

echo "==> [1/5] database dump (pg_dump -Fc)"
docker compose exec -T db pg_dump -U wc -Fc wc2026 > "$STAGE/db.dump"

echo "==> [2/5] app secrets (.env)"
cp .env "$STAGE/env.app"

echo "==> [3/5] Vertex WIF credential (if any)"
if [ -f /etc/gemini/cred.json ]; then
  cp /etc/gemini/cred.json "$STAGE/gemini-cred.json"
else
  echo "    (no /etc/gemini/cred.json — AI overlay not configured on this host)"
fi

echo "==> [4/5] self-contained repo (git bundle --all)"
git bundle create "$STAGE/repo.bundle" --all >/dev/null

echo "==> [5/5] manifest"
cat > "$STAGE/MANIFEST.txt" <<EOF
WC-Tournament backup
  taken:    $TS
  host:     $(hostname 2>/dev/null || echo '?')
  commit:   $(git rev-parse HEAD)
  branch:   $(git rev-parse --abbrev-ref HEAD)
  db:       pg_dump -Fc (custom format) of database 'wc2026'
Restore on a fresh host (needs docker + docker compose + git):
  ./scripts/restore.sh $NAME.tar.gz
Contents: db.dump, env.app, gemini-cred.json (SENSITIVE), repo.bundle, MANIFEST.txt
NOTE: env.app and gemini-cred.json hold SECRETS — keep this tarball private.
NOTE: the Vertex WIF cred is host/identity-bound; the AI assistant may need WIF
re-provisioning on a new host. The pool itself runs fully without it.
EOF

mkdir -p "$OUT_DIR"
tar -C "$STAGE" -czf "$OUT_DIR/$NAME.tar.gz" .
echo "==> DONE: $OUT_DIR/$NAME.tar.gz ($(du -h "$OUT_DIR/$NAME.tar.gz" | cut -f1))"
