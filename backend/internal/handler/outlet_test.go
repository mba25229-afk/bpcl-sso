package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetOutletHandler_Success(t *testing.T) {
	outletSvc := &mockOutletSvc{}
	h := newHandler(nil, outletSvc, nil, nil, nil, nil, nil)

	tc := "DELHI-01"
	outlet := &model.RetailOutlet{CCNumber: "123456", Name: "Test Outlet", TerritoryCode: &tc, OutletType: "regular"}

	outletSvc.On("GetOutlet", mock.Anything, "123456", mock.Anything).Return(outlet, nil)

	r := httptest.NewRequest(http.MethodGet, "/outlets/123456", nil)
	r.SetPathValue("cc", "123456")
	r = injectClaims(r, "territory_manager", &tc)
	w := httptest.NewRecorder()

	h.GetOutlet(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var got map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "123456", got["cc_number"])
}

func TestGetOutletHandler_NotFound(t *testing.T) {
	outletSvc := &mockOutletSvc{}
	h := newHandler(nil, outletSvc, nil, nil, nil, nil, nil)

	tc := "DELHI-01"
	outletSvc.On("GetOutlet", mock.Anything, "XXXXX", mock.Anything).Return(nil, model.ErrNotFound)

	r := httptest.NewRequest(http.MethodGet, "/outlets/XXXXX", nil)
	r.SetPathValue("cc", "XXXXX")
	r = injectClaims(r, "territory_manager", &tc)
	w := httptest.NewRecorder()

	h.GetOutlet(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetOutletHandler_Forbidden(t *testing.T) {
	outletSvc := &mockOutletSvc{}
	h := newHandler(nil, outletSvc, nil, nil, nil, nil, nil)

	tc := "DELHI-01"
	outletSvc.On("GetOutlet", mock.Anything, "123456", mock.Anything).Return(nil, model.ErrForbidden)

	r := httptest.NewRequest(http.MethodGet, "/outlets/123456", nil)
	r.SetPathValue("cc", "123456")
	r = injectClaims(r, "territory_manager", &tc)
	w := httptest.NewRecorder()

	h.GetOutlet(w, r)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestListOutletsHandler_Success(t *testing.T) {
	outletSvc := &mockOutletSvc{}
	h := newHandler(nil, outletSvc, nil, nil, nil, nil, nil)

	tc := "DELHI-01"
	outlets := []*model.RetailOutlet{
		{CCNumber: "123456", Name: "Outlet A", TerritoryCode: &tc, OutletType: "regular"},
	}
	outletSvc.On("ListOutlets", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(outlets, nil)

	r := httptest.NewRequest(http.MethodGet, "/outlets", nil)
	r = injectClaims(r, "territory_manager", &tc)
	w := httptest.NewRecorder()

	h.ListOutlets(w, r)
	assert.Equal(t, http.StatusOK, w.Code)

	var got []map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.Len(t, got, 1)
	assert.Equal(t, "Outlet A", got[0]["name"])
}

// uuid.UUID satisfies the AnythingOfType check only if we use the correct type name.
func init() {
	_ = uuid.Nil // ensure uuid package is used
}
