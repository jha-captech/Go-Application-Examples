package app

import (
	"database/sql/driver"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestDeleteUser(t *testing.T) {
	testcases := map[string]struct {
		mockCalled    bool
		mockInputArgs []driver.Value
		mockResult    driver.Result
		mockError     error
		id            string
		wantStatus    int
	}{
		"success": {
			mockCalled:    true,
			mockInputArgs: []driver.Value{1},
			mockResult:    sqlmock.NewResult(0, 1), // 1 row affected
			mockError:     nil,
			id:            "1",
			wantStatus:    http.StatusNoContent,
		},
		"invalid_id": {
			mockCalled:    false,
			mockInputArgs: nil,
			mockResult:    nil,
			mockError:     nil,
			id:            "abc",
			wantStatus:    http.StatusBadRequest,
		},
		"db_error": {
			mockCalled:    true,
			mockInputArgs: []driver.Value{2},
			mockResult:    nil,
			mockError:     errors.New("db error"),
			id:            "2",
			wantStatus:    http.StatusInternalServerError,
		},
		"not_found": {
			mockCalled:    true,
			mockInputArgs: []driver.Value{3},
			mockResult:    sqlmock.NewResult(0, 0), // 0 rows affected
			mockError:     nil,
			id:            "3",
			wantStatus:    http.StatusNotFound,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			logger := slog.New(slog.DiscardHandler)
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unexpected error when opening stub db: %v", err)
			}
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "pgx")

			if tc.mockCalled {
				mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users WHERE id = $1")).
					WithArgs(tc.mockInputArgs...).
					WillReturnResult(tc.mockResult).
					WillReturnError(tc.mockError)
			}

			req := httptest.NewRequest("DELETE", "/user/"+tc.id, nil)
			req.SetPathValue("id", tc.id)
			rec := httptest.NewRecorder()
			handler := deleteUser(logger, sqlxDB)
			handler.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, rec.Code)
			}
		})
	}
}
