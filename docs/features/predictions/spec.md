# Spec — Predictions (CAP-01)

Owns: FR-001, FR-002, FR-003. Related: [ADR-0007](../../adr/0007-knockout-bracket-from-fifa.md).

### FR-001 — submit / edit before kickoff
- **Given** a signed-in `rw` player and a fixture that has not kicked off
- **When** they submit a score (home, away, each 0–30)
- **Then** the prediction is saved and can be edited until kickoff.

### FR-002 — kickoff lock
- **Given** a fixture that has kicked off
- **When** a non-admin tries to create or change a prediction for it
- **Then** the write is rejected (locked); admins may still set/correct (audited).

### FR-003 — knockout advancer required on a predicted draw
- **Given** a knockout fixture
- **When** the player predicts equal scores (a regulation draw)
- **Then** they must also pick which team advances before the prediction can be saved;
- **And** for a decisive predicted scoreline, no advancer pick is requested (winner implied).
