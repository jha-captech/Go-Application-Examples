package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel/codes"

	"example.com/examples/api/layered/internal/services"
)

// userCreator represents a type capable of reading a user from storage and
// returning it or an error.
type healthChecker interface {
	DeepHealthCheck(ctx context.Context) ([]services.HealthStatus, error)
}

// healthResponse represents the response for the health check.
type healthResponse struct {
	Status        string                  `json:"status"`
	HealthDetails []services.HealthStatus `json:"details"`
}

// HandleHealthCheck handles the deep health check endpoint.
//
//	@Summary		Health Check
//	@Description	Health Check endpoint
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	healthResponse
//	@Failure		500		{object}	healthResponse
//	@Router			/health	[GET]
func HandleHealthCheck(logger *slog.Logger, userHealth healthChecker) http.HandlerFunc {
	const name = "handlers.HandleHealthCheck"
	logger = logger.With(slog.String("func", name))

	return func(w http.ResponseWriter, r *http.Request) {
		const upStatus = "up"
		ctx, span := tracer.Start(r.Context(), name)
		defer span.End()
		logger.InfoContext(ctx, "health check called")

		checks, err := userHealth.DeepHealthCheck(ctx)
		status := upStatus
		code := http.StatusOK
		for _, check := range checks {
			if check.Status != upStatus {
				status = "unhealthy"
				code = http.StatusInternalServerError
				break
			}
		}
		if err != nil {
			logger.ErrorContext(ctx, "health check failed", slog.String("error", err.Error()))
			span.SetStatus(codes.Error, "health check failed")
			span.RecordError(err)
		}

		_ = encodeResponseJSON(
			w, code, healthResponse{
				Status:        status,
				HealthDetails: checks,
			},
		)
	}
}
