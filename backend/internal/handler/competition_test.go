package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/bpcl/portal-api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetLeaderboardHandler_Success(t *testing.T) {
	compSvc := &mockCompSvc{}
	h := newHandler(nil, nil, nil, nil, nil, compSvc, nil)

	tc := "DELHI-01"
	score := 56.37
	resp := &service.LeaderboardResponse{
		CompetitionID: uuid.New(),
		Name:          "Boost and Win March 2026",
		Period:        "2026-03",
		Entries: []service.LeaderboardEntry{
			{Rank: 1, CCNumber: "112847", OutletName: "M.L. SETHI", TotalScore: score, IsHighlighted: true},
		},
	}
	compSvc.On("GetLeaderboard", mock.Anything, "DELHI-01", mock.Anything).Return(resp, nil)

	r := httptest.NewRequest(http.MethodGet, "/competition/leaderboard?territory=DELHI-01", nil)
	r = injectClaims(r, "territory_manager", &tc)
	w := httptest.NewRecorder()

	h.GetLeaderboard(w, r)
	assert.Equal(t, http.StatusOK, w.Code)

	var got map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "Boost and Win March 2026", got["name"])
}

func TestGetLeaderboardHandler_MissingTerritory(t *testing.T) {
	compSvc := &mockCompSvc{}
	h := newHandler(nil, nil, nil, nil, nil, compSvc, nil)

	tc := "DELHI-01"
	r := httptest.NewRequest(http.MethodGet, "/competition/leaderboard", nil)
	r = injectClaims(r, "territory_manager", &tc)
	w := httptest.NewRecorder()

	h.GetLeaderboard(w, r)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestGetDealerScorecardHandler_Success(t *testing.T) {
	compSvc := &mockCompSvc{}
	h := newHandler(nil, nil, nil, nil, nil, compSvc, nil)

	tc := "DELHI-01"
	compID := uuid.New()
	total := 56.37
	rank := 1
	resp := &service.ScorecardResponse{
		CCNumber:      "112847",
		OutletName:    "M.L. SETHI",
		CompetitionID: compID,
		TotalScore:    &total,
		Rank:          &rank,
		Parameters:    make([]service.ScorecardParameter, 15),
	}
	compSvc.On("GetDealerScorecard", mock.Anything, "112847", compID, mock.Anything).Return(resp, nil)

	r := httptest.NewRequest(http.MethodGet, "/competition/"+compID.String()+"/dealers/112847", nil)
	r.SetPathValue("cc", "112847")
	r.SetPathValue("competition_id", compID.String())
	r = injectClaims(r, "territory_manager", &tc)
	w := httptest.NewRecorder()

	h.GetDealerScorecard(w, r)
	assert.Equal(t, http.StatusOK, w.Code)

	var got map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "112847", got["cc_number"])
}

func TestGetDealerScorecardHandler_InvalidCompetitionID(t *testing.T) {
	compSvc := &mockCompSvc{}
	h := newHandler(nil, nil, nil, nil, nil, compSvc, nil)

	tc := "DELHI-01"
	r := httptest.NewRequest(http.MethodGet, "/competition/bad-id/dealers/112847", nil)
	r.SetPathValue("cc", "112847")
	r.SetPathValue("competition_id", "not-a-uuid")
	r = injectClaims(r, "territory_manager", &tc)
	w := httptest.NewRecorder()

	h.GetDealerScorecard(w, r)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestGetDealerScorecardHandler_Forbidden(t *testing.T) {
	compSvc := &mockCompSvc{}
	h := newHandler(nil, nil, nil, nil, nil, compSvc, nil)

	tc := "DELHI-01"
	compID := uuid.New()
	compSvc.On("GetDealerScorecard", mock.Anything, "112847", compID, mock.Anything).Return(nil, model.ErrForbidden)

	r := httptest.NewRequest(http.MethodGet, "/competition/"+compID.String()+"/dealers/112847", nil)
	r.SetPathValue("cc", "112847")
	r.SetPathValue("competition_id", compID.String())
	r = injectClaims(r, "territory_manager", &tc)
	w := httptest.NewRecorder()

	h.GetDealerScorecard(w, r)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
