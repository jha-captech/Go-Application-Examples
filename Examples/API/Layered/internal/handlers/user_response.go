package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// createUserResponse represents the response for creating a user.
type UserResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// encodeResponse encodes data as a JSON response.
func encodeResponse(w http.ResponseWriter, logger *slog.Logger, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Error while marshaling data", "err", err, "data", data)

		h := w.Header()
		h.Del("Content-Length")
		h.Set("Content-Type", "application/json; charset=utf-8")
		h.Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintln(w, `{"Error":"Internal server error"}`)
	}
}
