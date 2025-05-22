package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http/httptest"
	"reflect"
	"testing"

	"example.com/examples/api/layered/internal/handlers/mock"
	"example.com/examples/api/layered/internal/models"
)

func TestHandleListUser(t *testing.T) {
	tests := map[string]struct {
		wantStatus int
		wantBody   []models.User
	}{
		"happy path": {
			wantStatus: 200,
			wantBody: []models.User{
				{
					ID:       1,
					Name:     "john",
					Email:    "john@mail.com",
					Password: "password123!",
				},
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Create a new request
			req := httptest.NewRequest("GET", "/users", nil)

			// Create a new response recorder
			rec := httptest.NewRecorder()

			// Create a new logger
			logger := slog.Default()

			userLister := new(mock.UsersLister)
			userLister.On("ListUsers", context.Background(), "").Return(tc.wantBody, nil)

			// Call the handler
			handler := HandleListUsers(logger, userLister)

			handler.ServeHTTP(rec, req)
			// Check the status code
			if rec.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, rec.Code)
			}

			// Check the body
			type usersResponse struct {
				Users []models.User `json:"Users"`
			}

			// Check the body
			var resp usersResponse
			_ = json.Unmarshal(rec.Body.Bytes(), &resp)
			respBody := resp.Users

			if !reflect.DeepEqual(respBody, tc.wantBody) {
				t.Errorf("want body %q, got %q", tc.wantBody, rec.Body.String())
			}
		})
	}
}
