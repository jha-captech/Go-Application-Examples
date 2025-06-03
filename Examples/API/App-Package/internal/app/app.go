package app

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
)

// NewHandler creates and returns a new HTTP handler with all application routes registered.
// It takes a logger and a database connection as dependencies.
func NewHandler(logger *slog.Logger, db *sqlx.DB) http.Handler {
	mux := http.NewServeMux()

	addRoutes(mux, logger, db)

	return mux
}

// encodeResponseJSON encodes the provided data as a JSON response and writes it to the ResponseWriter.
// It sets the Content-Type header to "application/json" and writes the specified HTTP status code.
// Returns an error if encoding fails.
func encodeResponseJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return fmt.Errorf("failed to encode response: %w", err)
	}

	return nil
}
