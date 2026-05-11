package service_test

import (
	"context"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// ── UserRepository mock ──────────────────────────────────────────────────────

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) GetByEmployeeID(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	if u := args.Get(0); u != nil {
		return u.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	if u := args.Get(0); u != nil {
		return u.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockUserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockUserRepo) List(ctx context.Context, role, territory string, limit, offset int) ([]*model.User, int, error) {
	args := m.Called(ctx, role, territory, limit, offset)
	if u := args.Get(0); u != nil {
		return u.([]*model.User), args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}
func (m *mockUserRepo) Create(ctx context.Context, u *model.User) error {
	return m.Called(ctx, u).Error(0)
}
func (m *mockUserRepo) Update(ctx context.Context, id uuid.UUID, name string, role model.UserRole, territory *string, isActive bool) error {
	return m.Called(ctx, id, name, role, territory, isActive).Error(0)
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if u := args.Get(0); u != nil {
		return u.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockUserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error {
	return m.Called(ctx, id, hash).Error(0)
}

// ── OutletRepository mock ────────────────────────────────────────────────────

type mockOutletRepo struct{ mock.Mock }

func (m *mockOutletRepo) GetByCC(ctx context.Context, cc string) (*model.RetailOutlet, error) {
	args := m.Called(ctx, cc)
	if o := args.Get(0); o != nil {
		return o.(*model.RetailOutlet), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockOutletRepo) ListByTerritory(ctx context.Context, tc string) ([]*model.RetailOutlet, error) {
	args := m.Called(ctx, tc)
	if o := args.Get(0); o != nil {
		return o.([]*model.RetailOutlet), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockOutletRepo) ListAll(ctx context.Context) ([]*model.RetailOutlet, error) {
	args := m.Called(ctx)
	if o := args.Get(0); o != nil {
		return o.([]*model.RetailOutlet), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockOutletRepo) GetByTerritoryWithPerformance(ctx context.Context, territoryCode string, period time.Time) ([]*model.TerritoryOutlet, error) {
	args := m.Called(ctx, territoryCode, period)
	if o := args.Get(0); o != nil {
		return o.([]*model.TerritoryOutlet), args.Error(1)
	}
	return nil, args.Error(1)
}

// ── PerformanceRepository mock ───────────────────────────────────────────────

type mockPerfRepo struct{ mock.Mock }

func (m *mockPerfRepo) GetByPeriod(ctx context.Context, cc string, period time.Time) ([]*model.PerformanceRecord, error) {
	args := m.Called(ctx, cc, period)
	if r := args.Get(0); r != nil {
		return r.([]*model.PerformanceRecord), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockPerfRepo) GetDateRange(ctx context.Context, cc string, from, to time.Time) ([]*model.PerformanceRecord, error) {
	args := m.Called(ctx, cc, from, to)
	if r := args.Get(0); r != nil {
		return r.([]*model.PerformanceRecord), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockPerfRepo) Upsert(ctx context.Context, rec *model.PerformanceRecord) error {
	return m.Called(ctx, rec).Error(0)
}
func (m *mockPerfRepo) GetTrend(ctx context.Context, cc string, months int) ([]*model.TrendRow, error) {
	args := m.Called(ctx, cc, months)
	if r := args.Get(0); r != nil {
		return r.([]*model.TrendRow), args.Error(1)
	}
	return nil, args.Error(1)
}

// ── TargetRepository mock ────────────────────────────────────────────────────

type mockTargetRepo struct{ mock.Mock }

func (m *mockTargetRepo) GetByPeriod(ctx context.Context, cc string, period time.Time) ([]*model.Target, error) {
	args := m.Called(ctx, cc, period)
	if t := args.Get(0); t != nil {
		return t.([]*model.Target), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockTargetRepo) UpsertBatch(ctx context.Context, targets []*model.Target) error {
	return m.Called(ctx, targets).Error(0)
}

// ── UploadRepository mock ────────────────────────────────────────────────────

type mockUploadRepo struct{ mock.Mock }

func (m *mockUploadRepo) Create(ctx context.Context, f *model.UploadedFile) error {
	return m.Called(ctx, f).Error(0)
}
func (m *mockUploadRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.UploadedFile, error) {
	args := m.Called(ctx, id)
	if f := args.Get(0); f != nil {
		return f.(*model.UploadedFile), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockUploadRepo) ListByUser(ctx context.Context, userID uuid.UUID, cc string, limit, offset int) ([]*model.UploadedFile, int, error) {
	args := m.Called(ctx, userID, cc, limit, offset)
	if items := args.Get(0); items != nil {
		return items.([]*model.UploadedFile), args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}
func (m *mockUploadRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status, errMsg string, rowCount int) error {
	return m.Called(ctx, id, status, errMsg, rowCount).Error(0)
}

// ── CompetitionRepository mock ───────────────────────────────────────────────

type mockCompRepo struct{ mock.Mock }

func (m *mockCompRepo) GetActivePeriod(ctx context.Context, tc string) (*model.CompetitionPeriod, error) {
	args := m.Called(ctx, tc)
	if p := args.Get(0); p != nil {
		return p.(*model.CompetitionPeriod), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockCompRepo) GetScores(ctx context.Context, id uuid.UUID) ([]*model.CompetitionScore, error) {
	args := m.Called(ctx, id)
	if s := args.Get(0); s != nil {
		return s.([]*model.CompetitionScore), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockCompRepo) GetDealerScore(ctx context.Context, compID uuid.UUID, cc string) (*model.CompetitionScore, error) {
	args := m.Called(ctx, compID, cc)
	if s := args.Get(0); s != nil {
		return s.(*model.CompetitionScore), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockCompRepo) GetAuditScore(ctx context.Context, cc string, period time.Time) (*model.DealerAuditScore, error) {
	args := m.Called(ctx, cc, period)
	if a := args.Get(0); a != nil {
		return a.(*model.DealerAuditScore), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockCompRepo) GetBonus(ctx context.Context, compID uuid.UUID, cc string) (*model.CompetitionBonus, error) {
	args := m.Called(ctx, compID, cc)
	if b := args.Get(0); b != nil {
		return b.(*model.CompetitionBonus), args.Error(1)
	}
	return nil, args.Error(1)
}

// ── AuditRepository mock ─────────────────────────────────────────────────────

type mockAuditRepo struct{ mock.Mock }

func (m *mockAuditRepo) Log(ctx context.Context, userID uuid.UUID, action, cc string, payload any) error {
	return m.Called(ctx, userID, action, cc, payload).Error(0)
}
