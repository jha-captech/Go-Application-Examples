package app

import (
	"context"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"
)

// middlewareFunc is a middleware function that wraps an http.Handler.
type middlewareFunc = func(next http.Handler) http.Handler

// WrapHandler applies a list of middlewares to an http.Handler in reverse order.
func WrapHandler(handler http.Handler, middlewares ...middlewareFunc) http.Handler {
	if len(middlewares) == 0 {
		return handler
	}

	next := handler

	for _, middleware := range slices.Backward(middlewares) {
		next = middleware(next)
	}

	return next
}

// wrappedWriter is a custom http.ResponseWriter that captures the status code.
type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

// LoggingMiddleware logs the HTTP request method, path, duration, and status code.
func LoggingMiddleware(logger *slog.Logger) middlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				start := time.Now()

				wrapped := &wrappedWriter{
					ResponseWriter: w,
					statusCode:     http.StatusOK,
				}

				next.ServeHTTP(wrapped, r)

				logger.InfoContext(
					r.Context(),
					"request completed",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("duration", time.Since(start).String()),
					slog.Int("status", wrapped.statusCode),
				)
			},
		)
	}
}

// RecoveryMiddleware recovers from panics in the handler chain, logs the error, and returns a 500 status code.
func RecoveryMiddleware(logger *slog.Logger) middlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer func() {
					if rc := recover(); rc != nil {
						logger.InfoContext(
							r.Context(),
							"panic recovered",
							slog.Any("error", rc),
							slog.Int("status", http.StatusInternalServerError),
						)

						w.WriteHeader(http.StatusInternalServerError)
					}
				}()
				next.ServeHTTP(w, r)
			},
		)
	}
}

// traceIDKey is a unique type for storing the trace ID in the context.
type traceIDKey struct{}

// TraceIDMiddleware injects a trace ID into the request context.
func TraceIDMiddleware() middlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				traceID := r.Header.Get("X-Trace-Id")

				if traceID == "" {
					traceID = uuid.NewString()
				}

				// Set the trace ID in the request context
				ctx := r.Context()
				ctx = context.WithValue(ctx, traceIDKey{}, traceID)
				r = r.WithContext(ctx)

				// Call the next handler
				next.ServeHTTP(w, r)
			},
		)
	}
}

// getTraceID retrieves the trace ID from the context, if present.
func getTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	traceID, ok := ctx.Value(traceIDKey{}).(string)
	if !ok {
		return ""
	}

	return traceID
}

// getTraceIDAsAttr returns the trace ID as a slog.Attr for structured logging.
func getTraceIDAsAttr(ctx context.Context) slog.Attr {
	traceID := getTraceID(ctx)
	if traceID == "" {
		return slog.Attr{}
	}

	return slog.String("trace_id", traceID)
}
