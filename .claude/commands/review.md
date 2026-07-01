---
description: Run the maker≠checker adversarial review with the reviewer sub-agents and persist findings.
---

Run an independent review of the current change. **You are the maker — do not
grade your own work.** Dispatch the reviewer sub-agents (a separate context):

- `scoring-correctness-reviewer` — logic vs specs (scoring / bracket / recap).
- `security-reviewer` — secrets, auth/access, MCP read-only, injection, CI.

Give each the changed files + the relevant spec/ADR. Collect their findings into
`docs/qa/review-findings.json` as:

```json
{ "reviewedAt": "<date>", "clean": true,
  "findings": [ { "id": "S1", "area": "…", "severity": "…", "issue": "…", "fix": "…", "status": "fixed|accepted|wontfix" } ] }
```

Address every critical/high finding at the root before merge; record medium/low
with a status. Set `clean: true` only when no unresolved critical/high remains.
CodeRabbit is the external second checker on the PR — reconcile its comments here too.
