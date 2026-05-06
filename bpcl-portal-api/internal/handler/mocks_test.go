package handler_test

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/bpcl/portal-api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// ── Auth service mock ────────────────────────────────────────────────────────

type mockAuthSvc struct{ mock.Mock }

func (m *mockAuthSvc) Login(ctx context.Context, empID, password string) (*service.LoginResponse, error) {
	args := m.Called(ctx, empID, password)
	if r := args.Get(0); r != nil {
		return r.(*service.LoginResponse), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockAuthSvc) ValidateToken(token string) (*service.Claims, error) {
	args := m.Called(token)
	if c := args.Get(0); c != nil {
		return c.(*service.Claims), args.Error(1)
	}
	return nil, args.Error(1)
}

// ── Outlet service mock ──────────────────────────────────────────────────────

type mockOutletSvc struct{ mock.Mock }

func (m *mockOutletSvc) GetOutlet(ctx context.Context, cc string, userID uuid.UUID) (*model.RetailOutlet, error) {
	args := m.Called(ctx, cc, userID)
	if o := args.Get(0); o != nil {
		return o.(*model.RetailOutlet), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockOutletSvc) ListOutlets(ctx context.Context, userID uuid.UUID) ([]*model.RetailOutlet, error) {
	args := m.Called(ctx, userID)
	if o := args.Get(0); o != nil {
		return o.([]*model.RetailOutlet), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockOutletSvc) GetTerritorySummary(ctx context.Context, periodStr string) (*model.TerritorySummary, error) {
	args := m.Called(ctx, periodStr)
	if r := args.Get(0); r != nil {
		return r.(*model.TerritorySummary), args.Error(1)
	}
	return nil, args.Error(1)
}

// ── Performance service mock ─────────────────────────────────────────────────

type mockPerfSvc struct{ mock.Mock }

func (m *mockPerfSvc) GetPerformance(ctx context.Context, cc string, period time.Time, userID uuid.UUID) (*service.PerformanceResponse, error) {
	args := m.Called(ctx, cc, period, userID)
	if r := args.Get(0); r != nil {
		return r.(*service.PerformanceResponse), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockPerfSvc) GetAnalysis(ctx context.Context, cc string, from, to time.Time, userID uuid.UUID) (*service.AnalysisResponse, error) {
	args := m.Called(ctx, cc, from, to, userID)
	if r := args.Get(0); r != nil {
		return r.(*service.AnalysisResponse), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockPerfSvc) GetTrend(ctx context.Context, cc string, months int) (*model.TrendResponse, error) {
	args := m.Called(ctx, cc, months)
	if r := args.Get(0); r != nil {
		return r.(*model.TrendResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

// ── Target service mock ──────────────────────────────────────────────────────

type mockTargetSvc struct{ mock.Mock }

func (m *mockTargetSvc) GetTargets(ctx context.Context, cc string, period time.Time, userID uuid.UUID) (*model.TargetsResponse, error) {
	args := m.Called(ctx, cc, period, userID)
	if r := args.Get(0); r != nil {
		return r.(*model.TargetsResponse), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockTargetSvc) SetTargets(ctx context.Context, cc string, period time.Time, input service.TargetInput, userID uuid.UUID) (*model.TargetsResponse, error) {
	args := m.Called(ctx, cc, period, input, userID)
	if r := args.Get(0); r != nil {
		return r.(*model.TargetsResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

// ── Upload service mock ──────────────────────────────────────────────────────

type mockUploadSvc struct{ mock.Mock }

func (m *mockUploadSvc) HandleUpload(ctx context.Context, file multipart.File, header *multipart.FileHeader, cc string, period time.Time, userID uuid.UUID) (*service.UploadResponse, error) {
	args := m.Called(ctx, file, header, cc, period, userID)
	if r := args.Get(0); r != nil {
		return r.(*service.UploadResponse), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockUploadSvc) GetHistory(ctx context.Context, userID uuid.UUID, cc string, limit, offset int) (*service.UploadListResponse, error) {
	args := m.Called(ctx, userID, cc, limit, offset)
	if r := args.Get(0); r != nil {
		return r.(*service.UploadListResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

// ── Competition service mock ─────────────────────────────────────────────────

type mockCompSvc struct{ mock.Mock }

func (m *mockCompSvc) GetLeaderboard(ctx context.Context, territory string, userID uuid.UUID) (*service.LeaderboardResponse, error) {
	args := m.Called(ctx, territory, userID)
	if r := args.Get(0); r != nil {
		return r.(*service.LeaderboardResponse), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockCompSvc) GetDealerScorecard(ctx context.Context, cc string, compID uuid.UUID, userID uuid.UUID) (*service.ScorecardResponse, error) {
	args := m.Called(ctx, cc, compID, userID)
	if r := args.Get(0); r != nil {
		return r.(*service.ScorecardResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

// ── MarketShare service mock ─────────────────────────────────────────────────

type mockMarketShareSvc struct{ mock.Mock }

func (m *mockMarketShareSvc) IngestDelhiMaster(ctx context.Context, filePath string, uploadedFileID uuid.UUID) (int, int, error) {
	args := m.Called(ctx, filePath, uploadedFileID)
	return args.Int(0), args.Int(1), args.Error(2)
}
func (m *mockMarketShareSvc) GetMarketShareStatus(ctx context.Context, competitionID uuid.UUID) (*model.MarketShareStatusResponse, error) {
	args := m.Called(ctx, competitionID)
	if r := args.Get(0); r != nil {
		return r.(*model.MarketShareStatusResponse), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockMarketShareSvc) FullRecompute(ctx context.Context, competitionID uuid.UUID) (*model.RecomputeResponse, error) {
	args := m.Called(ctx, competitionID)
	if r := args.Get(0); r != nil {
		return r.(*model.RecomputeResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

// ── User service mock ────────────────────────────────────────────────────────

type mockUserSvc struct{ mock.Mock }

func (m *mockUserSvc) ListUsers(ctx context.Context, p service.ListUsersParams) (*service.ListUsersResponse, error) {
	args := m.Called(ctx, p)
	if r := args.Get(0); r != nil {
		return r.(*service.ListUsersResponse), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockUserSvc) CreateUser(ctx context.Context, input service.CreateUserInput) (*model.User, error) {
	args := m.Called(ctx, input)
	if r := args.Get(0); r != nil {
		return r.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockUserSvc) UpdateUser(ctx context.Context, id uuid.UUID, input service.UpdateUserInput) (*model.User, error) {
	args := m.Called(ctx, id, input)
	if r := args.Get(0); r != nil {
		return r.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockUserSvc) ResetPassword(ctx context.Context, id uuid.UUID, input service.ResetPasswordInput) error {
	return m.Called(ctx, id, input).Error(0)
}
