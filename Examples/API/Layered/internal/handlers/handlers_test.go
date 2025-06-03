package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeResponseJSON(t *testing.T) {
	t.Parallel()

	type fields struct {
		status int
		data   any
	}
	type want struct {
		status       int
		body         string
		bodyContains string
	}
	tests := map[string]struct {
		fields fields
		want   want
	}{
		"valid JSON encoding": {
			fields: fields{
				status: http.StatusOK,
				data:   map[string]string{"foo": "bar"},
			},
			want: want{
				status: http.StatusOK,
				body:   `{"foo":"bar"}`,
			},
		},
		"JSON encoding error": {
			fields: fields{
				status: http.StatusCreated,
				data:   make(chan int), // not JSON serializable
			},
			want: want{
				status: http.StatusCreated,
				body:   "",
			},
		},
		"nil data": {
			fields: fields{
				status: http.StatusOK,
				data:   nil,
			},
			want: want{
				status: http.StatusOK,
				body:   `null`,
			},
		},
		"slice data": {
			fields: fields{
				status: http.StatusOK,
				data:   []int{1, 2, 3},
			},
			want: want{
				status: http.StatusOK,
				body:   `[1,2,3]`,
			},
		},
		"struct data": {
			fields: fields{
				status: http.StatusCreated,
				data: struct {
					ID   int    `json:"id"`
					Name string `json:"name"`
				}{ID: 42, Name: "Alice"},
			},
			want: want{
				status: http.StatusCreated,
				body:   `{"id":42,"name":"Alice"}`,
			},
		},
		"string data": {
			fields: fields{
				status: http.StatusAccepted,
				data:   "hello",
			},
			want: want{
				status: http.StatusAccepted,
				body:   `"hello"`,
			},
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rec := httptest.NewRecorder()

			encodeResponseJSON(rec, tc.fields.status, tc.fields.data)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.want.status, res.StatusCode)
			assert.Equal(t, "application/json", res.Header.Get("Content-Type")[:16])

			body := strings.TrimSpace(rec.Body.String())
			if tc.want.body != "" {
				// Compare JSON ignoring whitespace
				var got, want any
				_ = json.Unmarshal([]byte(body), &got)
				_ = json.Unmarshal([]byte(tc.want.body), &want)
				assert.Equal(t, want, got)
			}
			if tc.want.bodyContains != "" {
				assert.Contains(t, body, tc.want.bodyContains)
			}
		})
	}
}
