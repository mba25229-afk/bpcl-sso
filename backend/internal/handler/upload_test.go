package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bpcl/portal-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetUploadHistoryHandler_Success(t *testing.T) {
	uploadSvc := &mockUploadSvc{}
	h := newHandler(nil, nil, nil, nil, uploadSvc, nil, nil)

	tc := "DELHI-01"
	resp := &service.UploadListResponse{
		Items:  nil,
		Total:  0,
		Limit:  20,
		Offset: 0,
	}
	uploadSvc.On("GetHistory", mock.Anything, mock.AnythingOfType("uuid.UUID"), "", 20, 0).Return(resp, nil)

	r := httptest.NewRequest(http.MethodGet, "/uploads", nil)
	r = injectClaims(r, "territory_manager", &tc)
	w := httptest.NewRecorder()

	h.GetUploadHistory(w, r)
	assert.Equal(t, http.StatusOK, w.Code)

	var got map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, float64(0), got["total"])
}

func TestGetUploadHistoryHandler_WithParams(t *testing.T) {
	uploadSvc := &mockUploadSvc{}
	h := newHandler(nil, nil, nil, nil, uploadSvc, nil, nil)

	tc := "DELHI-01"
	resp := &service.UploadListResponse{
		Items:  nil,
		Total:  1,
		Limit:  10,
		Offset: 5,
	}
	uploadSvc.On("GetHistory", mock.Anything, mock.AnythingOfType("uuid.UUID"), "123456", 10, 5).Return(resp, nil)

	r := httptest.NewRequest(http.MethodGet, "/uploads?cc=123456&limit=10&offset=5", nil)
	r = injectClaims(r, "territory_manager", &tc)
	w := httptest.NewRecorder()

	h.GetUploadHistory(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
}
