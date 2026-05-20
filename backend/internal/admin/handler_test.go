package admin_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bpcl/portal-api/internal/admin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubRepo implements admin.RepositoryInterface for unit tests — no DB needed.
type stubRepo struct {
	dealers      []admin.Dealer
	periods      []admin.CompetitionPeriod
	targets      []admin.TargetRow
	ingestSummary []admin.IngestSummaryRow
	makgeRows    []admin.MAKGERow
	manualRows   []admin.ManualScoreRow

	toggledCC     string
	toggledActive bool
	createdName   string
	activatedMonth time.Time
	savedScores   []admin.ManualScoreRow

	toggleErr  error
	createErr  error
	activateErr error
	saveErr    error
}

func (s *stubRepo) ListDealers(ctx context.Context) ([]admin.Dealer, error) {
	return s.dealers, nil
}
func (s *stubRepo) ToggleDealer(ctx context.Context, ccCode string, active bool) error {
	s.toggledCC = ccCode
	s.toggledActive = active
	return s.toggleErr
}
func (s *stubRepo) ListPeriods(ctx context.Context) ([]admin.CompetitionPeriod, error) {
	return s.periods, nil
}
func (s *stubRepo) CreatePeriod(ctx context.Context, name string, monthYear time.Time) (*admin.CompetitionPeriod, error) {
	s.createdName = name
	if s.createErr != nil {
		return nil, s.createErr
	}
	return &admin.CompetitionPeriod{ID: "uuid-1", Name: name, MonthYear: monthYear, TotalSlots: 40, IsActive: false}, nil
}
func (s *stubRepo) ActivatePeriod(ctx context.Context, monthYear time.Time) error {
	s.activatedMonth = monthYear
	return s.activateErr
}
func (s *stubRepo) GetTargets(ctx context.Context, monthYear time.Time) ([]admin.TargetRow, error) {
	return s.targets, nil
}
func (s *stubRepo) GetIngestSummary(ctx context.Context, metric, table, col string, monthYear time.Time) ([]admin.IngestSummaryRow, error) {
	return s.ingestSummary, nil
}
func (s *stubRepo) GetMAKGEReadings(ctx context.Context, monthYear time.Time) ([]admin.MAKGERow, error) {
	return s.makgeRows, nil
}
func (s *stubRepo) GetManualScores(ctx context.Context, monthYear time.Time) ([]admin.ManualScoreRow, error) {
	return s.manualRows, nil
}
func (s *stubRepo) UpsertManualScores(ctx context.Context, monthYear time.Time, rows []admin.ManualScoreRow) error {
	s.savedScores = rows
	return s.saveErr
}
func (s *stubRepo) GetLastETLRun(ctx context.Context) (map[string]any, error) {
	return map[string]any{"status": "never_run", "message": "No ETL runs recorded"}, nil
}

func newHandler(repo *stubRepo) *admin.Handler {
	return admin.NewHandler(repo)
}

// ── Dealers ──────────────────────────────────────────────────────────────────

func TestListDealers_ReturnsJSON(t *testing.T) {
	repo := &stubRepo{
		dealers: []admin.Dealer{
			{CCCode: "112458", ROName: "Test RO", Area: "Central", IsActive: true},
		},
	}
	h := newHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/dealers", nil)
	rr := httptest.NewRecorder()
	h.ListDealers(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var out []admin.Dealer
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
	require.Len(t, out, 1)
	assert.Equal(t, "112458", out[0].CCCode)
}

func TestToggleDealer_CallsRepoWithCorrectArgs(t *testing.T) {
	repo := &stubRepo{}
	h := newHandler(repo)

	body := `{"is_active": false}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/dealers/112458", bytes.NewBufferString(body))
	req.SetPathValue("cc_code", "112458")
	rr := httptest.NewRecorder()
	h.ToggleDealer(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "112458", repo.toggledCC)
	assert.False(t, repo.toggledActive)
}

func TestToggleDealer_BadJSON_Returns400(t *testing.T) {
	repo := &stubRepo{}
	h := newHandler(repo)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/dealers/112458", bytes.NewBufferString("{bad"))
	req.SetPathValue("cc_code", "112458")
	rr := httptest.NewRecorder()
	h.ToggleDealer(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// ── Competition Periods ───────────────────────────────────────────────────────

func TestCreatePeriod_ValidRequest_Returns201(t *testing.T) {
	repo := &stubRepo{}
	h := newHandler(repo)

	body := `{"name": "Boost and Win May 2026", "month_year": "2026-05-01"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/competition-periods", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.CreatePeriod(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, "Boost and Win May 2026", repo.createdName)
}

func TestCreatePeriod_NonFirstOfMonth_Returns400(t *testing.T) {
	repo := &stubRepo{}
	h := newHandler(repo)

	body := `{"name": "Bad", "month_year": "2026-05-15"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/competition-periods", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.CreatePeriod(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestActivatePeriod_ValidMonthYear_Returns200(t *testing.T) {
	repo := &stubRepo{}
	h := newHandler(repo)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/competition-periods/2026-05-01/activate", nil)
	req.SetPathValue("month_year", "2026-05-01")
	rr := httptest.NewRecorder()
	h.ActivatePeriod(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, 2026, repo.activatedMonth.Year())
	assert.Equal(t, time.May, repo.activatedMonth.Month())
}

func TestActivatePeriod_InvalidMonthYear_Returns400(t *testing.T) {
	repo := &stubRepo{}
	h := newHandler(repo)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/competition-periods/not-a-date/activate", nil)
	req.SetPathValue("month_year", "not-a-date")
	rr := httptest.NewRecorder()
	h.ActivatePeriod(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// ── Targets ───────────────────────────────────────────────────────────────────

func TestGetTargets_ValidMonth_ReturnsRows(t *testing.T) {
	ms := 100.0
	repo := &stubRepo{
		targets: []admin.TargetRow{
			{CCCode: "112458", ROName: "Test RO", MSKL: &ms},
		},
	}
	h := newHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/targets/2026-05-01", nil)
	req.SetPathValue("month_year", "2026-05-01")
	rr := httptest.NewRecorder()
	h.GetTargets(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var out []admin.TargetRow
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
	assert.Len(t, out, 1)
}

// ── Ingest Summary ────────────────────────────────────────────────────────────

func TestIngestSummary_ValidMetric_ReturnsRows(t *testing.T) {
	repo := &stubRepo{
		ingestSummary: []admin.IngestSummaryRow{
			{CCCode: "112458", ROName: "Test RO", MTDSum: 250.5, DaysFilled: 10},
		},
	}
	h := newHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/ingest/summary?month=2026-05-01&metric=ms", nil)
	rr := httptest.NewRecorder()
	h.IngestSummary(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var out []admin.IngestSummaryRow
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
	assert.Len(t, out, 1)
}

func TestIngestSummary_UnknownMetric_Returns400(t *testing.T) {
	repo := &stubRepo{}
	h := newHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/ingest/summary?month=2026-05-01&metric=bogus", nil)
	rr := httptest.NewRecorder()
	h.IngestSummary(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// ── Manual Scores ─────────────────────────────────────────────────────────────

func TestSaveManualScores_ValidRequest_CallsRepo(t *testing.T) {
	repo := &stubRepo{}
	h := newHandler(repo)

	bonus := 7.5
	grade := "Good"
	reqBody := admin.BulkManualScoresRequest{
		MonthYear: "2026-05-01",
		Rows: []admin.ManualScoreRow{
			{CCCode: "112458", ROName: "Test RO", BonusMarks: &bonus, CleanlinessGrade: &grade},
		},
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/manual-scores", bytes.NewReader(b))
	rr := httptest.NewRecorder()
	h.SaveManualScores(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	require.Len(t, repo.savedScores, 1)
	assert.Equal(t, "112458", repo.savedScores[0].CCCode)
}

func TestSaveManualScores_InvalidMonthYear_Returns400(t *testing.T) {
	repo := &stubRepo{}
	h := newHandler(repo)

	reqBody := admin.BulkManualScoresRequest{MonthYear: "bad-date"}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/manual-scores", bytes.NewReader(b))
	rr := httptest.NewRecorder()
	h.SaveManualScores(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
