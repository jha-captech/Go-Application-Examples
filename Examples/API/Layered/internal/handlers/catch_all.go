package handlers

import (
	"net/http"

	"example.com/examples/api/layered/internal/middleware"
)

// HandleCatchAll handles all unmatched routes and returns a 404 Not Found response.
func HandleCatchAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = encodeResponseJSON(w, http.StatusNotFound, ProblemDetail{
			Title:   "Path Not Found",
			Status:  http.StatusNotFound,
			Detail:  "The requested path does not exist.",
			TraceID: middleware.GetTraceID(r.Context()),
		})
	}

}
