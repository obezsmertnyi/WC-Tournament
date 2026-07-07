// Command server runs the WC-Tournament HTTP backend.
//
// Subcommands:
//
//	server               run the HTTP server (default)
//	server healthcheck   probe /healthz and exit 0/1 (container HEALTHCHECK)
//	server sync          run one FIFA calendar sync and exit
//	server announce      post unannounced finished results to Telegram and exit
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	_ "time/tzdata" // embed the tz database so Europe/Kyiv resolves in distroless

	"github.com/gin-gonic/gin"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/announce"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/api"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/auth"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/digest"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/gemini"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/notify"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/remind"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/resolve"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/results"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/scorers"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/scoring"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
	syncpkg "github.com/obezsmertnyi/WC-Tournament/backend/internal/sync"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/winners"
)

// version is the build version, injected at build time via
// -ldflags "-X main.version=<tag>" (defaults to "dev" for local builds).
var version = "dev"

// Kyiv is the display/scheduling timezone for the digest. Falls back to UTC if
// the zone can't be loaded.
var kyivLoc = mustLoadKyiv()

func mustLoadKyiv() *time.Location {
	loc, err := time.LoadLocation("Europe/Kyiv")
	if err != nil {
		return time.UTC
	}
	return loc
}

const (
	digestStateKey          = "last_digest_day"
	bonusesResolvedStateKey = "bonuses_resolved"
)

// fastTickInterval is how often the background loop syncs from FIFA and runs the
// reminder/announce/digest jobs. 5 min keeps results fresh while staying gentle
// on the FIFA calendar endpoint (one request per tick).
const fastTickInterval = 5 * time.Minute

const (
	defaultPort     = "8080"
	readTimeout     = 10 * time.Second
	writeTimeout    = 10 * time.Second
	idleTimeout     = 60 * time.Second
	shutdownTimeout = 15 * time.Second
	bootSyncTimeout = 2 * time.Minute
	syncCmdTimeout  = 3 * time.Minute
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// `server healthcheck` probes the local /healthz endpoint and exits 0/1.
	// Used as the container HEALTHCHECK since the distroless image has no shell/curl.
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(healthcheck())
	}

	// `server sync` runs a single FIFA sync and exits.
	if len(os.Args) > 1 && os.Args[1] == "sync" {
		if err := runSyncOnce(logger); err != nil {
			logger.Error("sync failed", slog.Any("error", err))
			os.Exit(1)
		}
		return
	}

	// `server recompute-scores` re-materializes points for all scored matches.
	if len(os.Args) > 1 && os.Args[1] == "recompute-scores" {
		if err := runRecomputeScores(logger); err != nil {
			logger.Error("recompute-scores failed", slog.Any("error", err))
			os.Exit(1)
		}
		return
	}

	// `server announce` posts unannounced finished results to Telegram and exits.
	if len(os.Args) > 1 && os.Args[1] == "announce" {
		if err := runAnnounce(logger); err != nil {
			logger.Error("announce failed", slog.Any("error", err))
			os.Exit(1)
		}
		return
	}

	// `server remind` posts pre-match "you haven't predicted" nudges and exits.
	if len(os.Args) > 1 && os.Args[1] == "remind" {
		if err := runRemind(logger); err != nil {
			logger.Error("remind failed", slog.Any("error", err))
			os.Exit(1)
		}
		return
	}

	// `server digest` posts the morning summary (results + table + today) and exits.
	if len(os.Args) > 1 && os.Args[1] == "digest" {
		if err := runDigest(logger); err != nil {
			logger.Error("digest failed", slog.Any("error", err))
			os.Exit(1)
		}
		return
	}

	// `server resolve-bonuses` awards champion/finalist/top-scorer after the final.
	if len(os.Args) > 1 && os.Args[1] == "resolve-bonuses" {
		if err := runResolveBonuses(logger); err != nil {
			logger.Error("resolve-bonuses failed", slog.Any("error", err))
			os.Exit(1)
		}
		return
	}

	if err := run(logger); err != nil {
		logger.Error("server exited with error", slog.Any("error", err))
		os.Exit(1)
	}
}

// healthcheck performs a self-probe against /healthz and returns a process exit code.
func healthcheck() int {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://127.0.0.1:" + port + "/healthz")
	if err != nil {
		return 1
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}

// openStore connects to Postgres and applies migrations. Returns (nil, nil)
// when DATABASE_URL is unset so the dev shell can still boot without a DB.
func openStore(ctx context.Context, logger *slog.Logger) (*storage.Store, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		logger.Warn("DATABASE_URL not set — skipping database (DB-backed routes will be unavailable)")
		return nil, nil
	}

	store, err := storage.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := store.Migrate(ctx); err != nil {
		store.Close()
		return nil, err
	}
	logger.Info("database connected and migrations applied")
	return store, nil
}

// runSyncOnce executes a single FIFA sync against the database and exits.
func runSyncOnce(logger *slog.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), syncCmdTimeout)
	defer cancel()

	store, err := openStore(ctx, logger)
	if err != nil {
		return err
	}
	if store == nil {
		return errors.New("DATABASE_URL is required for sync")
	}
	defer store.Close()

	syncer := syncpkg.New(results.NewFIFAClient(), store, logger)
	_, err = syncer.Run(ctx)
	if err != nil {
		return err
	}
	// Re-materialize points after the result refresh (idempotent).
	return scoring.NewRecomputer(store, scoring.DefaultRules()).RecomputeAll(ctx)
}

// runRecomputeScores re-materializes points for every match with a result.
func runRecomputeScores(logger *slog.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), syncCmdTimeout)
	defer cancel()

	store, err := openStore(ctx, logger)
	if err != nil {
		return err
	}
	if store == nil {
		return errors.New("DATABASE_URL is required for recompute-scores")
	}
	defer store.Close()

	if err := scoring.NewRecomputer(store, scoring.DefaultRules()).RecomputeAll(ctx); err != nil {
		return err
	}
	logger.Info("recompute-scores complete")
	return nil
}

// runAnnounce posts any finished-but-unannounced results to Telegram and exits.
func runAnnounce(logger *slog.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), syncCmdTimeout)
	defer cancel()

	store, err := openStore(ctx, logger)
	if err != nil {
		return err
	}
	if store == nil {
		return errors.New("DATABASE_URL is required for announce")
	}
	defer store.Close()

	tg, err := notify.TelegramFromEnv()
	if err != nil {
		return err
	}
	if !tg.Enabled() {
		logger.Warn("announce: TELEGRAM_BOT_TOKEN/TELEGRAM_CHAT_ID not set — marking results announced without posting")
	}
	sent, err := announce.New(store, tg, logger).Run(ctx)
	if err != nil {
		return err
	}
	logger.Info("announce complete", slog.Int("posted", sent))
	return nil
}

// runRemind sends pre-match reminders to players who haven't predicted, and exits.
func runRemind(logger *slog.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), syncCmdTimeout)
	defer cancel()

	store, err := openStore(ctx, logger)
	if err != nil {
		return err
	}
	if store == nil {
		return errors.New("DATABASE_URL is required for remind")
	}
	defer store.Close()

	tg, err := notify.TelegramFromEnv()
	if err != nil {
		return err
	}
	sent, err := remind.New(store, tg, logger).Run(ctx)
	if err != nil {
		return err
	}
	bonusSent, err := remind.NewBonus(store, tg, logger).Run(ctx, time.Now().UTC())
	if err != nil {
		return err
	}
	logger.Info("remind complete", slog.Int("sent", sent), slog.Bool("bonusReminder", bonusSent))
	return nil
}

// runDigest posts the morning digest once and exits (manual / cron use).
func runDigest(logger *slog.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), syncCmdTimeout)
	defer cancel()

	store, err := openStore(ctx, logger)
	if err != nil {
		return err
	}
	if store == nil {
		return errors.New("DATABASE_URL is required for digest")
	}
	defer store.Close()

	tg, err := notify.TelegramFromEnv()
	if err != nil {
		return err
	}
	sent, err := digest.New(store, tg, kyivLoc, logger).Run(ctx)
	if err != nil {
		return err
	}
	logger.Info("digest complete", slog.Bool("posted", sent))
	return nil
}

// runResolveBonuses awards champion/finalist/top-scorer once and exits.
func runResolveBonuses(logger *slog.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), syncCmdTimeout)
	defer cancel()

	store, err := openStore(ctx, logger)
	if err != nil {
		return err
	}
	if store == nil {
		return errors.New("DATABASE_URL is required for resolve-bonuses")
	}
	defer store.Close()

	done, err := resolve.New(store, results.NewFIFAClient(), logger).Run(ctx)
	if err != nil {
		return err
	}
	if err := scoring.NewRecomputer(store, scoring.DefaultRules()).RecomputeAll(ctx); err != nil {
		return err
	}
	logger.Info("resolve-bonuses complete", slog.Bool("finalPlayed", done))
	return nil
}

func run(logger *slog.Logger) error {
	// Fail closed: the server must never start with a missing/weak signing key.
	if err := auth.ValidateSecret(); err != nil {
		logger.Error("refusing to start: invalid JWT_SECRET", slog.Any("error", err))
		return err
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Migrations run BEFORE the HTTP server starts.
	startupCtx, cancelStartup := context.WithTimeout(context.Background(), 30*time.Second)
	store, err := openStore(startupCtx, logger)
	cancelStartup()
	if err != nil {
		return err
	}
	if store != nil {
		defer store.Close()
	}

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(requestLogger(logger))

	api.Version = version
	logger.Info("starting server", slog.String("version", version))
	api.RegisterHealthRoutes(engine)
	if store != nil {
		recomputer := scoring.NewRecomputer(store, scoring.DefaultRules())

		// Auth endpoints stay public (login/logout/Google); health is public too.
		auth.RegisterRoutes(engine, store)

		// Demo-mode gate: when demo mode is ON, non-admin users are gated by their
		// access_level (none/ro/rw). Installed globally; it no-ops for open routes
		// and for everyone when demo mode is OFF. Gin resolves the route before the
		// chain runs, so it can inspect the matched path.
		engine.Use(auth.DemoGate(store))

		// Auth wall: ALL data reads require a logged-in user. Anonymous callers
		// get 401 (the SPA shows the login screen). health + auth stay public.
		authed := engine.Group("", auth.RequireUser())
		api.RegisterReadRoutes(authed, store)                                 // /api/matches, /api/teams
		api.RegisterMatchDetailRoutes(authed, store, results.NewFIFAClient()) // /api/matches/:id/detail
		api.RegisterStandingsRoutes(authed, store)                            // /api/standings
		api.RegisterLeaderboardRoutes(authed, store)
		api.RegisterAuditRoutes(authed, store)

		// These self-gate per-route (RequireUser / RequireAdmin) internally.
		api.RegisterProfileRoutes(engine, store)
		api.RegisterHistoryRoutes(authed, store)
		api.RegisterTopScorersRoutes(authed, store)
		api.RegisterPredictionRoutes(engine, store, recomputer)
		api.RegisterBonusRoutes(engine, store)
		api.RegisterAdminRoutes(engine, store, recomputer)

		// Football AI assistant (ADR-0017): auth-only (on the authed group) but
		// NOT tier-gated — every logged-in user incl. `none` can use it. If ADC/WIF
		// is absent the assistant is Unavailable and its endpoints return 503.
		// Grounded in the app's live tournament data via tools (ADR-0018).
		aiAssistant := gemini.New(context.Background())
		aiAssistant.SetTools(api.NewAITools(store))
		api.RegisterAIRoutes(authed, aiAssistant)
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      engine,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	// Listen for interrupt/terminate signals to trigger graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Best-effort sync on boot so the calendar populates, then a periodic loop
	// keeps results fresh and drives Telegram announcements/reminders. Non-fatal.
	if store != nil {
		go bootSync(ctx, store, logger)
		go backgroundLoop(ctx, store, logger)
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("starting server", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		logger.Info("shutdown signal received, draining connections")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	logger.Info("server stopped cleanly")
	return nil
}

// bootSync runs a best-effort FIFA sync in the background. Failures are logged
// and never crash the server (resilience requirement).
func bootSync(ctx context.Context, store *storage.Store, logger *slog.Logger) {
	syncCtx, cancel := context.WithTimeout(ctx, bootSyncTimeout)
	defer cancel()

	syncer := syncpkg.New(results.NewFIFAClient(), store, logger)
	if _, err := syncer.Run(syncCtx); err != nil {
		logger.Warn("boot fifa sync failed (continuing)", slog.Any("error", err))
		return
	}
	// Resolve any knockout advancers, then re-score (both best-effort, idempotent).
	if _, err := winners.New(store, results.NewFIFAClient(), logger).Run(syncCtx); err != nil {
		logger.Warn("boot winners resolve failed (continuing)", slog.Any("error", err))
	}
	if err := scoring.NewRecomputer(store, scoring.DefaultRules()).RecomputeAll(syncCtx); err != nil {
		logger.Warn("boot recompute failed (continuing)", slog.Any("error", err))
	}
	// Best-effort: post any freshly-finished results to Telegram (idempotent —
	// each match is announced at most once).
	if tg, err := notify.TelegramFromEnv(); err != nil {
		logger.Warn("boot announce: telegram config error (continuing)", slog.Any("error", err))
	} else if sent, err := announce.New(store, tg, logger).Run(syncCtx); err != nil {
		logger.Warn("boot announce failed (continuing)", slog.Any("error", err))
	} else if sent > 0 {
		logger.Info("boot announce posted results", slog.Int("posted", sent))
	}

	// Best-effort: build the top-scorers board so it's populated immediately.
	if err := scorers.New(store, results.NewFIFAClient(), logger).Run(syncCtx); err != nil {
		logger.Warn("boot top-scorers aggregation failed (continuing)", slog.Any("error", err))
	}
}

// backgroundLoop drives the recurring jobs while the server runs. Every
// fastTickInterval it posts pre-match reminders and freshly-finished results to
// Telegram (cheap, DB-only); every syncEvery it also runs a FIFA sync + rescore
// to keep the calendar/results current. All steps are best-effort and never
// crash the server. Stops when ctx is cancelled (shutdown).
func backgroundLoop(ctx context.Context, store *storage.Store, logger *slog.Logger) {
	tg, _ := notify.TelegramFromEnv() // nil-safe: a disabled notifier no-ops
	ticker := time.NewTicker(fastTickInterval)
	defer ticker.Stop()

	var lastScorerAgg time.Time // top-scorer aggregation is heavier → hourly

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Every tick: refresh from FIFA + rescore so the calendar, live
			// statuses and leaderboard stay current (one light calendar request).
			syncCtx, cancel := context.WithTimeout(ctx, bootSyncTimeout)
			if _, err := syncpkg.New(results.NewFIFAClient(), store, logger).Run(syncCtx); err != nil {
				logger.Warn("periodic sync failed (continuing)", slog.Any("error", err))
			} else {
				// Resolve knockout advancers (ET/pens) BEFORE rescoring so the
				// +1 advancer point is awarded correctly.
				if _, err := winners.New(store, results.NewFIFAClient(), logger).Run(syncCtx); err != nil {
					logger.Warn("periodic winners resolve failed (continuing)", slog.Any("error", err))
				}
				if err := scoring.NewRecomputer(store, scoring.DefaultRules()).RecomputeAll(syncCtx); err != nil {
					logger.Warn("periodic recompute failed (continuing)", slog.Any("error", err))
				}
			}
			cancel()

			// Hourly: rebuild the top-scorers board (fetches match detail per
			// finished match, so it's heavier than the calendar sync).
			if time.Since(lastScorerAgg) >= time.Hour {
				lastScorerAgg = time.Now()
				aggCtx, cancelAgg := context.WithTimeout(ctx, syncCmdTimeout)
				if err := scorers.New(store, results.NewFIFAClient(), logger).Run(aggCtx); err != nil {
					logger.Warn("periodic top-scorers aggregation failed (continuing)", slog.Any("error", err))
				}
				cancelAgg()
			}

			// Every tick: pre-match reminders + result announcements.
			jobCtx, cancel := context.WithTimeout(ctx, syncCmdTimeout)
			if n, err := remind.New(store, tg, logger).Run(jobCtx); err != nil {
				logger.Warn("periodic remind failed (continuing)", slog.Any("error", err))
			} else if n > 0 {
				logger.Info("periodic remind sent", slog.Int("sent", n))
			}
			// One-off nudge to set tournament bonuses before they lock at R16 start.
			if sent, err := remind.NewBonus(store, tg, logger).Run(jobCtx, time.Now().UTC()); err != nil {
				logger.Warn("periodic bonus remind failed (continuing)", slog.Any("error", err))
			} else if sent {
				logger.Info("bonus deadline reminder sent")
			}
			if n, err := announce.New(store, tg, logger).Run(jobCtx); err != nil {
				logger.Warn("periodic announce failed (continuing)", slog.Any("error", err))
			} else if n > 0 {
				logger.Info("periodic announce posted", slog.Int("posted", n))
			}

			// Once per day, at/after the configured Kyiv hour, post the digest.
			maybeDailyDigest(jobCtx, store, tg, logger)
			// Once the final is played, award the tournament bonuses (runs once).
			maybeResolveBonuses(jobCtx, store, logger)
			cancel()
		}
	}
}

// maybeDailyDigest posts the morning digest at most once per Kyiv calendar day,
// firing on the first tick at/after DIGEST_HOUR_KYIV (default 10), within a short
// window so a late server start doesn't post at an odd hour. The "last sent day"
// is persisted in app_state so it survives restarts.
func maybeDailyDigest(ctx context.Context, store *storage.Store, tg *notify.Telegram, logger *slog.Logger) {
	hour := envInt("DIGEST_HOUR_KYIV", 10)
	now := time.Now().In(kyivLoc)
	// Only within [hour, hour+3) so a restart at, say, 18:00 waits for tomorrow.
	if now.Hour() < hour || now.Hour() >= hour+3 {
		return
	}
	today := now.Format("2006-01-02")
	last, _, err := store.GetAppState(ctx, digestStateKey)
	if err != nil {
		logger.Warn("digest: read state failed (continuing)", slog.Any("error", err))
		return
	}
	if last == today {
		return // already done today
	}
	sent, err := digest.New(store, tg, kyivLoc, logger).Run(ctx)
	if err != nil {
		logger.Warn("digest failed (continuing)", slog.Any("error", err))
		return // don't stamp; retry next tick
	}
	// Stamp the day even when nothing was sent (quiet/rest day) so we don't retry.
	if err := store.SetAppState(ctx, digestStateKey, today); err != nil {
		logger.Warn("digest: write state failed (continuing)", slog.Any("error", err))
	}
	if sent {
		logger.Info("daily digest posted", slog.String("day", today))
	}
}

// maybeResolveBonuses awards the tournament bonuses (champion/finalist/top
// scorer) once the final has been played, then rescores. Gated by an app_state
// flag so the heavy top-scorer aggregation runs at most once.
func maybeResolveBonuses(ctx context.Context, store *storage.Store, logger *slog.Logger) {
	if done, _, err := store.GetAppState(ctx, bonusesResolvedStateKey); err == nil && done == "true" {
		return
	}
	resolved, err := resolve.New(store, results.NewFIFAClient(), logger).Run(ctx)
	if err != nil {
		logger.Warn("resolve bonuses failed (continuing)", slog.Any("error", err))
		return
	}
	if !resolved {
		return // final not played yet
	}
	if err := scoring.NewRecomputer(store, scoring.DefaultRules()).RecomputeAll(ctx); err != nil {
		logger.Warn("resolve: rescore failed (continuing)", slog.Any("error", err))
	}
	if err := store.SetAppState(ctx, bonusesResolvedStateKey, "true"); err != nil {
		logger.Warn("resolve: write state failed (continuing)", slog.Any("error", err))
	}
	logger.Info("tournament bonuses resolved & awarded")
}

// envInt reads an int env var, falling back to def when unset/invalid.
func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// requestLogger is a minimal Gin middleware that logs each request via slog.
func requestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		logger.Info("request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}
