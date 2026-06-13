package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// newTestStore connects to DATABASE_URL or skips. Intended to run against a
// throwaway/test database; it applies migrations and cleans the M1 tables.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping DB integration test")
	}
	ctx := context.Background()
	store, err := New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(store.Close)

	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if _, err := store.pool.Exec(ctx, `DELETE FROM matches; DELETE FROM teams;`); err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	return store
}

func TestUpsertAndList(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	kick := time.Date(2026, 6, 11, 19, 0, 0, 0, time.UTC)
	mn := 1

	err := store.WithTx(ctx, func(tx pgx.Tx) error {
		if _, err := store.UpsertTeam(ctx, tx, UpsertTeam{FifaID: "43911", Name: "Mexico", Code: "MEX", FlagURL: "https://flag/MEX", GroupLabel: "Group A"}); err != nil {
			return err
		}
		if _, err := store.UpsertTeam(ctx, tx, UpsertTeam{FifaID: "43883", Name: "South Africa", Code: "RSA", FlagURL: "https://flag/RSA", GroupLabel: "Group A"}); err != nil {
			return err
		}
		hs, as := 2, 0
		return store.UpsertMatch(ctx, tx, UpsertMatch{
			FifaID: "400021443", Stage: "group", GroupLabel: "Group A", MatchNumber: &mn,
			HomeFifaID: "43911", AwayFifaID: "43883", KickoffAt: &kick, Status: "finished",
			HomeScore: &hs, AwayScore: &as,
			VenueStadium: "Estadio Azteca", VenueCity: "Mexico City", VenueCountry: "MEX",
		})
	})
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}

	// Idempotent re-run with a changed score should update in place.
	err = store.WithTx(ctx, func(tx pgx.Tx) error {
		hs, as := 3, 1
		return store.UpsertMatch(ctx, tx, UpsertMatch{
			FifaID: "400021443", Stage: "group", GroupLabel: "Group A", MatchNumber: &mn,
			HomeFifaID: "43911", AwayFifaID: "43883", KickoffAt: &kick, Status: "finished",
			HomeScore: &hs, AwayScore: &as,
		})
	})
	if err != nil {
		t.Fatalf("re-upsert: %v", err)
	}

	matches, err := store.ListMatches(ctx)
	if err != nil {
		t.Fatalf("list matches: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	m := matches[0]
	if m.HomeScore == nil || *m.HomeScore != 3 {
		t.Errorf("score not updated idempotently: %v", m.HomeScore)
	}
	if m.Home == nil || m.Home.Name != "Mexico" {
		t.Errorf("home join failed: %+v", m.Home)
	}

	// Manual rows must not be overwritten by a subsequent fifa upsert.
	if _, err := store.pool.Exec(ctx, `UPDATE matches SET result_source='manual', home_score=9 WHERE fifa_id='400021443'`); err != nil {
		t.Fatalf("set manual: %v", err)
	}
	err = store.WithTx(ctx, func(tx pgx.Tx) error {
		hs, as := 0, 0
		return store.UpsertMatch(ctx, tx, UpsertMatch{
			FifaID: "400021443", Stage: "group", HomeFifaID: "43911", AwayFifaID: "43883",
			Status: "finished", HomeScore: &hs, AwayScore: &as,
		})
	})
	if err != nil {
		t.Fatalf("upsert over manual: %v", err)
	}
	matches, _ = store.ListMatches(ctx)
	if matches[0].HomeScore == nil || *matches[0].HomeScore != 9 {
		t.Errorf("manual row was overwritten: %v", matches[0].HomeScore)
	}

	teams, err := store.ListTeams(ctx)
	if err != nil {
		t.Fatalf("list teams: %v", err)
	}
	if len(teams) != 2 {
		t.Errorf("expected 2 teams, got %d", len(teams))
	}
}
