package router

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/bpcl/portal-api/internal/admin"
	"github.com/bpcl/portal-api/internal/config"
	"github.com/bpcl/portal-api/internal/crystal/ingest"
	"github.com/bpcl/portal-api/internal/dashboard"
	"github.com/bpcl/portal-api/internal/handler"
	"github.com/bpcl/portal-api/internal/middleware"
	"github.com/bpcl/portal-api/internal/portal"
	"github.com/bpcl/portal-api/internal/scoring"
)

// dbPinger is satisfied by *pgxpool.Pool.
type dbPinger interface {
	Ping(ctx context.Context) error
}

// New builds and returns the fully-wired HTTP handler.
// Middleware chain (outermost → innermost): RequestID → Logger → CORS → RateLimit → [Auth on protected].
func New(
	h *handler.Handler,
	cfg *config.Config,
	authSvc middleware.TokenValidator,
	db dbPinger,
	crystalH *ingest.Handler,
	adminH *admin.Handler,
	portalH *portal.Handler,
	dashH *dashboard.Handler,
	dashRepo *dashboard.Repository,
	scoringH *scoring.Handler,
) http.Handler {
	mux := http.NewServeMux()

	authMw := middleware.Auth(authSvc)

	// ── Public routes ──────────────────────────────────────────────────────────
	mux.HandleFunc("GET /health", healthHandler(db))
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)

	// ── Protected: auth ────────────────────────────────────────────────────────
	mux.Handle("GET /api/v1/auth/me", authMw(http.HandlerFunc(h.Me)))

	// ── Protected: outlets ─────────────────────────────────────────────────────
	mux.Handle("GET /api/v1/outlets", authMw(http.HandlerFunc(h.ListOutlets)))
	mux.Handle("GET /api/v1/outlets/{cc}", authMw(http.HandlerFunc(h.GetOutlet)))
	mux.Handle("GET /api/v1/outlets/{cc}/performance", authMw(http.HandlerFunc(h.GetPerformance)))
	mux.Handle("GET /api/v1/outlets/{cc}/analysis", authMw(http.HandlerFunc(h.GetAnalysis)))
	mux.Handle("GET /api/v1/outlets/{cc}/trend", authMw(http.HandlerFunc(h.GetTrend)))
	mux.Handle("GET /api/v1/outlets/{cc}/targets", authMw(http.HandlerFunc(h.GetTargets)))
	mux.Handle("PUT /api/v1/outlets/{cc}/targets", authMw(http.HandlerFunc(h.SetTargets)))

	// ── Protected: territory ────────────────────────────────────────────────
	mux.Handle("GET /api/v1/territory/outlets", authMw(http.HandlerFunc(h.GetTerritoryOutlets)))
	mux.Handle("GET /api/v1/territory/summary", authMw(http.HandlerFunc(h.GetTerritorySummary)))

	// ── Protected: admin ────────────────────────────────────────────────
	mux.Handle("GET /api/v1/admin/users", authMw(http.HandlerFunc(h.ListUsers)))
	mux.Handle("POST /api/v1/admin/users", authMw(http.HandlerFunc(h.CreateUser)))
	mux.Handle("PUT /api/v1/admin/users/{id}", authMw(http.HandlerFunc(h.UpdateUser)))
	mux.Handle("PUT /api/v1/admin/users/{id}/reset-password", authMw(http.HandlerFunc(h.ResetPassword)))

	// ── Protected: uploads ─────────────────────────────────────────────────────
	// HandleUpload uses r.PathValue("cc"), so cc must be in the path.
	mux.Handle("POST /api/v1/outlets/{cc}/uploads", authMw(http.HandlerFunc(h.HandleUpload)))
	mux.Handle("GET /api/v1/uploads", authMw(http.HandlerFunc(h.GetUploadHistory)))

	// ── Protected: competition ─────────────────────────────────────────────────
	mux.Handle("GET /api/v1/competition/leaderboard", authMw(http.HandlerFunc(h.GetLeaderboard)))
	// {competition_id} comes before /dealers/{cc} — fixed "leaderboard" resolves before wildcard.
	mux.Handle("GET /api/v1/competition/{competition_id}/dealers/{cc}", authMw(http.HandlerFunc(h.GetDealerScorecard)))
	mux.Handle("GET /api/v1/competition/{id}/market-share-status", authMw(http.HandlerFunc(h.GetMarketShareStatus)))
	mux.Handle("POST /api/v1/competition/{id}/recompute", authMw(http.HandlerFunc(h.RecomputeCompetitionScores)))

	// ── Protected: uploads (Delhi Master) ─────────────────────────────────────
	mux.Handle("POST /api/v1/uploads/delhi-master", authMw(http.HandlerFunc(h.HandleDelhiMasterUpload)))

	// ── Crystal ingest routes (auth added in Chunk 2b) ─────────────────────────
	mux.HandleFunc("POST /api/v1/ingest/daily-bulk", crystalH.DailyBulk)
	mux.HandleFunc("POST /api/v1/ingest/mak-ge", crystalH.MAKGE)
	mux.HandleFunc("POST /api/v1/ingest/google-rating", crystalH.GoogleRating)
	mux.HandleFunc("POST /api/v1/targets/bulk", crystalH.TargetsBulk)

	// ── Crystal Chunk 3: scoring engine ───────────────────────────────────────
	mux.Handle("POST /api/v1/scoring/compute", authMw(http.HandlerFunc(scoringH.Compute)))

	// ── Crystal Chunk 4: admin portal ─────────────────────────────────────────
	mux.Handle("GET /api/v1/admin/dealers", authMw(http.HandlerFunc(adminH.ListDealers)))
	mux.Handle("PUT /api/v1/admin/dealers/{cc_code}", authMw(http.HandlerFunc(adminH.ToggleDealer)))
	mux.Handle("GET /api/v1/admin/competition-periods", authMw(http.HandlerFunc(adminH.ListPeriods)))
	mux.Handle("POST /api/v1/admin/competition-periods", authMw(http.HandlerFunc(adminH.CreatePeriod)))
	mux.Handle("PATCH /api/v1/admin/competition-periods/{month_year}/activate", authMw(http.HandlerFunc(adminH.ActivatePeriod)))
	mux.Handle("GET /api/v1/admin/targets/{month_year}", authMw(http.HandlerFunc(adminH.GetTargets)))
	mux.Handle("GET /api/v1/admin/ingest/summary", authMw(http.HandlerFunc(adminH.IngestSummary)))
	mux.Handle("GET /api/v1/admin/mak-ge/{month_year}", authMw(http.HandlerFunc(adminH.GetMAKGE)))
	mux.Handle("GET /api/v1/admin/manual-scores/{month_year}", authMw(http.HandlerFunc(adminH.GetManualScores)))
	mux.Handle("POST /api/v1/admin/manual-scores", authMw(http.HandlerFunc(adminH.SaveManualScores)))

	// ── Crystal Chunk 5: SSO portal ───────────────────────────────────────────
	mux.Handle("GET /api/v1/portal/{cc_code}/scorecard", authMw(http.HandlerFunc(portalH.ScoreCard)))

	// ── Crystal Chunk 6: dashboard ────────────────────────────────────────────
	mux.Handle("GET /api/v1/dashboard", authMw(http.HandlerFunc(dashH.GetDashboard)))
	mux.Handle("GET /api/v1/dashboard/export", authMw(dashboard.ExportHandler(dashRepo)))

	// ── Global middleware wraps the entire mux ─────────────────────────────────
	rl := middleware.NewRateLimiter(cfg.RateLimitRPM)
	return middleware.SecurityHeaders(
		middleware.RequestTimeout(30*time.Second)(
			middleware.RequestID(
				middleware.Logger(cfg.Env)(
					middleware.CORS(cfg.CORSOrigins)(
						rl.Middleware()(mux),
					),
				),
			),
		),
	)
}

func healthHandler(db dbPinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		dbStatus := "ok"
		if err := db.Ping(ctx); err != nil {
			dbStatus = "error"
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"db":      dbStatus,
			"version": "1.0.0",
		})
	}
}
