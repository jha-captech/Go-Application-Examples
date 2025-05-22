package telemetry

import (
	"context"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type spanKey struct{}

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

func HandleFunc(mux *http.ServeMux, pattern string, handler http.HandlerFunc) {
	mux.HandleFunc(
		pattern, func(w http.ResponseWriter, r *http.Request) {
			setRootSpanName(r.Context(), pattern)
			handler.ServeHTTP(w, r)
		},
	)
}

func Handle(mux *http.ServeMux, pattern string, handler http.Handler) {
	HandleFunc(mux, pattern, handler.ServeHTTP)
}

func NewRootInstrumentedHandler(
	handler http.Handler,
	defaultName string,
	opts ...otelhttp.Option,
) http.Handler {
	return otelhttp.NewHandler(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()

				span := trace.SpanFromContext(ctx)

				ctx = context.WithValue(ctx, spanKey{}, span)

				handler.ServeHTTP(w, r.WithContext(ctx))
			},
		), defaultName, opts...,
	)
}
