#!/usr/bin/env bash
# Commit-message gate. Install: ln -sf ../../.claude/hooks/commit-msg.sh .git/hooks/commit-msg
# Enforces the conventional type prefix (AGENTS.md) and nudges for a Slice:/Refs:
# trailer on feature/fix commits so the trace chain (ADR-0014) stays intact.
set -euo pipefail
msg_file="$1"
subject="$(head -1 "$msg_file")"

# Allow merge/revert/fixup commits through untouched.
case "$subject" in
  Merge*|Revert*|"fixup!"*|"squash!"*) exit 0 ;;
esac

if ! printf '%s' "$subject" | grep -qE '^(feat|fix|docs|chore|ci|refactor|test|perf)(\([^)]+\))?: .+'; then
  echo "✗ commit subject must be conventional: <type>(scope?): <desc>"
  echo "  types: feat fix docs chore ci refactor test perf — got: $subject"
  exit 1
fi

# feat/fix should carry a trace trailer (advisory — warn, don't block).
if printf '%s' "$subject" | grep -qE '^(feat|fix)'; then
  grep -qE '^(Slice|Refs):' "$msg_file" || \
    echo "⚠ tip: add a 'Refs: FR-…' or 'Slice: <cap>' trailer to keep traceability (ADR-0014)."
fi
exit 0
