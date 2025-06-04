package app

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestHandleHealthCheck(t *testing.T) {
	t.Parallel()

	type fields struct {
		dbErr error
	}

	testcases := map[string]struct {
		fields fields
		wantStatus     int
		wantStatusBody string
	}{
		"healthy": {
			fields: fields{
				dbErr: nil,
			},
			wantStatus:     http.StatusOK,
			wantStatusBody: `{"status":"healthy","details":[{"name":"db","status":"healthy"}]}`,
		},
		"unhealthy": {
			fields: fields{
				dbErr: errors.New("db connection error"),
			},
			wantStatus:     http.StatusInternalServerError,
			wantStatusBody: `{"status":"unhealthy","details":[{"name":"db","status":"unhealthy"}]}`,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			assert.NoError(t, err)
			defer db.Close()

			// Mock DB ping
			mock.ExpectPing().WillReturnError(tc.fields.dbErr)
			
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()

			handler := HandleHealthCheck(slog.Default(), sqlx.NewDb(db, "sqlmock"))
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)

			// Compare JSON bodies (ignoring whitespace and field order)
			var gotBody, wantBody map[string]interface{}
			err = json.Unmarshal(rec.Body.Bytes(), &gotBody)
			assert.NoError(t, err)
			err = json.Unmarshal([]byte(tc.wantStatusBody), &wantBody)
			assert.NoError(t, err)
			assert.Equal(t, wantBody, gotBody)
		})
	}
}
