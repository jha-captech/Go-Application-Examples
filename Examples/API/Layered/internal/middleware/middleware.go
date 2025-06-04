package middleware

import (
	"net/http"
	"slices"
)

// Func is a middleware function that wraps an http.Handler.
type Func = func(next http.Handler) http.Handler

func WrapHandler(handler http.Handler, middlewares ...Func) http.Handler {
	if len(middlewares) == 0 {
		return handler
	}

	next := handler

	for _, middleware := range slices.Backward(middlewares) {
		next = middleware(next)
	}

	return next
}
