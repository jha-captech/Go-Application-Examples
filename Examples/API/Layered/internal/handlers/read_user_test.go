package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/examples/api/layered/internal/handlers/mock"
	"example.com/examples/api/layered/internal/models"
)

func TestHandleReadUser(t *testing.T) {
	tests := map[string]struct {
		wantStatus  int
		wantBody    string
		wantResults models.User
	}{
		"happy path": {
			wantStatus: 200,
			wantBody: `
				"id":       1,
				"name":     "john",
				"email":    "john@mail.com",
				"password": "password123!",
			`,
			wantResults: models.User{
				ID:       1,
				Name:     "john",
				Email:    "john@mail.com",
				Password: "password123!",
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Create a new request
			req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
			req.SetPathValue("id", "1")

			// Create a new response recorder
			rec := httptest.NewRecorder()

			// Create a new logger
			logger := slog.Default()

			userReader := new(mock.UserReader)
			userReader.On("ReadUser", context.Background(), uint64(1)).Return(tc.wantResults, nil)
			// Call the handler
			handler := HandleReadUser(logger, userReader)

			handler.ServeHTTP(rec, req)
			// Check the status code
			if rec.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, rec.Code)
			}

			// Check the body
			json, _ := json.Marshal(tc.wantResults)
			if strings.Trim(rec.Body.String(), "\n") != string(json) {
				t.Errorf("want body %q, got %q", string(json), rec.Body.String())
			}
		})
	}
}
