package api

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/gemini"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/standings"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
)

// aiDataReader is the read-only storage surface the AI grounding tools use
// (satisfied by *storage.Store). Kept as an interface for testability.
type aiDataReader interface {
	ListMatches(ctx context.Context) ([]storage.Match, error)
	ListTeams(ctx context.Context) ([]storage.Team, error)
	ListFinishedGroupMatches(ctx context.Context) ([]storage.Match, error)
	Leaderboard(ctx context.Context) ([]storage.LeaderboardRow, error)
	ListTopScorers(ctx context.Context, limit int) ([]storage.ScorerRow, error)
}

// aiTools implements gemini.Tools over storage — this is what grounds the AI in the
// app's OWN live tournament state instead of the model's stale knowledge (ADR-0018).
type aiTools struct{ r aiDataReader }

// NewAITools builds the grounding tools over storage.
func NewAITools(r aiDataReader) gemini.Tools { return &aiTools{r: r} }

var stageOrder = map[string]int{"group": 1, "r32": 2, "r16": 3, "qf": 4, "sf": 5, "third": 6, "final": 7}

func isFinished(m storage.Match) bool {
	return m.Status == "finished" && m.HomeScore != nil && m.AwayScore != nil
}

// kickoffBefore orders by kickoff; a nil kickoff sorts last.
func kickoffBefore(a, b storage.Match) bool {
	switch {
	case a.KickoffAt == nil:
		return false
	case b.KickoffAt == nil:
		return true
	default:
		return a.KickoffAt.Before(*b.KickoffAt)
	}
}

func matchFact(m storage.Match) gemini.MatchFact {
	f := gemini.MatchFact{Stage: m.Stage, Group: m.GroupLabel, Status: m.Status}
	if m.KickoffAt != nil {
		f.Kickoff = m.KickoffAt.Format("2006-01-02 15:04")
	}
	if m.Home != nil {
		f.Home = m.Home.Name
	}
	if m.Away != nil {
		f.Away = m.Away.Name
	}
	if m.HomeScore != nil && m.AwayScore != nil {
		f.Score = fmt.Sprintf("%d:%d", *m.HomeScore, *m.AwayScore)
	}
	return f
}

func (t *aiTools) TournamentOverview(ctx context.Context) (gemini.OverviewFact, error) {
	ms, err := t.r.ListMatches(ctx)
	if err != nil {
		return gemini.OverviewFact{}, err
	}
	of := gemini.OverviewFact{MatchesTotal: len(ms)}
	var finished, upcoming []storage.Match
	maxStage := 0
	for _, m := range ms {
		if isFinished(m) {
			finished = append(finished, m)
			if o := stageOrder[m.Stage]; o > maxStage {
				maxStage, of.CurrentStage = o, m.Stage
			}
		} else {
			upcoming = append(upcoming, m)
		}
	}
	of.MatchesPlayed = len(finished)
	sort.Slice(finished, func(i, j int) bool { return kickoffBefore(finished[j], finished[i]) }) // most recent first
	for i, m := range finished {
		if i >= 5 {
			break
		}
		of.Recent = append(of.Recent, matchFact(m))
	}
	sort.Slice(upcoming, func(i, j int) bool { return kickoffBefore(upcoming[i], upcoming[j]) }) // soonest first
	for i, m := range upcoming {
		if i >= 3 {
			break
		}
		of.Upcoming = append(of.Upcoming, matchFact(m))
	}
	if lb, err := t.r.Leaderboard(ctx); err == nil && len(lb) > 0 {
		of.PoolLeader, of.PoolLeaderPts = lb[0].Nickname, lb[0].Points
	}
	return of, nil
}

func (t *aiTools) RecentResults(ctx context.Context, limit int) ([]gemini.MatchFact, error) {
	ms, err := t.r.ListMatches(ctx)
	if err != nil {
		return nil, err
	}
	fin := make([]storage.Match, 0, len(ms))
	for _, m := range ms {
		if isFinished(m) {
			fin = append(fin, m)
		}
	}
	sort.Slice(fin, func(i, j int) bool { return kickoffBefore(fin[j], fin[i]) })
	out := make([]gemini.MatchFact, 0, limit)
	for i, m := range fin {
		if i >= limit {
			break
		}
		out = append(out, matchFact(m))
	}
	return out, nil
}

func (t *aiTools) TeamMatches(ctx context.Context, team string) ([]gemini.MatchFact, error) {
	q := strings.ToLower(strings.TrimSpace(team))
	if q == "" {
		return nil, nil
	}
	ms, err := t.r.ListMatches(ctx)
	if err != nil {
		return nil, err
	}
	hit := func(tm *storage.Team) bool {
		return tm != nil && (strings.EqualFold(tm.Code, q) || strings.Contains(strings.ToLower(tm.Name), q))
	}
	var out []gemini.MatchFact
	for _, m := range ms {
		if hit(m.Home) || hit(m.Away) {
			out = append(out, matchFact(m))
		}
	}
	return out, nil
}

func (t *aiTools) GroupStandings(ctx context.Context, group string) ([]gemini.StandingFact, error) {
	g := strings.TrimSpace(strings.TrimPrefix(strings.ToUpper(strings.TrimSpace(group)), "GROUP "))
	teamRows, err := t.r.ListTeams(ctx)
	if err != nil {
		return nil, err
	}
	matchRows, err := t.r.ListFinishedGroupMatches(ctx)
	if err != nil {
		return nil, err
	}
	teams := make([]standings.Team, 0, len(teamRows))
	for _, tr := range teamRows {
		teams = append(teams, standings.Team{ID: tr.ID, Name: tr.Name, Code: tr.Code, FlagURL: tr.FlagURL, GroupLabel: tr.GroupLabel})
	}
	matches := make([]standings.Match, 0, len(matchRows))
	for _, m := range matchRows {
		var homeID, awayID *int64
		if m.Home != nil {
			homeID = &m.Home.ID
		}
		if m.Away != nil {
			awayID = &m.Away.ID
		}
		matches = append(matches, standings.Match{HomeTeamID: homeID, AwayTeamID: awayID, HomeScore: m.HomeScore, AwayScore: m.AwayScore})
	}
	for _, gs := range standings.ComputeStandings(teams, matches) {
		if !strings.EqualFold(gs.Group, g) {
			continue
		}
		out := make([]gemini.StandingFact, 0, len(gs.Rows))
		for _, r := range gs.Rows {
			out = append(out, gemini.StandingFact{
				Rank: r.Rank, Team: r.Name, Played: r.Played, Win: r.Win, Draw: r.Draw,
				Loss: r.Loss, GF: r.GF, GA: r.GA, GD: r.GD, Points: r.Points,
			})
		}
		return out, nil
	}
	return nil, nil
}

func (t *aiTools) TopScorers(ctx context.Context, limit int) ([]gemini.ScorerFact, error) {
	rows, err := t.r.ListTopScorers(ctx, limit)
	if err != nil {
		return nil, err
	}
	out := make([]gemini.ScorerFact, 0, len(rows))
	for _, r := range rows {
		out = append(out, gemini.ScorerFact{Name: r.Name, Team: r.TeamCode, Goals: r.Goals})
	}
	return out, nil
}

func (t *aiTools) Leaderboard(ctx context.Context, limit int) ([]gemini.LeaderFact, error) {
	rows, err := t.r.Leaderboard(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]gemini.LeaderFact, 0, limit)
	for i, r := range rows {
		if i >= limit {
			break
		}
		out = append(out, gemini.LeaderFact{Rank: i + 1, Nickname: r.Nickname, Points: r.Points, Exact: r.ExactCount})
	}
	return out, nil
}
