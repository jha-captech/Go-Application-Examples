package routes

// import (
// 	"log/slog"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"example.com/examples/api/layered/internal/services"
// )

// func TestAddRoutes(t *testing.T) {
// 	mux := http.NewServeMux()
// 	logger := slog.Default()
// 	usersService := &services.UsersService{} // Use a mock if needed

// 	// Replace handlers with mocks for isolation if needed.
// 	// For now, this test just checks that routes are registered and return 200.
// 	AddRoutes(mux, logger, usersService)

// 	tests := []struct {
// 		method string
// 		path   string
// 	}{
// 		{"GET", "/api/user/1"},
// 		{"GET", "/api/user"},
// 		{"POST", "/api/user"},
// 		{"PUT", "/api/user/1"},
// 		{"DELETE", "/api/user/1"},
// 		{"GET", "/api/health"},
// 	}

// 	for _, tt := range tests {
// 		req := httptest.NewRequest(tt.method, tt.path, nil)
// 		rr := httptest.NewRecorder()
// 		mux.ServeHTTP(rr, req)

// 		if rr.Code == http.StatusNotFound {
// 			t.Errorf("Route %s %s not registered (got 404)", tt.method, tt.path)
// 		}
// 	}
// }
