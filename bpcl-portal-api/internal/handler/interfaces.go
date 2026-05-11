package handler

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/bpcl/portal-api/internal/service"
	"github.com/google/uuid"
)

type AuthServiceI interface {
	Login(ctx context.Context, employeeID, password string) (*service.LoginResponse, error)
	ValidateToken(tokenStr string) (*service.Claims, error)
	RefreshAccessToken(refreshToken string) (*service.RefreshResponse, error)
}

type OutletServiceI interface {
	GetOutlet(ctx context.Context, cc string, userID uuid.UUID) (*model.RetailOutlet, error)
	ListOutlets(ctx context.Context, userID uuid.UUID) ([]*model.RetailOutlet, error)
	GetTerritorySummary(ctx context.Context, periodStr string) (*model.TerritorySummary, error)
}

type PerformanceServiceI interface {
	GetPerformance(ctx context.Context, cc string, period time.Time, userID uuid.UUID) (*service.PerformanceResponse, error)
	GetAnalysis(ctx context.Context, cc string, from, to time.Time, userID uuid.UUID) (*service.AnalysisResponse, error)
	GetTrend(ctx context.Context, cc string, months int) (*model.TrendResponse, error)
}

type TargetServiceI interface {
	GetTargets(ctx context.Context, cc string, period time.Time, userID uuid.UUID) (*model.TargetsResponse, error)
	SetTargets(ctx context.Context, cc string, period time.Time, input service.TargetInput, userID uuid.UUID) (*model.TargetsResponse, error)
}

type UploadServiceI interface {
	HandleUpload(ctx context.Context, file multipart.File, header *multipart.FileHeader, cc string, period time.Time, userID uuid.UUID) (*service.UploadResponse, error)
	GetHistory(ctx context.Context, userID uuid.UUID, cc string, limit, offset int) (*service.UploadListResponse, error)
}

type CompetitionServiceI interface {
	GetLeaderboard(ctx context.Context, territoryCode string, userID uuid.UUID) (*service.LeaderboardResponse, error)
	GetDealerScorecard(ctx context.Context, cc string, competitionID uuid.UUID, userID uuid.UUID) (*service.ScorecardResponse, error)
}

type UserServiceI interface {
	ListUsers(ctx context.Context, p service.ListUsersParams) (*service.ListUsersResponse, error)
	CreateUser(ctx context.Context, input service.CreateUserInput) (*model.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, input service.UpdateUserInput) (*model.User, error)
	ResetPassword(ctx context.Context, id uuid.UUID, input service.ResetPasswordInput) error
}

type OTPServiceI interface {
	GenerateAndSend(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, email, otp, newPassword string) error
}
