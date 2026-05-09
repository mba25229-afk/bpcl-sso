package ingest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bpcl/portal-api/internal/crystal/ingest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSvc is a service mock for handler tests.
type mockSvc struct {
	dailyResult  *ingest.IngestResult
	dailyErr     error
	makgeResult  *ingest.IngestResult
	makgeErr     error
	ratingResult *ingest.IngestResult
	ratingErr    error
	targetResult *ingest.IngestResult
	targetErr    error
}

func (m *mockSvc) IngestDaily(_ context.Context, _ ingest.BulkDailyRequest) (*ingest.IngestResult, error) {
	return m.dailyResult, m.dailyErr
}
func (m *mockSvc) IngestMAKGE(_ context.Context, _ ingest.MAKGERequest) (*ingest.IngestResult, error) {
	return m.makgeResult, m.makgeErr
}
func (m *mockSvc) IngestGoogleRating(_ context.Context, _ ingest.GoogleRatingRequest) (*ingest.IngestResult, error) {
	return m.ratingResult, m.ratingErr
}
func (m *mockSvc) IngestTargets(_ context.Context, _ ingest.BulkTargetsRequest) (*ingest.IngestResult, error) {
	return m.targetResult, m.targetErr
}

var _ ingest.ServiceInterface = (*mockSvc)(nil) // compile-time interface check

func okDaily() *ingest.IngestResult    { return &ingest.IngestResult{Inserted: 2} }
func okMAKGE() *ingest.IngestResult    { return &ingest.IngestResult{Inserted: 1} }
func okRating() *ingest.IngestResult   { return &ingest.IngestResult{Inserted: 1} }
func okTargets() *ingest.IngestResult  { return &ingest.IngestResult{Inserted: 3} }

// --- DailyBulk handler tests ---

func TestHandler_DailyBulk_ValidRequest_Returns200(t *testing.T) {
	svc := &mockSvc{dailyResult: okDaily()}
	h := ingest.NewHandler(svc)

	body := `{"date":"2026-05-07","metric":"ufill","rows":[{"cc_code":"112385","value":52}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ingest/daily-bulk", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.DailyBulk(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var res ingest.IngestResult
	require.NoError(t, json.NewDecoder(w.Body).Decode(&res))
	assert.Equal(t, 2, res.Inserted)
}

func TestHandler_DailyBulk_InvalidJSON_Returns400(t *testing.T) {
	h := ingest.NewHandler(&mockSvc{dailyResult: okDaily()})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not json"))
	w := httptest.NewRecorder()

	h.DailyBulk(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_DailyBulk_EmptyRows_Returns400(t *testing.T) {
	h := ingest.NewHandler(&mockSvc{dailyResult: okDaily()})

	body := `{"date":"2026-05-07","metric":"ufill","rows":[]}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.DailyBulk(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_DailyBulk_TooManyRows_Returns400(t *testing.T) {
	h := ingest.NewHandler(&mockSvc{dailyResult: okDaily()})

	rows := make([]ingest.DailyRow, 101)
	for i := range rows {
		rows[i] = ingest.DailyRow{CCCode: "112385", Value: float64(i)}
	}
	body, _ := json.Marshal(ingest.BulkDailyRequest{
		Date:   "2026-05-07",
		Metric: ingest.MetricUfill,
		Rows:   rows,
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.DailyBulk(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_DailyBulk_ServiceError_Returns422(t *testing.T) {
	h := ingest.NewHandler(&mockSvc{dailyErr: fmt.Errorf("date outside period")})

	body := `{"date":"2026-05-07","metric":"ufill","rows":[{"cc_code":"112385","value":1}]}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.DailyBulk(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// --- MAKGE handler tests ---

func TestHandler_MAKGE_ValidRequest_Returns200(t *testing.T) {
	h := ingest.NewHandler(&mockSvc{makgeResult: okMAKGE()})

	body := `{"cc_code":"112390","reading_date":"2026-05-08","meter_reading":14523.50}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.MAKGE(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_MAKGE_InvalidJSON_Returns400(t *testing.T) {
	h := ingest.NewHandler(&mockSvc{makgeResult: okMAKGE()})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{bad"))
	w := httptest.NewRecorder()

	h.MAKGE(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_MAKGE_ServiceError_Returns422(t *testing.T) {
	h := ingest.NewHandler(&mockSvc{makgeErr: fmt.Errorf("unknown cc_code")})

	body := `{"cc_code":"NONE","reading_date":"2026-05-08","meter_reading":100}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.MAKGE(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// --- GoogleRating handler tests ---

func TestHandler_GoogleRating_ValidRequest_Returns200(t *testing.T) {
	h := ingest.NewHandler(&mockSvc{ratingResult: okRating()})

	body := `{"cc_code":"112461","snapshot_date":"2026-05-08","rating":4.4,"review_count":1469}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.GoogleRating(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_GoogleRating_ServiceError_Returns422(t *testing.T) {
	h := ingest.NewHandler(&mockSvc{ratingErr: fmt.Errorf("rating out of range")})

	body := `{"cc_code":"112461","snapshot_date":"2026-05-08","rating":9.9,"review_count":1}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.GoogleRating(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// --- TargetsBulk handler tests ---

func TestHandler_TargetsBulk_ValidRequest_Returns200(t *testing.T) {
	h := ingest.NewHandler(&mockSvc{targetResult: okTargets()})

	body := `{"month_year":"2026-05-01","rows":[{"cc_code":"112385"}]}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.TargetsBulk(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_TargetsBulk_EmptyRows_Returns400(t *testing.T) {
	h := ingest.NewHandler(&mockSvc{targetResult: okTargets()})

	body := `{"month_year":"2026-05-01","rows":[]}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.TargetsBulk(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_TargetsBulk_ServiceError_Returns422(t *testing.T) {
	h := ingest.NewHandler(&mockSvc{targetErr: fmt.Errorf("invalid month_year")})

	body := `{"month_year":"2026-05-15","rows":[{"cc_code":"112385"}]}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.TargetsBulk(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}
