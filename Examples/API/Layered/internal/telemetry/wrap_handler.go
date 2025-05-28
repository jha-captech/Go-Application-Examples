package telemetry

import (
	"context"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type spanKey struct{}

type InstrumentedServeMux struct {
	*http.ServeMux
}

func InstrumentServeMux(mux *http.ServeMux) *InstrumentedServeMux {
	return &InstrumentedServeMux{
		ServeMux: mux,
	}
}

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

func (m *InstrumentedServeMux) HandleFunc(pattern string, handler http.HandlerFunc) {
	m.Handle(pattern, handler)
}

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

func (m *InstrumentedServeMux) InstrumentRootHandler(
	defaultName string,
	opts ...otelhttp.Option,
) http.Handler {
	return otelhttp.NewHandler(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()

				span := trace.SpanFromContext(ctx)

				ctx = context.WithValue(ctx, spanKey{}, span)

				m.ServeMux.ServeHTTP(w, r.WithContext(ctx))
			},
		), defaultName, opts...,
	)
}
