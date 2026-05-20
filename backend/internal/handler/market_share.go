package handler

import (
	"context"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/google/uuid"
)

type MarketShareServiceI interface {
	IngestDelhiMaster(ctx context.Context, filePath string, uploadedFileID uuid.UUID) (int, int, error)
	GetMarketShareStatus(ctx context.Context, competitionID uuid.UUID) (*model.MarketShareStatusResponse, error)
	FullRecompute(ctx context.Context, competitionID uuid.UUID) (*model.RecomputeResponse, error)
}

func (h *Handler) HandleDelhiMasterUpload(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "file field required", "MISSING_FILE", "")
		return
	}
	defer file.Close()

	userID := userIDFromCtx(r)
	_ = userID

	uploadID := uuid.New()
	ext := filepath.Ext(header.Filename)
	filePath := filepath.Join("./uploads", "delhi_master_"+uploadID.String()+ext)

	if err := os.MkdirAll("./uploads", 0755); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create upload directory", "DIR_ERROR", "")
		return
	}

	dst, err := os.Create(filePath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save file", "FILE_SAVE_ERROR", "")
		return
	}
	defer dst.Close()

	if _, err := dst.ReadFrom(file); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save file", "FILE_SAVE_ERROR", "")
		return
	}

	go func() {
		h.MarketShare.IngestDelhiMaster(r.Context(), filePath, uploadID)
	}()

	writeJSON(w, http.StatusAccepted, map[string]interface{}{
		"upload_id": uploadID,
		"message":   "Delhi Master file uploaded. Processing in background.",
	})
}

func (h *Handler) GetMarketShareStatus(w http.ResponseWriter, r *http.Request) {
	competitionID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid competition id", "INVALID_ID", "")
		return
	}

	status, err := h.MarketShare.GetMarketShareStatus(r.Context(), competitionID)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, status)
}

func (h *Handler) RecomputeCompetitionScores(w http.ResponseWriter, r *http.Request) {
	competitionID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid competition id", "INVALID_ID", "")
		return
	}

	if r.URL.Query().Get("confirm") != "true" {
		writeError(w, http.StatusBadRequest, "add ?confirm=true to confirm this destructive operation", "CONFIRM_REQUIRED", "")
		return
	}

	result, err := h.MarketShare.FullRecompute(r.Context(), competitionID)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}
