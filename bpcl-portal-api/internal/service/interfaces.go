package service

import (
	"context"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/google/uuid"
)

type UserRepository interface {
	GetByEmployeeID(ctx context.Context, employeeID string) (*model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	List(ctx context.Context, role, territory string, limit, offset int) ([]*model.User, int, error)
	Create(ctx context.Context, u *model.User) error
	Update(ctx context.Context, id uuid.UUID, name string, role model.UserRole, territory *string, isActive bool) error
	UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
}

type OutletRepository interface {
	GetByCC(ctx context.Context, cc string) (*model.RetailOutlet, error)
	ListByTerritory(ctx context.Context, territoryCode string) ([]*model.RetailOutlet, error)
	ListAll(ctx context.Context) ([]*model.RetailOutlet, error)
	GetByTerritoryWithPerformance(ctx context.Context, territoryCode string, period time.Time) ([]*model.TerritoryOutlet, error)
}

type PerformanceRepository interface {
	GetByPeriod(ctx context.Context, cc string, period time.Time) ([]*model.PerformanceRecord, error)
	GetDateRange(ctx context.Context, cc string, from, to time.Time) ([]*model.PerformanceRecord, error)
	GetTrend(ctx context.Context, cc string, months int) ([]*model.TrendRow, error)
	Upsert(ctx context.Context, rec *model.PerformanceRecord) error
}

type TargetRepository interface {
	GetByPeriod(ctx context.Context, cc string, period time.Time) ([]*model.Target, error)
	UpsertBatch(ctx context.Context, targets []*model.Target) error
}

type UploadRepository interface {
	Create(ctx context.Context, f *model.UploadedFile) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.UploadedFile, error)
	ListByUser(ctx context.Context, userID uuid.UUID, cc string, limit, offset int) ([]*model.UploadedFile, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status, errMsg string, rowCount int) error
}

type CompetitionRepository interface {
	GetActivePeriod(ctx context.Context, territoryCode string) (*model.CompetitionPeriod, error)
	GetScores(ctx context.Context, competitionID uuid.UUID) ([]*model.CompetitionScore, error)
	GetDealerScore(ctx context.Context, competitionID uuid.UUID, cc string) (*model.CompetitionScore, error)
	GetAuditScore(ctx context.Context, cc string, period time.Time) (*model.DealerAuditScore, error)
	GetBonus(ctx context.Context, competitionID uuid.UUID, cc string) (*model.CompetitionBonus, error)
}

type AuditRepository interface {
	Log(ctx context.Context, userID uuid.UUID, action, cc string, payload any) error
}

type OTPRepository interface {
	DeleteByEmail(ctx context.Context, email string) error
	Insert(ctx context.Context, email, otpHash string, expiresAt time.Time) error
	GetLatestUnused(ctx context.Context, email string) (otpHash string, expiresAt time.Time, err error)
	MarkUsed(ctx context.Context, email string) error
}
