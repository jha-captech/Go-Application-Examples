package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	tracesdk "go.opentelemetry.io/otel/sdk/trace/tracetest"
	traceapi "go.opentelemetry.io/otel/trace"
)

// mockSpan is a mock implementation of trace.Span for testing
type mockSpan struct {
	traceapi.Span
	name       string
	attributes []attribute.KeyValue
}

func (m *mockSpan) SetName(name string) {
	m.name = name
}

func (m *mockSpan) SetAttributes(attrs ...attribute.KeyValue) {
	m.attributes = append(m.attributes, attrs...)
}

func TestInstrumentServeMux(t *testing.T) {
	tests := map[string]struct {
		expected *InstrumentedServeMux
	}{
		"creates instrumented mux": {},
	}

	for name := range tests {
		t.Run(
			name, func(t *testing.T) {
				mux := http.NewServeMux()
				result := InstrumentServeMux(mux)

				assert.Equal(t, mux, result.ServeMux)
				assert.Equal(t, 0, len(result.middlewares))
			},
		)
	}
}

func TestSetRootSpanName(t *testing.T) {
	tests := map[string]struct {
		createContext func() context.Context
		name          string
		expectedName  string
		expectedAttrs []attribute.KeyValue
	}{
		"sets span name when span exists in context": {
			createContext: func() context.Context {
				mockSpan := &mockSpan{}
				return context.WithValue(context.Background(), spanKey{}, mockSpan)
			},
			name:         "/api/users",
			expectedName: "/api/users",
			expectedAttrs: []attribute.KeyValue{
				attribute.String("http.route", "/api/users"),
			},
		},
		"does nothing when no span in context": {
			createContext: func() context.Context {
				return context.Background()
			},
			name:          "/api/users",
			expectedName:  "",
			expectedAttrs: nil,
		},
		"does nothing when context value is not a span": {
			createContext: func() context.Context {
				return context.WithValue(context.Background(), spanKey{}, "not a span")
			},
			name:          "/api/users",
			expectedName:  "",
			expectedAttrs: nil,
		},
	}

	for name, tc := range tests {
		t.Run(
			name, func(t *testing.T) {
				ctx := tc.createContext()
				setRootSpanName(ctx, tc.name)

				if tc.expectedName != "" {
					span := ctx.Value(spanKey{}).(*mockSpan)
					assert.Equal(t, tc.expectedName, span.name)
					assert.Equal(t, tc.expectedAttrs, span.attributes)
				}
			},
		)
	}
}

func TestInstrumentedServeMux_Handle(t *testing.T) {
	pattern := "/test-path"

	mux := InstrumentServeMux(http.NewServeMux())

	mux.Handle(
		pattern, http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		),
	)

	handlerRootSpan(t, pattern, mux.ServeHTTP)
}

func TestInstrumentedServeMux_HandleFunc(t *testing.T) {
	pattern := "/test-path"

	mux := InstrumentServeMux(http.NewServeMux())

	mux.HandleFunc(
		pattern,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	)

	handlerRootSpan(t, pattern, mux.ServeHTTP)
}

func handlerRootSpan(t *testing.T, expectedSpanName string, handler http.HandlerFunc) {
	// Create a test request
	req := httptest.NewRequest(http.MethodGet, expectedSpanName, nil)
	w := httptest.NewRecorder()

	// setup OpenTelemetry span
	sr := tracesdk.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))

	ctx, span := tp.Tracer("test").Start(context.TODO(), "test-span")
	defer span.End()

	ctx = context.WithValue(ctx, spanKey{}, span)

	req = req.WithContext(ctx)
	req = req.WithContext(context.WithValue(req.Context(), spanKey{}, span))

	// Serve the request
	handler.ServeHTTP(w, req)

	span.End()

	// Verify the response
	assert.Equal(t, http.StatusOK, w.Code)

	spans := sr.Ended()
	require.GreaterOrEqual(t, len(spans), 1)
	assert.Equal(t, expectedSpanName, spans[0].Name())
}

func TestInstrumentedServeMux_InstrumentRootHandler(t *testing.T) {
	// Setup mux with middleware
	mux := InstrumentServeMux(http.NewServeMux())
	requestPath := "/test"

	// Add middleware that adds headers
	mux.AddMiddleware(
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("X-Test-1", "first")
					next.ServeHTTP(w, r)
				},
			)
		},
	)

	mux.AddMiddleware(
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("X-Test-2", "second")
					next.ServeHTTP(w, r)
				},
			)
		},
	)

	// Add a test route
	mux.HandleFunc(
		requestPath, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	)

	// Create and run the test
	handler := mux.InstrumentRootHandler()
	req := httptest.NewRequest("GET", requestPath, nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Validate results
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "first", w.Header().Get("X-Test-1"))
	assert.Equal(t, "second", w.Header().Get("X-Test-2"))
}

func TestInstrumentedServeMux_AddMiddleware(t *testing.T) {
	// Setup
	mux := InstrumentServeMux(http.NewServeMux())
	initialLen := len(mux.middlewares)

	// Add a middleware
	testMiddleware := func(next http.Handler) http.Handler {
		return next
	}
	mux.AddMiddleware(testMiddleware)

	// Verify middleware was added
	assert.Equal(t, initialLen+1, len(mux.middlewares))
	// Cannot directly compare functions, but we can verify it's not nil
	assert.NotNil(t, mux.middlewares[initialLen])
}
