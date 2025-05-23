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
	return func(w http.ResponseWriter, r *http.Request) {
		logger.InfoContext(r.Context(), "health check called")

		encodeResponse(w, logger, http.StatusOK, healthResponse{Status: "ok"})
	}
}
