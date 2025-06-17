package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel"
)

const name = "example.com/examples/api/layered/internal/middleware"

var tracer = otel.Tracer(name)

// Func is a middleware function that wraps an http.Handler.
type Func = func(next http.Handler) http.Handler
