package router

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/bpcl/portal-api/internal/config"
	"github.com/bpcl/portal-api/internal/handler"
	"github.com/bpcl/portal-api/internal/middleware"
)

// dbPinger is satisfied by *pgxpool.Pool.
type dbPinger interface {
	Ping(ctx context.Context) error
}

// New builds and returns the fully-wired HTTP handler.
// Middleware chain (outermost → innermost): RequestID → Logger → CORS → RateLimit → [Auth on protected].
func New(h *handler.Handler, cfg *config.Config, authSvc middleware.TokenValidator, db dbPinger) http.Handler {
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
