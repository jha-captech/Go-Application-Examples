package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type traceIDKey struct{}

type traceIdOptions struct {
	header string
}

type TraceIDOption func(*traceIdOptions)

// WithHeader sets the header name for the trace ID.
func WithHeader(header string) TraceIDOption {
	return func(opts *traceIdOptions) {
		opts.header = header
	}
}

func TraceID(options ...TraceIDOption) Func {
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

func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	traceID, ok := ctx.Value(traceIDKey{}).(string)
	if !ok {
		return ""
	}

	return traceID
}

func GetRaceIDAsAtter(ctx context.Context) slog.Attr {
	traceID := GetTraceID(ctx)
	if traceID == "" {
		return slog.Attr{}
	}

	return slog.String("trace_id", traceID)
}
