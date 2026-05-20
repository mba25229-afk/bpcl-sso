package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetTargetsHandler_Success(t *testing.T) {
	targetSvc := &mockTargetSvc{}
	h := newHandler(nil, nil, nil, targetSvc, nil, nil, nil)

	tc := "DELHI-01"
	period := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	resp := &model.TargetsResponse{
		Period:  "2026-03",
		Fuel:    map[string]float64{"MS": 500},
		NonFuel: map[string]float64{},
	}
	targetSvc.On("GetTargets", mock.Anything, "123456", period, mock.Anything).Return(resp, nil)

	r := httptest.NewRequest(http.MethodGet, "/outlets/123456/targets?period=2026-03", nil)
	r.SetPathValue("cc", "123456")
	r = injectClaims(r, "territory_manager", &tc)
	w := httptest.NewRecorder()

	h.GetTargets(w, r)
	assert.Equal(t, http.StatusOK, w.Code)

	var got map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "2026-03", got["period"])
}

func TestSetTargetsHandler_Success(t *testing.T) {
	targetSvc := &mockTargetSvc{}
	h := newHandler(nil, nil, nil, targetSvc, nil, nil, nil)

	tc := "DELHI-01"
	period := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	body := `{"fuel":{"MS":600},"non_fuel":{"QOC":50}}`

	resp := &model.TargetsResponse{
		Period:  "2026-03",
		Fuel:    map[string]float64{"MS": 600},
		NonFuel: map[string]float64{"QOC": 50},
	}
	targetSvc.On("SetTargets", mock.Anything, "123456", period, mock.Anything, mock.Anything).Return(resp, nil)

	r := httptest.NewRequest(http.MethodPut, "/outlets/123456/targets?period=2026-03", bytes.NewBufferString(body))
	r.Header.Set("Content-Type", "application/json")
	r.SetPathValue("cc", "123456")
	r = injectClaims(r, "territory_manager", &tc)
	w := httptest.NewRecorder()

	h.SetTargets(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSetTargetsHandler_Forbidden(t *testing.T) {
	targetSvc := &mockTargetSvc{}
	h := newHandler(nil, nil, nil, targetSvc, nil, nil, nil)

	tc := "DELHI-01"
	period := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	body := `{"fuel":{"MS":600},"non_fuel":{}}`

	targetSvc.On("SetTargets", mock.Anything, "123456", period, mock.Anything, mock.Anything).Return(nil, model.ErrForbidden)

	r := httptest.NewRequest(http.MethodPut, "/outlets/123456/targets?period=2026-03", bytes.NewBufferString(body))
	r.Header.Set("Content-Type", "application/json")
	r.SetPathValue("cc", "123456")
	r = injectClaims(r, "ro_manager", &tc)
	w := httptest.NewRecorder()

	h.SetTargets(w, r)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
