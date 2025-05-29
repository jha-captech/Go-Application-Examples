package handlers

import (
	"log/slog"
	"net/http"
)

// healthResponse represents the response for the health check.
type healthResponse struct {
	Status string `json:"status"`
}

// HandleHealthCheck handles the health check endpoint
//
//	@Summary		Health Check
//	@Description	Health Check endpoint
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	healthResponse
//	@Router			/health	[GET]
func HandleHealthCheck(logger *slog.Logger) http.HandlerFunc {
	const name = "handlers.HandleHealthCheck"
	logger = logger.With(slog.String("func", name))

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), name)
		defer span.End()
		logger.InfoContext(ctx, "health check called")

		_ = encodeResponse(w, http.StatusOK, healthResponse{Status: "ok"})
	}
}
