package app

import (
	"database/sql"
	"log/slog"
	"net/http"
)

func addRoutes(mux *http.ServeMux, logger *slog.Logger, db *sql.DB) {
	mux.Handle("GET /api/user/{id}", readUser(logger, db))
}
