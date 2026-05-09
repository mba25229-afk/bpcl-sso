package scoring

import (
	"context"
	"fmt"
	"time"
)

type Engine struct {
	actuals    *ActualsRepo
	paramsRepo *ParamsRepo
	scoreRepo  *ScoreRepository
}

func NewEngine(actuals *ActualsRepo, params *ParamsRepo, scores *ScoreRepository) *Engine {
	return &Engine{actuals: actuals, paramsRepo: params, scoreRepo: scores}
}

func (e *Engine) Run(ctx context.Context, monthYear, asOf time.Time) (int, error) {
	params, err := e.paramsRepo.FetchActive(ctx, monthYear)
	if err != nil {
		return 0, fmt.Errorf("load scoring params: %w", err)
	}
	if len(params) == 0 {
		return 0, fmt.Errorf("no active scoring params for %s", monthYear.Format("2006-01"))
	}

	actuals, err := e.actuals.FetchAll(ctx, monthYear, asOf)
	if err != nil {
		return 0, fmt.Errorf("fetch actuals: %w", err)
	}
	if len(actuals) == 0 {
		return 0, fmt.Errorf("no active dealers found")
	}

	n := 40

	var scoreRows []ScoreRow
	for _, param := range params {
		if !param.IsActive {
			continue
		}

		switch param.MetricKey {
		case "bonus":
			for _, a := range actuals {
				v := a.BonusMarks
				m := min(v, param.MaxMarks)
				scoreRows = append(scoreRows, ScoreRow{
					CCCode:      a.CCCode,
					MonthYear:   monthYear,
					MetricKey:   param.MetricKey,
					ActualValue: &v,
					MarksScored: &m,
					MaxMarks:    param.MaxMarks,
				})
			}

		case "cleanliness_audit":
			for _, a := range actuals {
				m := CleanlinessScore(a.CleanlinessGrade)
				scoreRows = append(scoreRows, ScoreRow{
					CCCode:      a.CCCode,
					MonthYear:   monthYear,
					MetricKey:   param.MetricKey,
					MarksScored: &m,
					MaxMarks:    param.MaxMarks,
				})
			}

		case "mak_ge_sales":
			ranks := Rank(actuals, param.MetricKey, n)
			for _, a := range actuals {
				v := a.MAKGESalesKL
				row := ScoreRow{
					CCCode:      a.CCCode,
					MonthYear:   monthYear,
					MetricKey:   param.MetricKey,
					ActualValue: &v,
					MaxMarks:    param.MaxMarks,
				}
				if a.MAKGEHasData {
					rank := ranks[a.CCCode]
					m := StandardScore(rank, n, param.MaxMarks)
					row.RankInGroup = &rank
					row.MarksScored = &m
				}
				scoreRows = append(scoreRows, row)
			}

		default:
			ranks := Rank(actuals, param.MetricKey, n)
			for _, a := range actuals {
				v := metricValue(a, param.MetricKey)
				rank := ranks[a.CCCode]
				var m float64
				if param.NegativeScaleEnabled {
					m = NegativeScaleScore(rank, n, param.MaxMarks)
				} else {
					m = StandardScore(rank, n, param.MaxMarks)
				}
				scoreRows = append(scoreRows, ScoreRow{
					CCCode:      a.CCCode,
					MonthYear:   monthYear,
					MetricKey:   param.MetricKey,
					ActualValue: &v,
					RankInGroup: &rank,
					MarksScored: &m,
					MaxMarks:    param.MaxMarks,
				})
			}
		}
	}

	if err := e.scoreRepo.BulkUpsert(ctx, scoreRows); err != nil {
		return 0, fmt.Errorf("write scores: %w", err)
	}
	return len(scoreRows), nil
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}