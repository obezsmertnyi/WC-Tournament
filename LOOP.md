# LOOP — the tight cycle

```
        ┌──────────────────────────────────────────────────────────┐
        │                                                          │
        ▼                                                          │
   ┌─────────┐   ┌───────────┐   ┌────────┐   ┌────────┐   ┌──────────┐
   │  SPEC   │──▶│ IMPLEMENT │──▶│ TRACE  │──▶│ VERIFY │──▶│  REVIEW  │
   │ FR + GWT│   │ smallest  │   │ @trace │   │ gates  │   │ maker ≠  │
   │  + ADR  │   │  change   │   │  FR-id │   │  green │   │ checker  │
   └─────────┘   └───────────┘   └────────┘   └────────┘   └────┬─────┘
        ▲                                          │ fail       │ findings
        │                                          └────────────┤
        │                                                       ▼
        │                                                  ┌─────────┐
        └──────────────── handoff (current-state) ─────────│ COMMIT  │
                                                           │ + trailers
                                                           └─────────┘
```

Repeat per capability slice. Never advance past a red gate; never mark a
requirement done until a `@trace`'d test proves it and a separate checker has
looked. Full detail in `WORKFLOW.md`; live status in `CHECKLIST.md`.

Commands (`.claude/commands/`): `/verify` (run the gates), `/trace` (regenerate +
check the traceability matrix), `/review` (dispatch the reviewer sub-agents),
`/new-capability` (scaffold a spec + FR ids).
