package service

import (
	"context"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/google/uuid"
)

type TargetInput struct {
	Fuel    map[string]float64 `json:"fuel"`
	NonFuel map[string]float64 `json:"non_fuel"`
}

type TargetService struct {
	outlets OutletRepository
	targets TargetRepository
	audit   AuditRepository
}

func NewTargetService(outlets OutletRepository, targets TargetRepository, audit AuditRepository) *TargetService {
	return &TargetService{outlets: outlets, targets: targets, audit: audit}
}

func (s *TargetService) GetTargets(ctx context.Context, cc string, period time.Time, userID uuid.UUID) (*model.TargetsResponse, error) {
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

	tgts, err := s.targets.GetByPeriod(ctx, cc, period)
	if err != nil {
		return nil, err
	}
	return buildTargetsResponse(period, tgts), nil
}

func (s *TargetService) SetTargets(ctx context.Context, cc string, period time.Time, input TargetInput, userID uuid.UUID) (*model.TargetsResponse, error) {
	claims, err := ExtractClaims(ctx)
	if err != nil {
		return nil, err
	}
	role := model.UserRole(claims.Role)
	if role != model.RoleTerritoryManager && role != model.RoleAdmin {
		return nil, model.ErrForbidden
	}

	outlet, err := s.outlets.GetByCC(ctx, cc)
	if err != nil {
		return nil, err
	}
	if !canAccess(claims, outlet) {
		return nil, model.ErrForbidden
	}

	var batch []*model.Target
	addTargets := func(m map[string]float64) {
		for code, val := range m {
			productID := productCodeToID(code)
			if productID == 0 {
				continue
			}
			batch = append(batch, &model.Target{
				CCNumber:    cc,
				ProductID:   productID,
				Period:      period,
				TargetValue: val,
				SetBy:       &userID,
			})
		}
	}
	addTargets(input.Fuel)
	addTargets(input.NonFuel)

	if err := s.targets.UpsertBatch(ctx, batch); err != nil {
		return nil, err
	}

	_ = s.audit.Log(ctx, userID, "set_target", cc, map[string]any{
		"period": period.Format("2006-01"),
		"count":  len(batch),
	})

	tgts, err := s.targets.GetByPeriod(ctx, cc, period)
	if err != nil {
		return nil, err
	}
	return buildTargetsResponse(period, tgts), nil
}

func buildTargetsResponse(period time.Time, tgts []*model.Target) *model.TargetsResponse {
	resp := &model.TargetsResponse{
		Period:  period.Format("2006-01"),
		Fuel:    make(map[string]float64),
		NonFuel: make(map[string]float64),
	}
	for _, t := range tgts {
		prod, ok := productCatalog[t.ProductID]
		if !ok {
			continue
		}
		if prod.Category == model.CategoryFuel {
			resp.Fuel[prod.Code] = t.TargetValue
		} else {
			resp.NonFuel[prod.Code] = t.TargetValue
		}
	}
	return resp
}

func productCodeToID(code string) int16 {
	for id, p := range productCatalog {
		if p.Code == code {
			return id
		}
	}
	return 0
}
