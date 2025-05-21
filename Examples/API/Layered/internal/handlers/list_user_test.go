package handlers

import (
	"context"
	"log/slog"
	"net/http/httptest"
	"testing"

	"example.com/examples/api/layered/handlers/mock"
	"example.com/examples/api/layered/models"
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

			// // Check the body
			// if strings.Trim(rec.Body.String(), "\n") != fmt.Sprintf("%+v", tc.wantBody) {
			// 	t.Errorf("want body %q, got %q", tc.wantBody, rec.Body.String())
			// }
		})
	}
}
