package handler

import (
	"encoding/json"
	"net/http"
)

type errorBody struct {
	Error     string `json:"error"`
	Code      string `json:"code,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg, code, requestID string) {
	writeJSON(w, status, errorBody{Error: msg, Code: code, RequestID: requestID})
}
