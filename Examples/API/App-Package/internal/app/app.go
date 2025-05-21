package app

import (
	"database/sql"
	"log/slog"
	"net/http"
)

func NewHandler(logger *slog.Logger, db *sql.DB) http.Handler {
	mux := http.NewServeMux()

	addRoutes(mux, logger, db)

	return mux
}
