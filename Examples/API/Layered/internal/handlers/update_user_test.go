package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"example.com/examples/api/layered/internal/models"
)

func TestHandleUpdateUser(t *testing.T) {
	tests := map[string]struct {
		wantStatus int
		wantBody   models.User
		input      models.User
	}{
		"happy path": {
			wantStatus: 200,
			wantBody: models.User{
				ID:       1,
				Name:     "john",
				Email:    "john@mail.com",
				Password: "password123!",
			},
			input: models.User{
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
				reqBody, _ := json.Marshal(tc.input)
				req := httptest.NewRequest(http.MethodPut, "/users", bytes.NewBuffer(reqBody))
				req.SetPathValue("id", "1")

				// Create a new response recorder
				rec := httptest.NewRecorder()

				// Create a new ctxhandler
				logger := slog.Default()

				mockedUserUpdater := &moquserUpdater{
					UpdateUserFunc: func(
						_ context.Context,
						_ uint64,
						_ models.User,
					) (models.User, error) {
						return tc.wantBody, nil
					},
				}

				// Call the handler
				handler := HandleUpdateUser(logger, mockedUserUpdater)

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
