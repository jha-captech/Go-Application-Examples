package app

import (
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "example.com/examples/api/app-package/cmd/api/docs" // import for Swagger docs generation
)

// addRoutes registers all HTTP API routes for user operations to the provided ServeMux.
// It wires each route to its corresponding handler, passing the logger and database connection.
func addRoutes(mux *http.ServeMux, logger *slog.Logger, db *sqlx.DB, enableSwagger bool) {
	mux.Handle("GET /api/user/{id}", readUser(logger, db))
	mux.Handle("POST /api/user", createUser(logger, db))
	mux.Handle("PUT /api/user/{id}", updateUser(logger, db))
	mux.Handle("DELETE /api/user/{id}", deleteUser(logger, db))
	mux.Handle("GET /api/user", listUsers(logger, db))

	mux.Handle("GET /health", HandleHealthCheck(logger, db))

	if enableSwagger {
		// Swagger docs
		mux.Handle(
			"GET /swagger/",
			httpSwagger.Handler(httpSwagger.URL("http://localhost:8080/swagger/doc.json")),
		)
	}

	// catch-all route for 404 Not Found with problem detail
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {		
		_ = encodeResponseJSON(w, http.StatusNotFound, problemDetail{
					Title:   "Path Not Found",
					Status:  http.StatusNotFound,
					Detail:  "The requested path does not exist.",
					TraceID: getTraceID(r.Context()),
				})
	})
}
