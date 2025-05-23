package app

import (
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
)

func NewHandler(logger *slog.Logger, db *sqlx.DB) http.Handler {
	mux := http.NewServeMux()

	addRoutes(mux, logger, db)

	return mux
}
