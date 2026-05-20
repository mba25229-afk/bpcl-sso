package handler

import (
	"net/http"
	"strconv"
)

func (h *Handler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	cc := r.PathValue("cc")
	if cc == "" {
		writeError(w, http.StatusUnprocessableEntity, "cc path param required", "MISSING_CC", "")
		return
	}

	period, err := parsePeriod(r.FormValue("period"))
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "period must be YYYY-MM", "INVALID_PERIOD", "")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "file field required", "MISSING_FILE", "")
		return
	}
	defer file.Close()

	userID := userIDFromCtx(r)
	resp, err := h.Upload.HandleUpload(r.Context(), file, header, cc, period, userID)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, resp)
}

func (h *Handler) GetUploadHistory(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	cc := q.Get("cc")
	limit := intParam(q.Get("limit"), 20)
	offset := intParam(q.Get("offset"), 0)

	userID := userIDFromCtx(r)
	resp, err := h.Upload.GetHistory(r.Context(), userID, cc, limit, offset)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func intParam(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 0 {
		return def
	}
	return v
}
