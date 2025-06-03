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
		mockStatus   []services.HealthStatus
		mockErr      error
		wantStatus   int
		wantResponse healthResponse
	}{
		"healthy": {
			mockStatus: []services.HealthStatus{
				{Name: "db", Status: "up"},
				{Name: "cache", Status: "up"},
			},
			mockErr:    nil,
			wantStatus: http.StatusOK,
			wantResponse: healthResponse{
				Status: "up",
				HealthDetails: []services.HealthStatus{
					{Name: "db", Status: "up"},
					{Name: "cache", Status: "up"},
				},
			},
		},
		"db unhealthy": {
			mockStatus: []services.HealthStatus{
				{Name: "db", Status: "unhealthy"},
				{Name: "cache", Status: "up"},
			},
			mockErr:    errors.New("db down"),
			wantStatus: http.StatusInternalServerError,
			wantResponse: healthResponse{
				Status: "unhealthy",
				HealthDetails: []services.HealthStatus{
					{Name: "db", Status: "unhealthy"},
					{Name: "cache", Status: "up"},
				},
			},
		},
		"cache unhealthy": {
			mockStatus: []services.HealthStatus{
				{Name: "db", Status: "up"},
				{Name: "cache", Status: "unhealthy"},
			},
			mockErr:    errors.New("cache down"),
			wantStatus: http.StatusInternalServerError,
			wantResponse: healthResponse{
				Status: "unhealthy",
				HealthDetails: []services.HealthStatus{
					{Name: "db", Status: "up"},
					{Name: "cache", Status: "unhealthy"},
				},
			},
		},
		"both unhealthy": {
			mockStatus: []services.HealthStatus{
				{Name: "db", Status: "unhealthy"},
				{Name: "cache", Status: "unhealthy"},
			},
			mockErr:    errors.New("db and cache down"),
			wantStatus: http.StatusInternalServerError,
			wantResponse: healthResponse{
				Status: "unhealthy",
				HealthDetails: []services.HealthStatus{
					{Name: "db", Status: "unhealthy"},
					{Name: "cache", Status: "unhealthy"},
				},
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
					DeepHealthCheckFunc: func(ctx context.Context) ([]services.HealthStatus, error) {
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
