package app

import (
	"context"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"
)

// MiddlewareFunc is a middleware function that wraps an http.Handler.
type MiddlewareFunc = func(next http.Handler) http.Handler

func WrapHandler(handler http.Handler, middlewares ...MiddlewareFunc) http.Handler {
	if len(middlewares) <= 0 {
		return handler
	}

	next := handler

	for _, middleware := range slices.Backward(middlewares) {
		next = middleware(next)
	}

	return next
}

// logging middleware
type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) writeHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

// LoggingMiddleware is a middleware that logs the request method, path, duration, and
// status code.
func LoggingMiddleware(logger *slog.Logger) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		})
	}
}

// recoveryMiddleware is a middleware that recover from panics that occur in the handlers,
// logs the error, and returns a 500 status code.
func RecoveryMiddleware(logger *slog.Logger) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rc := recover(); rc != nil {

					logger.InfoContext(
						r.Context(),
						"panic recovered",
						slog.Any("error", rc),
						slog.Int("status", 500),
					)

					w.WriteHeader(500)
				}
			}()
			next.ServeHTTP(w, r)

		})
	}
}

// traceId middleware
type traceIDKey struct{}

type traceIdOptions struct {
	header string
}

type traceIDOption func(*traceIdOptions)

// WithHeader sets the header name for the trace ID.
func WithHeader(header string) traceIDOption {
	return func(opts *traceIdOptions) {
		opts.header = header
	}
}

func TraceIDMiddleware(options ...traceIDOption) MiddlewareFunc {
	opts := &traceIdOptions{
		header: "",
	}

	for _, opt := range options {
		opt(opts)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				traceID := ""

				if opts.header != "" {
					traceID = r.Header.Get("X-Trace-ID")
				}

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

func getTraceIDAsAtter(ctx context.Context) slog.Attr {
	traceID := getTraceID(ctx)
	if traceID == "" {
		return slog.Attr{}
	}

	return slog.String("trace_id", traceID)
}
