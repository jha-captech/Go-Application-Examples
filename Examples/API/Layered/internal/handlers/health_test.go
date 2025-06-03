package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"example.com/examples/api/layered/internal/services"
)

func TestHandleHealthCheck(t *testing.T) {
	tests := map[string]struct {
		name         string
		mockStatus   services.DeepHealthStatus
		mockErr      error
		wantStatus   int
		wantResponse healthResponse
	}{
		"healthy": {
			mockStatus: services.DeepHealthStatus{
				DB:    "ok",
				Cache: "ok",
			},
			mockErr:    nil,
			wantStatus: http.StatusOK,
			wantResponse: healthResponse{
				Status: "ok",
				Checks: services.DeepHealthStatus{DB: "ok", Cache: "ok"},
			},
		},
		"db unhealthy": {
			mockStatus: services.DeepHealthStatus{
				DB:    "unhealthy",
				Cache: "ok",
			},
			mockErr:    errors.New("db down"),
			wantStatus: http.StatusInternalServerError,
			wantResponse: healthResponse{
				Status: "unhealthy",
				Checks: services.DeepHealthStatus{DB: "unhealthy", Cache: "ok"},
			},
		},
		"cache unhealthy": {
			mockStatus: services.DeepHealthStatus{
				DB:    "ok",
				Cache: "unhealthy",
			},
			mockErr:    errors.New("cache down"),
			wantStatus: http.StatusInternalServerError,
			wantResponse: healthResponse{
				Status: "unhealthy",
				Checks: services.DeepHealthStatus{DB: "ok", Cache: "unhealthy"},
			},
		},
		"both unhealthy": {
			mockStatus: services.DeepHealthStatus{
				DB:    "unhealthy",
				Cache: "unhealthy",
			},
			mockErr:    errors.New("db and cache down"),
			wantStatus: http.StatusInternalServerError,
			wantResponse: healthResponse{
				Status: "unhealthy",
				Checks: services.DeepHealthStatus{DB: "unhealthy", Cache: "unhealthy"},
			},
		},
	}
	for name, tc := range tests {
		t.Run(
			name, func(t *testing.T) {
				// Create a new request
				req := httptest.NewRequest("GET", "/health", nil)

				// Create a new response recorder
				rec := httptest.NewRecorder()

				// Create a new ctxhandler
				logger := slog.Default()

				mockedUserHealth := &moqhealthChecker{
					DeepHealthCheckFunc: func(ctx context.Context) (services.DeepHealthStatus, error) {
						return tc.mockStatus, tc.mockErr
					},
				}

				// Call the handler
				HandleHealthCheck(logger, mockedUserHealth)(rec, req)

				// Check the status code
				assert.Equal(t, tc.wantStatus, rec.Code)

				// Check the body
				var resp healthResponse
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, tc.wantResponse, resp)
			},
		)
	}
}
