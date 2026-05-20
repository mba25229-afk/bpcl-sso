package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/bpcl/portal-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetPerformanceHandler_Success(t *testing.T) {
	perfSvc := &mockPerfSvc{}
	h := newHandler(nil, nil, perfSvc, nil, nil, nil, nil)

	tc := "DELHI-01"
	period := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	resp := &service.PerformanceResponse{
		CCNumber:   "123456",
		OutletName: "Test Outlet",
		Period:     "2026-03",
		Fuel:       []model.PerformanceSummary{{ProductCode: "MS"}},
		NonFuel:    []model.PerformanceSummary{},
	}

	perfSvc.On("GetPerformance", mock.Anything, "123456", period, mock.Anything).Return(resp, nil)

	r := httptest.NewRequest(http.MethodGet, "/outlets/123456/performance?period=2026-03", nil)
	r.SetPathValue("cc", "123456")
	r = injectClaims(r, "territory_manager", &tc)
	w := httptest.NewRecorder()

	h.GetPerformance(w, r)
	assert.Equal(t, http.StatusOK, w.Code)

	var got map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "123456", got["cc_number"])
}

func TestGetPerformanceHandler_MissingPeriod(t *testing.T) {
	perfSvc := &mockPerfSvc{}
	h := newHandler(nil, nil, perfSvc, nil, nil, nil, nil)

	tc := "DELHI-01"
	r := httptest.NewRequest(http.MethodGet, "/outlets/123456/performance", nil)
	r.SetPathValue("cc", "123456")
	r = injectClaims(r, "territory_manager", &tc)
	w := httptest.NewRecorder()

	h.GetPerformance(w, r)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestGetPerformanceHandler_Forbidden(t *testing.T) {
	perfSvc := &mockPerfSvc{}
	h := newHandler(nil, nil, perfSvc, nil, nil, nil, nil)

	tc := "DELHI-01"
	period := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	perfSvc.On("GetPerformance", mock.Anything, "123456", period, mock.Anything).Return(nil, model.ErrForbidden)

	r := httptest.NewRequest(http.MethodGet, "/outlets/123456/performance?period=2026-03", nil)
	r.SetPathValue("cc", "123456")
	r = injectClaims(r, "territory_manager", &tc)
	w := httptest.NewRecorder()

	h.GetPerformance(w, r)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
