package app

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
)

func NewHandler(logger *slog.Logger, db *sqlx.DB) http.Handler {
	mux := http.NewServeMux()

	addRoutes(mux, logger, db)

	return mux
}

// encodeResponse encodes data as a JSON response.
// It is important to note that once w.WriteHeader is called, the response headers are sent.
// Any subsequent calls to w.WriteHeader will have no effect.
func encodeResponse(w http.ResponseWriter, status int, data any) (error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return fmt.Errorf("failed to encode response: %w", err)
	}
	return nil
}
