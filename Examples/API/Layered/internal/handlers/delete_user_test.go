package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http/httptest"
	"testing"

	"example.com/examples/api/layered/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestHandleDeleteUser(t *testing.T) {
	tests := map[string]struct {
		wantStatus int
		wantBody   models.User
		input      models.User
	}{
		"happy path": {
			wantStatus: 200,
		},
	}
	for name, tc := range tests {
		t.Run(
			name, func(t *testing.T) {
				// Create a new request
				reqBody, _ := json.Marshal(tc.input)
				req := httptest.NewRequest("DELETE", "/users", bytes.NewBuffer(reqBody))
				req.SetPathValue("id", "1")

				// Create a new response recorder
				rec := httptest.NewRecorder()

				// Create a new ctxhandler
				logger := slog.Default()

				mockedUserDeleter := &moquserDeleter{
					DeleteUserFunc: func(ctx context.Context, id uint64) error {
						return nil
					},
				}

				// Call the handler
				handler := HandleDeleteUser(logger, mockedUserDeleter)

				handler.ServeHTTP(rec, req)
				// Check the status code
				assert.Equal(t, tc.wantStatus, rec.Code)

				// Check the body
				var respBody models.User
				_ = json.Unmarshal(rec.Body.Bytes(), &respBody)

				assert.ObjectsAreEqualValues(tc.wantBody, respBody)
			},
		)
	}
}
