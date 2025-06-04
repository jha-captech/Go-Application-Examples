package app

import (
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
)

// HealthStatus represents the status of a dependency.
type healthStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// healthResponse represents the response for the health check.
type healthResponse struct {
	Status        string         `json:"status"`
	HealthDetails []healthStatus `json:"details"`
}

// HandleHealthCheck handles the deep health check endpoint.
// The only dependency checked is the database connection.
// In an enterprise application, this could be extended to include other dependencies like caches, message queues, etc.
//
//	@Summary		Health Check
//	@Description	Health Check endpoint
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	healthResponse
//	@Failure		500		{object}	healthResponse
//	@Router			/health	[GET]
func HandleHealthCheck(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	const name = "app.HandleHealthCheck"
	logger = logger.With(slog.String("func", name))

	const (
		healthyStatus   = "healthy"
		unhealthyStatus = "unhealthy"
	)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		logger = logger.With(getTraceIDAsAttr(ctx))
		logger.InfoContext(ctx, "health check called")

		if err := db.PingContext(ctx); err != nil {
			logger.ErrorContext(ctx, "health check failed", slog.String("error", err.Error()))
			_ = encodeResponseJSON(w, http.StatusInternalServerError, healthResponse{
				Status: unhealthyStatus,
				HealthDetails: []healthStatus{
					{
						Name:   "db",
						Status: unhealthyStatus,
					},
				},
			})

			return
		}

		_ = encodeResponseJSON(w, http.StatusOK, healthResponse{
			Status: healthyStatus,
			HealthDetails: []healthStatus{
				{
					Name:   "db",
					Status: healthyStatus,
				},
			},
		},
		)
	}
}
