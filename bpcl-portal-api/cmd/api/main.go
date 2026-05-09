package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bpcl/portal-api/internal/admin"
	"github.com/bpcl/portal-api/internal/config"
	"github.com/bpcl/portal-api/internal/crystal/dealer"
	"github.com/bpcl/portal-api/internal/crystal/ingest"
	"github.com/bpcl/portal-api/internal/crystal/period"
	"github.com/bpcl/portal-api/internal/dashboard"
	"github.com/bpcl/portal-api/internal/handler"
	"github.com/bpcl/portal-api/internal/portal"
	"github.com/bpcl/portal-api/internal/repository"
	"github.com/bpcl/portal-api/internal/router"
	"github.com/bpcl/portal-api/internal/scoring"
	"github.com/bpcl/portal-api/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Validate production config.
	if err := cfg.Validate(); err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	// Structured logger: JSON in production, text in development.
	var logLevel slog.Level
	if cfg.LogLevel == "debug" {
		logLevel = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{Level: logLevel}
	var logger *slog.Logger
	if cfg.Env == "production" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	slog.SetDefault(logger)

	// Database pool.
	ctx := context.Background()
	pool, err := repository.Connect(ctx, cfg.DBUrl)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer repository.Close(pool)
	slog.Info("database connected")

// Repositories.
	userRepo := repository.NewUserRepo(pool)
	outletRepo := repository.NewOutletRepo(pool)
	perfRepo := repository.NewPerformanceRepo(pool)
	targetRepo := repository.NewTargetRepo(pool)
	uploadRepo := repository.NewUploadRepo(pool)
	competitionRepo := repository.NewCompetitionRepo(pool)
	auditRepo := repository.NewAuditRepo(pool)
	marketShareRepo := repository.NewMarketShareRepo(pool)

	// Services.
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiry)
	outletSvc := service.NewOutletService(outletRepo, auditRepo)
	perfSvc := service.NewPerformanceService(outletRepo, perfRepo, targetRepo)
	targetSvc := service.NewTargetService(outletRepo, targetRepo, auditRepo)
	uploadSvc := service.NewUploadService(uploadRepo, perfRepo, cfg.UploadDir, cfg.MaxUploadMB)
	compSvc := service.NewCompetitionService(competitionRepo, outletRepo)
	marketShareSvc := service.NewMarketShareService(marketShareRepo, outletRepo, cfg.UploadDir)
	userSvc := service.NewUserService(userRepo)

	// Handler.
	h := handler.New(authSvc, outletSvc, perfSvc, targetSvc, uploadSvc, compSvc, marketShareSvc, userSvc)

	// Crystal ingest wiring.
	dealerRepo := dealer.New(pool)
	periodRepo := period.New(pool)
	ingestRepo := ingest.NewRepository(pool)
	ingestSvc := ingest.NewService(ingestRepo, dealerRepo, periodRepo)
	crystalH := ingest.NewHandler(ingestSvc)

	// Crystal Chunk 3: scoring engine.
	actualsRepo := scoring.NewActualsRepo(pool)
	paramsRepo := scoring.NewParamsRepo(pool)
	scoreRepo := scoring.NewScoreRepository(pool)
	scoringEngine := scoring.NewEngine(actualsRepo, paramsRepo, scoreRepo)
	scoringH := scoring.NewHandler(scoringEngine)

	// Crystal Chunk 4: admin portal.
	adminRepo := admin.NewRepository(pool)
	adminH := admin.NewHandler(adminRepo)

	// Crystal Chunk 5: SSO portal.
	portalRepo := portal.NewRepository(pool)
	portalH := portal.NewHandler(portalRepo)

	// Crystal Chunk 6: dashboard.
	dashRepo := dashboard.NewRepository(pool)
	dashH := dashboard.NewHandler(dashRepo)

	// Router.
	r := router.New(h, cfg, authSvc, pool, crystalH, adminH, portalH, dashH, dashRepo, scoringH)

	// HTTP server.
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	slog.Info("starting server", "addr", srv.Addr, "env", cfg.Env)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown on SIGINT / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}
