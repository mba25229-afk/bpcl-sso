package service

import (
	"context"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/google/uuid"
)

type OutletService struct {
	outlets OutletRepository
	audit   AuditRepository
}

func NewOutletService(outlets OutletRepository, audit AuditRepository) *OutletService {
	return &OutletService{outlets: outlets, audit: audit}
}

func (s *OutletService) GetOutlet(ctx context.Context, cc string, userID uuid.UUID) (*model.RetailOutlet, error) {
	outlet, err := s.outlets.GetByCC(ctx, cc)
	if err != nil {
		return nil, err
	}
	claims, err := ExtractClaims(ctx)
	if err != nil {
		return nil, err
	}
	if !canAccess(claims, outlet) {
		return nil, model.ErrForbidden
	}
	_ = s.audit.Log(ctx, userID, "view_outlet", cc, nil)
	return outlet, nil
}

func (s *OutletService) GetTerritorySummary(ctx context.Context, periodStr string) (*model.TerritorySummary, error) {
	claims, err := ExtractClaims(ctx)
	if err != nil {
		return nil, err
	}

	role := model.UserRole(claims.Role)
	if role == model.RoleROManager {
		return nil, model.ErrForbidden
	}

	territoryCode := ""
	if claims.TerritoryCode != nil {
		territoryCode = *claims.TerritoryCode
	} else if role == model.RoleAdmin {
		territoryCode = ""
	}

	period, err := time.Parse("2006-01-02", periodStr+"-01")
	if err != nil {
		now := time.Now()
		period = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}

	var outlets []*model.TerritoryOutlet
	if territoryCode != "" {
		outlets, err = s.outlets.GetByTerritoryWithPerformance(ctx, territoryCode, period)
	} else {
		territories := []string{"DELHI-W", "DELHI-E", "DELHI-S", "DELHI-NW", "DELHI-C", "DELHI-NE", "DELHI-SW"}
		for _, t := range territories {
			o, er := s.outlets.GetByTerritoryWithPerformance(ctx, t, period)
			if er == nil {
				outlets = append(outlets, o...)
			}
		}
	}
	if err != nil {
		return nil, err
	}

	var aboveTarget, belowTarget int
	var totalAchievement float64
	for _, o := range outlets {
		totalAchievement += o.TotalAchievementPct
		if o.TotalAchievementPct >= 90 {
			aboveTarget++
		} else {
			belowTarget++
		}
	}

	avgAchievement := float64(0)
	if len(outlets) > 0 {
		avgAchievement = totalAchievement / float64(len(outlets))
	}

	return &model.TerritorySummary{
		Period:        periodStr,
		TerritoryCode: territoryCode,
		Summary: model.TerritoryStats{
			TotalOutlets:      len(outlets),
			AboveTarget:      aboveTarget,
			BelowTarget:      belowTarget,
			AvgAchievementPct: avgAchievement,
		},
		Outlets: outlets,
	}, nil
}

func (s *OutletService) ListOutlets(ctx context.Context, userID uuid.UUID) ([]*model.RetailOutlet, error) {
	claims, err := ExtractClaims(ctx)
	if err != nil {
		return nil, err
	}
	switch model.UserRole(claims.Role) {
	case model.RoleAdmin:
		return s.outlets.ListAll(ctx)
	case model.RoleTerritoryManager:
		if claims.TerritoryCode == nil {
			return nil, model.ErrForbidden
		}
		return s.outlets.ListByTerritory(ctx, *claims.TerritoryCode)
	default:
		if claims.TerritoryCode == nil {
			return nil, model.ErrForbidden
		}
		return s.outlets.ListByTerritory(ctx, *claims.TerritoryCode)
	}
}

// canAccess returns true if the claims allow viewing the given outlet.
func canAccess(claims *Claims, outlet *model.RetailOutlet) bool {
	if model.UserRole(claims.Role) == model.RoleAdmin {
		return true
	}
	if claims.TerritoryCode == nil || outlet.TerritoryCode == nil {
		return false
	}
	return *claims.TerritoryCode == *outlet.TerritoryCode
}
