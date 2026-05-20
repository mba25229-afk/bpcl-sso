package portal_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bpcl/portal-api/internal/portal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubPortalRepo struct {
	card    *portal.ScoreCard
	err     error
	gotCC   string
	gotMonth time.Time
}

func (s *stubPortalRepo) GetScoreCard(ctx context.Context, ccCode string, monthYear time.Time) (*portal.ScoreCard, error) {
	s.gotCC = ccCode
	s.gotMonth = monthYear
	return s.card, s.err
}

func TestScoreCard_ValidRequest_ReturnsScoreCard(t *testing.T) {
	marks := 45.5
	repo := &stubPortalRepo{
		card: &portal.ScoreCard{
			Dealer:       portal.DealerInfo{CCCode: "112458", ROName: "Test RO", Area: "Central"},
			MonthYear:    "2026-05",
			TotalMarks:   marks,
			MaxPossible:  100,
			RankOverall:  3,
			TotalDealers: 40,
		},
	}
	h := portal.NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/portal/112458/scorecard?month=2026-05-01", nil)
	req.SetPathValue("cc_code", "112458")
	rr := httptest.NewRecorder()
	h.ScoreCard(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var out portal.ScoreCard
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
	assert.Equal(t, "112458", out.Dealer.CCCode)
	assert.Equal(t, 45.5, out.TotalMarks)
	assert.Equal(t, 3, out.RankOverall)
}

func TestScoreCard_PassesCCCodeToRepo(t *testing.T) {
	repo := &stubPortalRepo{
		card: &portal.ScoreCard{Dealer: portal.DealerInfo{CCCode: "999888"}},
	}
	h := portal.NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/portal/999888/scorecard?month=2026-05-01", nil)
	req.SetPathValue("cc_code", "999888")
	rr := httptest.NewRecorder()
	h.ScoreCard(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "999888", repo.gotCC)
}

func TestScoreCard_InvalidMonth_Returns400(t *testing.T) {
	repo := &stubPortalRepo{}
	h := portal.NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/portal/112458/scorecard?month=not-a-date", nil)
	req.SetPathValue("cc_code", "112458")
	rr := httptest.NewRecorder()
	h.ScoreCard(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestScoreCard_NoMonthParam_DefaultsToCurrentMonth(t *testing.T) {
	repo := &stubPortalRepo{
		card: &portal.ScoreCard{MonthYear: "2026-05"},
	}
	h := portal.NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/portal/112458/scorecard", nil)
	req.SetPathValue("cc_code", "112458")
	rr := httptest.NewRecorder()
	h.ScoreCard(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	// repo should have been called with day=1 of some month
	assert.Equal(t, 1, repo.gotMonth.Day())
}

func TestScoreCard_MetricsAndDailyTrend_ReturnedInJSON(t *testing.T) {
	ms := 500.0
	marks := 8.0
	repo := &stubPortalRepo{
		card: &portal.ScoreCard{
			Dealer:    portal.DealerInfo{CCCode: "112458"},
			MonthYear: "2026-05",
			Metrics: []portal.MetricCard{
				{
					MetricKey:   "ms_absolute_vol",
					DisplayName: "MS Volume",
					ActualValue: &ms,
					MarksScored: &marks,
					MaxMarks:    10,
					Unit:        "KL",
				},
			},
			DailyTrend: []portal.DailyPoint{
				{Date: "2026-05-01", MSKL: 20.5, HSDKL: 10.0, QOC: 2, UFill: 5},
			},
		},
	}
	h := portal.NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/portal/112458/scorecard?month=2026-05-01", nil)
	req.SetPathValue("cc_code", "112458")
	rr := httptest.NewRecorder()
	h.ScoreCard(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var out portal.ScoreCard
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
	require.Len(t, out.Metrics, 1)
	assert.Equal(t, "ms_absolute_vol", out.Metrics[0].MetricKey)
	require.Len(t, out.DailyTrend, 1)
	assert.Equal(t, "2026-05-01", out.DailyTrend[0].Date)
	assert.Equal(t, 20.5, out.DailyTrend[0].MSKL)
}
