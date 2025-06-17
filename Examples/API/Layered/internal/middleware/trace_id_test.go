package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTraceID(t *testing.T) {
	tests := map[string]struct {
		options       []TraceIDOption
		headerName    string
		headerValue   string
		expectTraceID string
	}{
		"no header option set": {
			options:     nil,
			headerName:  "X-Trace-Id",
			headerValue: "existing-trace-id",
			// Expect a new UUID to be generated
			expectTraceID: "",
		},
		"header option set with existing trace ID": {
			options:       []TraceIDOption{WithHeader("X-Trace-Id")},
			headerName:    "X-Trace-Id",
			headerValue:   "existing-trace-id",
			expectTraceID: "existing-trace-id",
		},
		"header option set with empty trace ID": {
			options:     []TraceIDOption{WithHeader("X-Trace-Id")},
			headerName:  "X-Trace-Id",
			headerValue: "",
			// Expect a new UUID to be generated
			expectTraceID: "",
		},
		"header option set with different header name": {
			options:       []TraceIDOption{WithHeader("X-Custom-Trace-Id")},
			headerName:    "X-Custom-Trace-Id",
			headerValue:   "custom-trace-id",
			expectTraceID: "custom-trace-id",
		},
	}

	for name, tc := range tests {
		t.Run(
			name, func(t *testing.T) {
				// Create a test handler that verifies the trace ID is in the context
				var capturedTraceID string
				testHandler := http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						capturedTraceID = GetTraceID(r.Context())
						w.WriteHeader(http.StatusOK)
					},
				)

				// Apply middleware
				handler := TraceID(tc.options...)(testHandler)

				// Create request with header if needed
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				if tc.headerValue != "" {
					req.Header.Set(tc.headerName, tc.headerValue)
				}

				// Execute request
				recorder := httptest.NewRecorder()
				handler.ServeHTTP(recorder, req)

				// Verify response
				assert.Equal(t, http.StatusOK, recorder.Code)

				// Verify trace ID
				if tc.expectTraceID != "" {
					assert.Equal(t, tc.expectTraceID, capturedTraceID)
				} else {
					// If we expect a generated UUID, validate it's a valid UUID
					_, err := uuid.Parse(capturedTraceID)
					assert.NoError(t, err, "Generated trace ID should be a valid UUID")
					assert.NotEmpty(t, capturedTraceID, "Trace ID should not be empty")
				}
			},
		)
	}
}

func TestGetTraceID(t *testing.T) {
	tests := map[string]struct {
		ctx          context.Context
		expectedID   string
		expectedBool bool
	}{
		"nil context": {
			ctx:          nil,
			expectedID:   "",
			expectedBool: false,
		},
		"context without trace ID": {
			ctx:          context.Background(),
			expectedID:   "",
			expectedBool: false,
		},
		"context with trace ID": {
			ctx:          context.WithValue(context.Background(), traceIDKey{}, "test-trace-id"),
			expectedID:   "test-trace-id",
			expectedBool: true,
		},
		"context with non-string trace ID": {
			ctx:          context.WithValue(context.Background(), traceIDKey{}, 123),
			expectedID:   "",
			expectedBool: false,
		},
	}

	for name, tc := range tests {
		t.Run(
			name, func(t *testing.T) {
				result := GetTraceID(tc.ctx)
				assert.Equal(t, tc.expectedID, result)

				// Additional validation based on expectedBool
				if tc.expectedBool {
					assert.NotEmpty(t, result)
				} else {
					assert.Empty(t, result)
				}
			},
		)
	}
}

func TestGetTraceIDAsAttr(t *testing.T) {
	tests := map[string]struct {
		ctx           context.Context
		expectedAttr  slog.Attr
		expectedValid bool
	}{
		"nil context": {
			ctx:           nil,
			expectedAttr:  slog.Attr{},
			expectedValid: false,
		},
		"context without trace ID": {
			ctx:           context.Background(),
			expectedAttr:  slog.Attr{},
			expectedValid: false,
		},
		"context with trace ID": {
			ctx:           context.WithValue(context.Background(), traceIDKey{}, "test-trace-id"),
			expectedAttr:  slog.String("trace_id", "test-trace-id"),
			expectedValid: true,
		},
	}

	for name, tc := range tests {
		t.Run(
			name, func(t *testing.T) {
				result := GetTraceIDAsAttr(tc.ctx)

				if tc.expectedValid {
					// For valid trace IDs, verify the attribute
					require.Equal(t, tc.expectedAttr.Key, result.Key)
					require.Equal(t, tc.expectedAttr.Value.String(), result.Value.String())
				} else {
					// For invalid trace IDs, expect empty attribute
					assert.Equal(t, slog.Attr{}, result)
				}
			},
		)
	}
}

func TestWithHeader(t *testing.T) {
	tests := map[string]struct {
		headerName string
	}{
		"custom header": {
			headerName: "X-Custom-Header",
		},
		"empty header": {
			headerName: "",
		},
	}

	for name, tc := range tests {
		t.Run(
			name, func(t *testing.T) {
				opts := &traceIDOptions{}
				option := WithHeader(tc.headerName)
				option(opts)

				assert.Equal(
					t, tc.headerName, opts.header,
					"WithHeader should set the header name in options",
				)
			},
		)
	}
}
