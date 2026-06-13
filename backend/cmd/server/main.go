// Command server runs the WC-Tournament HTTP backend.
//
// Subcommands:
//
//	server               run the HTTP server (default)
//	server healthcheck   probe /healthz and exit 0/1 (container HEALTHCHECK)
//	server sync          run one FIFA calendar sync and exit
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/obezsmertnyi/WC-Tournament/backend/internal/api"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/auth"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/results"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/scoring"
	"github.com/obezsmertnyi/WC-Tournament/backend/internal/storage"
	syncpkg "github.com/obezsmertnyi/WC-Tournament/backend/internal/sync"
)

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

	api.RegisterHealthRoutes(engine)
	if store != nil {
		recomputer := scoring.NewRecomputer(store, scoring.DefaultRules())

		// Auth endpoints stay public (login/logout/Google); health is public too.
		auth.RegisterRoutes(engine, store)

		// Auth wall: ALL data reads require a logged-in user. Anonymous callers
		// get 401 (the SPA shows the login screen). health + auth stay public.
		authed := engine.Group("", auth.RequireUser())
		api.RegisterReadRoutes(authed, store)      // /api/matches, /api/teams
		api.RegisterStandingsRoutes(authed, store) // /api/standings
		api.RegisterLeaderboardRoutes(authed, store)
		api.RegisterAuditRoutes(authed, store)

		// These self-gate per-route (RequireUser / RequireAdmin) internally.
		api.RegisterProfileRoutes(engine, store)
		api.RegisterPredictionRoutes(engine, store, recomputer)
		api.RegisterBonusRoutes(engine, store)
		api.RegisterAdminRoutes(engine, store, recomputer)
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

	// Best-effort sync on boot so the calendar populates. Non-fatal on error.
	if store != nil {
		go bootSync(ctx, store, logger)
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
	// Best-effort re-score after the boot sync (idempotent).
	if err := scoring.NewRecomputer(store, scoring.DefaultRules()).RecomputeAll(syncCtx); err != nil {
		logger.Warn("boot recompute failed (continuing)", slog.Any("error", err))
	}
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
