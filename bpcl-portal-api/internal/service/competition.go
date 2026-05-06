package service

import (
	"context"
	"sort"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/google/uuid"
)

type LeaderboardEntry struct {
	Rank       int       `json:"rank"`
	CCNumber   string    `json:"cc_number"`
	OutletName string    `json:"outlet_name"`
	TotalScore float64   `json:"total_score"`
	IsHighlighted bool   `json:"is_highlighted"`
}

type LeaderboardResponse struct {
	CompetitionID uuid.UUID          `json:"competition_id"`
	Name          string             `json:"name"`
	Period        string             `json:"period"`
	Entries       []LeaderboardEntry `json:"entries"`
}

type ScorecardParameter struct {
	Name     string   `json:"name"`
	MaxMarks float64  `json:"max_marks"`
	Score    *float64 `json:"score"`
}

type ScorecardResponse struct {
	CCNumber      string               `json:"cc_number"`
	OutletName    string               `json:"outlet_name"`
	CompetitionID uuid.UUID            `json:"competition_id"`
	TotalScore    *float64             `json:"total_score"`
	Rank          *int                 `json:"rank"`
	Parameters    []ScorecardParameter `json:"parameters"`
}

type CompetitionService struct {
	competition CompetitionRepository
	outlets     OutletRepository
}

func NewCompetitionService(competition CompetitionRepository, outlets OutletRepository) *CompetitionService {
	return &CompetitionService{competition: competition, outlets: outlets}
}

func (s *CompetitionService) GetLeaderboard(ctx context.Context, territoryCode string, userID uuid.UUID) (*LeaderboardResponse, error) {
	period, err := s.competition.GetActivePeriod(ctx, territoryCode)
	if err != nil {
		return nil, err
	}

	scores, err := s.competition.GetScores(ctx, period.ID)
	if err != nil {
		return nil, err
	}

	// Build a map of cc_number → outlet for name lookup and type filter
	// Use ListAll so names resolve for cross-territory competitions
	outlets, err := s.outlets.ListAll(ctx)
	if err != nil {
		outlets = nil
	}
	outletMap := make(map[string]*model.RetailOutlet, len(outlets))
	for _, o := range outlets {
		outletMap[o.CCNumber] = o
	}

	// Filter: only regular outlets (Bug #5 fix)
	var filtered []*model.CompetitionScore
	for _, sc := range scores {
		o := outletMap[sc.CCNumber]
		if o != nil && o.OutletType != "regular" {
			continue
		}
		filtered = append(filtered, sc)
	}

	// Sort by total_score DESC
	sort.Slice(filtered, func(i, j int) bool {
		si := scoreValue(filtered[i].TotalScore)
		sj := scoreValue(filtered[j].TotalScore)
		return si > sj
	})

	entries := make([]LeaderboardEntry, 0, len(filtered))
	for i, sc := range filtered {
		rank := i + 1
		name := sc.CCNumber
		if o := outletMap[sc.CCNumber]; o != nil {
			name = o.Name
		}
		entries = append(entries, LeaderboardEntry{
			Rank:          rank,
			CCNumber:      sc.CCNumber,
			OutletName:    name,
			TotalScore:    scoreValue(sc.TotalScore),
			IsHighlighted: rank <= 10,
		})
	}

	return &LeaderboardResponse{
		CompetitionID: period.ID,
		Name:          period.Name,
		Period:        period.Period.Format("2006-01"),
		Entries:       entries,
	}, nil
}

func (s *CompetitionService) GetDealerScorecard(ctx context.Context, cc string, competitionID uuid.UUID, userID uuid.UUID) (*ScorecardResponse, error) {
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

	score, err := s.competition.GetDealerScore(ctx, competitionID, cc)
	if err != nil {
		return nil, err
	}

	params := []ScorecardParameter{
		{Name: "MS Volume (KL)", MaxMarks: 10, Score: score.ScoreMSVol},
		{Name: "MS Growth %", MaxMarks: 5, Score: score.ScoreMSGrowth},
		{Name: "MS Market Share Gain %", MaxMarks: 10, Score: score.ScoreMSTAGain},
		{Name: "HSD Volume (KL)", MaxMarks: 10, Score: score.ScoreHSDVol},
		{Name: "HSD Growth %", MaxMarks: 5, Score: score.ScoreHSDGrowth},
		{Name: "HSD Market Share Gain %", MaxMarks: 10, Score: score.ScoreHSDTAGain},
		{Name: "Oil Change Count", MaxMarks: 10, Score: score.ScoreQOC},
		{Name: "MAK/GE Sales (₹)", MaxMarks: 10, Score: score.ScoreLubricants},
		{Name: "UFill Transactions", MaxMarks: 5, Score: score.ScoreUFill},
		{Name: "Speed Volume (KL)", MaxMarks: 5, Score: score.ScoreSpeed},
		{Name: "Cleanliness Audit", MaxMarks: 5, Score: score.ScoreCleanliness},
		{Name: "Sangam Certifications", MaxMarks: 5, Score: score.ScoreSangam},
		{Name: "Google Rating", MaxMarks: 5, Score: score.ScoreGoogle},
		{Name: "IPS %", MaxMarks: 5, Score: score.ScoreIPS},
		{Name: "Bonus", MaxMarks: 10, Score: score.ScoreBonus},
	}

	return &ScorecardResponse{
		CCNumber:      cc,
		OutletName:    outlet.Name,
		CompetitionID: competitionID,
		TotalScore:    score.TotalScore,
		Rank:          score.Rank,
		Parameters:    params,
	}, nil
}

func scoreValue(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}
