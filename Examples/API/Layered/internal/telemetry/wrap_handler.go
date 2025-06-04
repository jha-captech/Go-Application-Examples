package telemetry

import (
	"context"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// spanKey is a key type used to store the root span in the context.
type spanKey struct{}

// InstrumentedServeMux is a wrapper around http.ServeMux that provides OpenTelemetry instrumentation.
type InstrumentedServeMux struct {
	*http.ServeMux
}

// InstrumentServeMux creates a new InstrumentedServeMux that wraps the provided http.ServeMux.
func InstrumentServeMux(mux *http.ServeMux) *InstrumentedServeMux {
	return &InstrumentedServeMux{
		ServeMux: mux,
	}
}

// setRootSpanName sets the name of the root span in the context to the given name.
func setRootSpanName(ctx context.Context, name string) {
	ctxValue := ctx.Value(spanKey{})
	if ctxValue == nil {
		return
	}

	span, ok := ctxValue.(trace.Span)
	if !ok {
		return
	}

	span.SetName(name)
	span.SetAttributes(attribute.String("http.route", name))
}

// HandleFunc wraps the ServeMux's Handle method to set the root span name
func (m *InstrumentedServeMux) HandleFunc(pattern string, handler http.HandlerFunc) {
	m.Handle(pattern, handler)
}

// Handle wraps the ServeMux's Handle method to set the root span name
func (m *InstrumentedServeMux) Handle(pattern string, handler http.Handler) {
	m.ServeMux.Handle(
		pattern, http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				setRootSpanName(r.Context(), pattern)
				handler.ServeHTTP(w, r)
			},
		),
	)
}

// InstrumentRootHandler wraps the ServeMux's ServeHTTP method with OpenTelemetry instrumentation.
func (m *InstrumentedServeMux) InstrumentRootHandler(
	opts ...otelhttp.Option,
) http.Handler {
	return otelhttp.NewHandler(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()

				span := trace.SpanFromContext(ctx)

				ctx = context.WithValue(ctx, spanKey{}, span)

				m.ServeHTTP(w, r.WithContext(ctx))
			},
		), "service", opts...,
	)
}
