package telemetry

import (
	"context"
	"net/http"
	"slices"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"example.com/examples/api/layered/internal/middleware"
)

// spanKey is a key type used to store the root span in the context.
type spanKey struct{}

// middlewareFunc is a function type that defines a middleware function.
type middlewareFunc = func(next http.Handler) http.Handler

// InstrumentedServeMux is a wrapper around http.ServeMux that provides OpenTelemetry instrumentation.
type InstrumentedServeMux struct {
	*http.ServeMux
	middlewares []middleware.Func
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

// InstrumentRootHandler wraps the ServeMux's ServeHTTP method with
// OpenTelemetry instrumentation and apply the middlewares in order in which it
// was applied.
//
// Note: this method is intended to be the last method called before adding the
// resulting http.Handler to the http.Server and starting the resulting server.
func (m *InstrumentedServeMux) InstrumentRootHandler(
	opts ...otelhttp.Option,
) http.Handler {
	return otelhttp.NewHandler(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				span := trace.SpanFromContext(ctx)
				ctx = context.WithValue(ctx, spanKey{}, span)

				var handler http.Handler = http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						m.ServeHTTP(w, r)
					},
				)

				for _, md := range slices.Backward(m.middlewares) {
					handler = md(handler)
				}

				handler.ServeHTTP(w, r.WithContext(ctx))
			},
		), "service", opts...,
	)
}

// AddMiddleware adds a middleware function to the InstrumentedServeMux.
// Middleware will be applied in the order they are added, with the last added
// middleware being applied first.
func (m *InstrumentedServeMux) AddMiddleware(md middlewareFunc) {
	m.middlewares = append(m.middlewares, md)
}
