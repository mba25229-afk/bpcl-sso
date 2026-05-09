package dashboard_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bpcl/portal-api/internal/dashboard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubDashRepo struct {
	resp     *dashboard.DashboardResponse
	err      error
	gotMonth time.Time
}

func (s *stubDashRepo) GetDashboard(ctx context.Context, monthYear time.Time) (*dashboard.DashboardResponse, error) {
	s.gotMonth = monthYear
	return s.resp, s.err
}

// ── trafficLight logic ────────────────────────────────────────────────────────

func TestTrafficLight_Thresholds(t *testing.T) {
	cases := []struct {
		pct    float64
		expect string
	}{
		{95, "green"},
		{90, "green"},
		{89.9, "amber"},
		{70, "amber"},
		{69.9, "red"},
		{0, "red"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.expect, dashboard.TrafficLight(tc.pct),
			"pct=%.1f", tc.pct)
	}
}

// ── buildSummary ──────────────────────────────────────────────────────────────

func TestBuildSummary_CountsCorrectly(t *testing.T) {
	dealers := []dashboard.DashboardRow{
		{TotalMarks: 92, MaxPossible: 100, AchievePct: 92, Status: "green"},
		{TotalMarks: 80, MaxPossible: 100, AchievePct: 80, Status: "amber"},
		{TotalMarks: 60, MaxPossible: 100, AchievePct: 60, Status: "red"},
	}
	s := dashboard.BuildSummary(dealers)
	assert.Equal(t, 3, s.TotalDealers)
	assert.Equal(t, 1, s.GreenCount)
	assert.Equal(t, 1, s.AmberCount)
	assert.Equal(t, 1, s.RedCount)
	assert.InDelta(t, 92.0, s.TopScore, 0.001)
	assert.InDelta(t, (92.0+80.0+60.0)/3, s.AvgScore, 0.001)
}

func TestBuildSummary_Empty(t *testing.T) {
	s := dashboard.BuildSummary(nil)
	assert.Equal(t, 0, s.TotalDealers)
	assert.Equal(t, float64(0), s.AvgScore)
}

// ── handler ───────────────────────────────────────────────────────────────────

func TestGetDashboard_ValidMonth_ReturnsDashboard(t *testing.T) {
	now := time.Now().UTC()
	repo := &stubDashRepo{
		resp: &dashboard.DashboardResponse{
			MonthYear:  now.Format("2006-01"),
			ComputedAt: now,
			Dealers: []dashboard.DashboardRow{
				{CCCode: "112458", ROName: "Alpha RO", Rank: 1, TotalMarks: 90, MaxPossible: 100, AchievePct: 90, Status: "green"},
				{CCCode: "112459", ROName: "Beta RO", Rank: 2, TotalMarks: 75, MaxPossible: 100, AchievePct: 75, Status: "amber"},
			},
			Summary: dashboard.DashboardSummary{TotalDealers: 2, GreenCount: 1, AmberCount: 1},
		},
	}
	h := dashboard.NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard?month=2026-05-01", nil)
	rr := httptest.NewRecorder()
	h.GetDashboard(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var out dashboard.DashboardResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
	require.Len(t, out.Dealers, 2)
	assert.Equal(t, "112458", out.Dealers[0].CCCode)
	assert.Equal(t, 1, out.Dealers[0].Rank)
}

func TestGetDashboard_InvalidMonth_Returns400(t *testing.T) {
	repo := &stubDashRepo{}
	h := dashboard.NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard?month=bad", nil)
	rr := httptest.NewRecorder()
	h.GetDashboard(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetDashboard_NoMonth_DefaultsToCurrentMonth(t *testing.T) {
	now := time.Now().UTC()
	repo := &stubDashRepo{
		resp: &dashboard.DashboardResponse{MonthYear: now.Format("2006-01")},
	}
	h := dashboard.NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	rr := httptest.NewRecorder()
	h.GetDashboard(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, 1, repo.gotMonth.Day())
}

func TestGetDashboard_SummaryCountsMatchDealers(t *testing.T) {
	repo := &stubDashRepo{
		resp: &dashboard.DashboardResponse{
			Dealers: []dashboard.DashboardRow{
				{Status: "green", AchievePct: 91},
				{Status: "red", AchievePct: 55},
				{Status: "red", AchievePct: 40},
			},
			Summary: dashboard.DashboardSummary{TotalDealers: 3, GreenCount: 1, RedCount: 2},
		},
	}
	h := dashboard.NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard?month=2026-05-01", nil)
	rr := httptest.NewRecorder()
	h.GetDashboard(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var out dashboard.DashboardResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
	assert.Equal(t, 3, out.Summary.TotalDealers)
	assert.Equal(t, 1, out.Summary.GreenCount)
	assert.Equal(t, 2, out.Summary.RedCount)
	assert.Equal(t, out.Summary.GreenCount+out.Summary.AmberCount+out.Summary.RedCount, out.Summary.TotalDealers)
}
