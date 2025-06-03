package routes

import (
	"log/slog"

	"example.com/examples/api/layered/internal/handlers"
	"example.com/examples/api/layered/internal/services"
	"example.com/examples/api/layered/internal/telemetry"
)

// AddRoutes registers the API routes with the provided ServeMux.
//
// @title						Blog Service API
// @version						1.0
// @description					Practice Go API using the Standard Library and Postgres
// @termsOfService				http://swagger.io/terms/
// @contact.name				API Support
// @contact.url					http://www.swagger.io/support
// @contact.email				support@swagger.io
// @license.name				Apache 2.0
// @license.url					http://www.apache.org/licenses/LICENSE-2.0.html
// @host						localhost:8000
// @BasePath					/api
// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func AddRoutes(
	mux *telemetry.InstrumentedServeMux,
	logger *slog.Logger,
	usersService *services.UsersService,
) {
	// User endpoints
	mux.Handle("GET /api/user/{id}", handlers.HandleReadUser(logger, usersService))
	mux.Handle("GET /api/user", handlers.HandleListUsers(logger, usersService))
	mux.Handle("POST /api/user", handlers.HandleCreateUser(logger, usersService))
	mux.Handle("PUT /api/user/{id}", handlers.HandleUpdateUser(logger, usersService))
	mux.Handle("DELETE /api/user/{id}", handlers.HandleDeleteUser(logger, usersService))

	// Health check
	mux.Handle("GET /api/health", handlers.HandleHealthCheck(logger, usersService))
}
