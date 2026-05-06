package web

import (
	"encoding/json"
	"net/http"
)

type ErrorBody struct {
	Error APIError `json:"error"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func WriteJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, `{"error":{"code":"ENCODE_ERROR","message":"failed to encode response"}}`, http.StatusInternalServerError)
	}
}

func WriteError(w http.ResponseWriter, statusCode int, code string, message string, details any) {
	WriteJSON(w, statusCode, ErrorBody{
		Error: APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}
