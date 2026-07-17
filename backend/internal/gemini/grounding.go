package gemini

import "context"

// Tools is the read-only window into the app's OWN tournament state that the
// assistant grounds itself in (ADR-0018). The real impl (api layer) reads storage;
// tests stub it. This is what makes the AI answer from live data instead of stale
// model knowledge — the WC 2026 pool is already underway in-app.
type Tools interface {
	// TournamentOverview is the PRIMARY tool: one call returns a compact snapshot
	// of the whole championship for a quick summary (FR-100).
	TournamentOverview(ctx context.Context) (OverviewFact, error)
	RecentResults(ctx context.Context, limit int) ([]MatchFact, error)
	TeamMatches(ctx context.Context, team string) ([]MatchFact, error)
	GroupStandings(ctx context.Context, group string) ([]StandingFact, error)
	Leaderboard(ctx context.Context, limit int) ([]LeaderFact, error)
	TopScorers(ctx context.Context, limit int) ([]ScorerFact, error)
}

// ScorerFact is one player's goal tally (golden-boot race).
type ScorerFact struct {
	Name  string `json:"name"`
	Team  string `json:"team,omitempty"`
	Goals int    `json:"goals"`
}

// OverviewFact is a one-shot snapshot of the tournament for a quick summary.
type OverviewFact struct {
	CurrentStage  string      `json:"current_stage"`  // furthest stage with a played match
	MatchesPlayed int         `json:"matches_played"` // finished
	MatchesTotal  int         `json:"matches_total"`
	Recent        []MatchFact `json:"recent_results"` // last few finished
	Upcoming      []MatchFact `json:"upcoming"`       // next few scheduled
	PoolLeader    string      `json:"pool_leader,omitempty"`
	PoolLeaderPts int         `json:"pool_leader_points,omitempty"`
}

// MatchFact is one fixture/result in a model-friendly shape.
type MatchFact struct {
	Stage   string `json:"stage"`
	Group   string `json:"group,omitempty"`
	Kickoff string `json:"kickoff,omitempty"`
	Status  string `json:"status"`
	Home    string `json:"home"`
	Away    string `json:"away"`
	Score   string `json:"score,omitempty"` // "2:1" (home:away) once played
	// Winner and Advanced are set ONLY for a finished match, computed from the
	// authoritative scores/advancer — so the model never has to (and must not)
	// infer the result from the raw score, which it gets wrong. Winner is the
	// team that won by score, or "draw" when level. Advanced is the knockout
	// team that went through (from extra time / penalties); it can differ from
	// the scoreline when a regulation draw was settled beyond 90 minutes.
	Winner   string `json:"winner,omitempty"`
	Advanced string `json:"advanced,omitempty"`
	// Resolution/ResolutionScore explain HOW a knockout level after 90' was
	// decided, so the model states it exactly instead of guessing. Resolution is
	// "extra_time" or "penalties" (absent = decided in normal time);
	// ResolutionScore is the aet score (extra time) or the shootout score.
	Resolution      string `json:"resolution,omitempty"`
	ResolutionScore string `json:"resolutionScore,omitempty"`
}

// StandingFact is one row of a group table.
type StandingFact struct {
	Rank   int    `json:"rank"`
	Team   string `json:"team"`
	Played int    `json:"played"`
	Win    int    `json:"win"`
	Draw   int    `json:"draw"`
	Loss   int    `json:"loss"`
	GF     int    `json:"gf"`
	GA     int    `json:"ga"`
	GD     int    `json:"gd"`
	Points int    `json:"points"`
}

// LeaderFact is one row of the prediction-pool leaderboard.
type LeaderFact struct {
	Rank     int    `json:"rank"`
	Nickname string `json:"nickname"`
	Points   int    `json:"points"`
	Exact    int    `json:"exact"`
}

// tool names shared by the declarations (vertex.go) and the dispatcher.
const (
	toolOverview       = "tournament_overview"
	toolRecentResults  = "recent_results"
	toolTeamMatches    = "team_matches"
	toolGroupStandings = "group_standings"
	toolLeaderboard    = "pool_leaderboard"
	toolTopScorers     = "top_scorers"
)

// dispatchTool runs one model-requested tool call against Tools and returns a
// JSON-object result for the function response. It never errors out to the caller:
// a failure becomes {"error": ...} so the model can recover gracefully.
func dispatchTool(ctx context.Context, t Tools, name string, args map[string]any) map[string]any {
	if t == nil {
		return map[string]any{"error": "no data source"}
	}
	wrap := func(v any, err error) map[string]any {
		if err != nil {
			return map[string]any{"error": "data unavailable"}
		}
		return map[string]any{"data": v}
	}
	switch name {
	case toolOverview:
		return wrap(t.TournamentOverview(ctx))
	case toolRecentResults:
		return wrap(t.RecentResults(ctx, argInt(args, "limit", 8)))
	case toolTeamMatches:
		return wrap(t.TeamMatches(ctx, argStr(args, "team")))
	case toolGroupStandings:
		return wrap(t.GroupStandings(ctx, argStr(args, "group")))
	case toolLeaderboard:
		return wrap(t.Leaderboard(ctx, argInt(args, "limit", 10)))
	case toolTopScorers:
		return wrap(t.TopScorers(ctx, argInt(args, "limit", 10)))
	default:
		return map[string]any{"error": "unknown tool"}
	}
}

func argStr(args map[string]any, key string) string {
	if s, ok := args[key].(string); ok {
		return s
	}
	return ""
}

// argInt reads a numeric arg (JSON numbers arrive as float64), clamped to [1,50].
func argInt(args map[string]any, key string, def int) int {
	n := def
	switch v := args[key].(type) {
	case float64:
		n = int(v)
	case int:
		n = v
	}
	if n < 1 {
		n = def
	}
	if n > 50 {
		n = 50
	}
	return n
}
