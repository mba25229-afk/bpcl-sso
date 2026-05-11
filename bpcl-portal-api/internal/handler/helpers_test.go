package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/bpcl/portal-api/internal/handler"
	"github.com/bpcl/portal-api/internal/service"
	"github.com/google/uuid"
)

func newHandler(auth *mockAuthSvc, outlet *mockOutletSvc, perf *mockPerfSvc, target *mockTargetSvc, upload *mockUploadSvc, comp *mockCompSvc, marketShare *mockMarketShareSvc) *handler.Handler {
	users := &mockUserSvc{}
	return handler.New(auth, outlet, perf, target, upload, comp, marketShare, users, nil)
}

func newRequest(method, path string, body http.Handler) *http.Request {
	r := httptest.NewRequest(method, path, nil)
	return r
}

// injectClaims puts JWT claims into the request context (simulates middleware).
func injectClaims(r *http.Request, role string, tc *string) *http.Request {
	claims := &service.Claims{
		UserID:        uuid.New(),
		Role:          role,
		TerritoryCode: tc,
	}
	ctx := context.WithValue(r.Context(), service.ClaimsKey, claims)
	return r.WithContext(ctx)
}
