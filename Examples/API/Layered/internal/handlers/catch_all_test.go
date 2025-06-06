package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleCatchAll(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantTitle  string
		wantDetail string
	}{
		{
			name:       "GET unknown path",
			method:     http.MethodGet,
			path:       "/does-not-exist",
			wantStatus: http.StatusNotFound,
			wantTitle:  "Path Not Found",
			wantDetail: "The requested path does not exist.",
		},
		{
			name:       "POST unknown path",
			method:     http.MethodPost,
			path:       "/random",
			wantStatus: http.StatusNotFound,
			wantTitle:  "Path Not Found",
			wantDetail: "The requested path does not exist.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()

			handler := HandleCatchAll()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.wantStatus, rr.Code, "HTTP status code mismatch")

			var pd ProblemDetail
			if err := json.NewDecoder(rr.Body).Decode(&pd); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			assert.Equal(t, tc.wantTitle, pd.Title, "Problem title mismatch")
			assert.Equal(t, tc.wantDetail, pd.Detail, "Problem detail mismatch")
			assert.Equal(t, tc.wantStatus, pd.Status, "Problem status mismatch")
		})
	}
}
