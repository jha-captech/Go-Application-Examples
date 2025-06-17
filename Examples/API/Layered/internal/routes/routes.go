package routes

import (
	"log/slog"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	// Import the generated Swagger docs
	_ "example.com/examples/api/layered/cmd/api/docs"
	"example.com/examples/api/layered/internal/handlers"
	"example.com/examples/api/layered/internal/services"
)

// endpointMapper is an interface that defines a method for registering HTTP handlers
type endpointMapper interface {
	Handle(pattern string, handler http.Handler)
}

// AddRoutes registers the API routes with the provided ServeMux.
//
//	@title						Blog Service API
//	@version					1.0
//	@description				Practice Go API using the Standard Library and Postgres
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.url				http://www.swagger.io/support
//	@contact.email				support@swagger.io
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:8080
//	@BasePath					/api
//	@externalDocs.description	OpenAPI
//	@externalDocs.url			https://swagger.io/resources/open-api/
func AddRoutes(
	mux endpointMapper,
	logger *slog.Logger,
	usersService *services.UsersService,
	swaggerEnabled bool,
) {
	// User endpoints
	mux.Handle("GET /api/user/{id}", handlers.HandleReadUser(logger, usersService))
	mux.Handle("GET /api/user", handlers.HandleListUsers(logger, usersService))
	mux.Handle("POST /api/user", handlers.HandleCreateUser(logger, usersService))
	mux.Handle("PUT /api/user/{id}", handlers.HandleUpdateUser(logger, usersService))
	mux.Handle("DELETE /api/user/{id}", handlers.HandleDeleteUser(logger, usersService))

	// Health check
	mux.Handle("GET /api/health", handlers.HandleHealthCheck(logger, usersService))

	if swaggerEnabled {
		// Swagger docs
		mux.Handle(
			"GET /swagger/",
			httpSwagger.Handler(httpSwagger.URL("http://localhost:8080/swagger/doc.json")),
		)
	}

	// Catch-all route for 404 Not Found with problem detail
	mux.Handle("/", handlers.HandleCatchAll())
}
