package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http/httptest"
	"reflect"
	"testing"

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
				req := httptest.NewRequest("PUT", "/users", bytes.NewBuffer(reqBody))
				req.SetPathValue("id", "1")

				// Create a new response recorder
				rec := httptest.NewRecorder()

				// Create a new ctxlogger
				logger := slog.Default()

				mockedUserUpdater := &moquserUpdater{
					UpdateUserFunc: func(
						ctx context.Context,
						id uint64,
						user models.User,
					) (models.User, error) {
						return tc.wantBody, nil
					},
				}

				// Call the handler
				handler := HandleUpdateUser(logger, mockedUserUpdater)

				handler.ServeHTTP(rec, req)
				// Check the status code
				if rec.Code != tc.wantStatus {
					t.Errorf("want status %d, got %d", tc.wantStatus, rec.Code)
				}

				// Check the body
				var respBody models.User
				_ = json.Unmarshal(rec.Body.Bytes(), &respBody)

				if !reflect.DeepEqual(respBody, tc.wantBody) {
					t.Errorf("want body %q, got %q", tc.wantBody, rec.Body.String())
				}
			},
		)
	}
}
