package app

import (
	"database/sql"
	"log/slog"
	"net/http"
)

func addRoutes(mux *http.ServeMux, logger *slog.Logger, db *sql.DB) {
	mux.Handle("GET /api/user/{id}", readUser(logger, db))
	mux.Handle("POST /api/user", createUser(logger, db))
	mux.Handle("PUT /api/user/{id}", updateUser(logger, db))
	mux.Handle("DELETE /api/user/{id}", deleteUser(logger, db))
	mux.Handle("GET /api/users", listUsers(logger, db))
}
