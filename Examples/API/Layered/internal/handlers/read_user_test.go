package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/examples/api/layered/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestHandleReadUser(t *testing.T) {
	tests := map[string]struct {
		wantStatus int
		wantBody   models.User
	}{
		"happy path": {
			wantStatus: 200,
			wantBody: models.User{
				ID:       1,
				Name:     "john",
				Email:    "john@mail.com",
				Password: "password123!",
			},
		},
	}
	for name, tc := range tests {
		t.Run(
			name, func(t *testing.T) {
				// Create a new request
				req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
				req.SetPathValue("id", "1")

				// Create a new response recorder
				rec := httptest.NewRecorder()

				// Create a new ctxhandler
				logger := slog.Default()

				mockedUserReader := &moquserReader{
					ReadUserFunc: func(ctx context.Context, id uint64) (models.User, error) {
						return tc.wantBody, nil
					},
				}
				// Call the handler
				handler := HandleReadUser(logger, mockedUserReader)

				handler.ServeHTTP(rec, req)
				// Check the status code
				assert.Equal(t, tc.wantStatus, rec.Code)

				// Check the body
				type usersResponse struct {
					Users []models.User `json:"Users"`
				}

				var resp usersResponse
				_ = json.Unmarshal(rec.Body.Bytes(), &resp)
				respBody := resp.Users

				assert.ObjectsAreEqualValues(tc.wantBody, respBody)
			},
		)
	}
}
