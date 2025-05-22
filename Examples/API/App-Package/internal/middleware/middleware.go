package middleware

import (
	"net/http"
	"slices"

	"go.opentelemetry.io/otel"
)

const name = "example.com/examples/api/api-package/internal/middleware"

var tracer = otel.Tracer(name)

type Func = func(next http.Handler) http.Handler

func WrapHandler(handler http.Handler, middlewares ...Func) http.Handler {
	if len(middlewares) <= 0 {
		return handler
	}

	next := handler

	for _, middleware := range slices.Backward(middlewares) {
		next = middleware(next)
	}

	return next
}
