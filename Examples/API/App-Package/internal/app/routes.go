package app

import (
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
)

// addRoutes registers all HTTP API routes for user operations to the provided ServeMux.
// It wires each route to its corresponding handler, passing the logger and database connection.
func addRoutes(mux *http.ServeMux, logger *slog.Logger, db *sqlx.DB) {
	mux.Handle("GET /api/user/{id}", readUser(logger, db))
	mux.Handle("POST /api/user", createUser(logger, db))
	mux.Handle("PUT /api/user/{id}", updateUser(logger, db))
	mux.Handle("DELETE /api/user/{id}", deleteUser(logger, db))
	mux.Handle("GET /api/users", listUsers(logger, db))
}
