#!/usr/bin/env bash
# Deterministic pre-commit guard. Install:  ln -sf ../../.claude/hooks/pre-commit.sh .git/hooks/pre-commit
# Fast gates only (full pyramid runs in CI) — fail closed, never auto-bypass.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

echo "• gofmt"
unformatted="$(cd backend && gofmt -l .)"
[ -z "$unformatted" ] || { echo "gofmt needed:"; echo "$unformatted"; exit 1; }

echo "• go vet"
(cd backend && go vet ./...)

echo "• traceability (fresh + non-regressing)"
node scripts/gen-traceability.mjs --check

echo "• secret scan (staged)"
if command -v gitleaks >/dev/null 2>&1; then
  gitleaks protect --staged --redact --no-banner
else
  echo "  (gitleaks not installed — CI still scans full history)"
fi

echo "✓ pre-commit gates passed"
