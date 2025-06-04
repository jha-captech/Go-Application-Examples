package app

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestListUsers(t *testing.T) {
	type mockDB struct {
		mockRows  *sqlmock.Rows
		mockError error
	}

	testcases := map[string]struct {
		mockDB
		wantStatus int
		wantUsers  []userResponse
	}{
		"success": {
			mockDB: mockDB{
				mockRows: sqlmock.NewRows([]string{"id", "name", "email", "password"}).
					AddRow(1, "Alice", "alice@example.com", "pw1").
					AddRow(2, "Bob", "bob@example.com", "pw2"),
				mockError: nil,
			},
			wantStatus: http.StatusOK,
			wantUsers: []userResponse{
				{ID: 1, Name: "Alice", Email: "alice@example.com"},
				{ID: 2, Name: "Bob", Email: "bob@example.com"},
			},
		},
		"empty": {
			mockDB: mockDB{
				mockRows:  sqlmock.NewRows([]string{"id", "name", "email", "password"}),
				mockError: nil,
			},
			wantStatus: http.StatusOK,
			wantUsers:  []userResponse{},
		},
		"scan_error": {
			mockDB: mockDB{
				mockRows: sqlmock.NewRows([]string{"id", "name", "email", "password"}).
					AddRow("bad_id", "Charlie", "charlie@example.com", "pw3"),
				mockError: nil,
			},
			wantStatus: http.StatusInternalServerError,
			wantUsers:  nil,
		},
		"db_error": {
			mockDB: mockDB{
				mockRows:  nil,
				mockError: errors.New("db error"),
			},
			wantStatus: http.StatusInternalServerError,
			wantUsers:  nil,
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

			expect := mock.ExpectQuery(regexp.QuoteMeta(`
						SELECT id, name, email, password
						FROM users
					`))
			if tc.mockRows != nil {
				expect.WillReturnRows(tc.mockRows)
			}
			expect.WillReturnError(tc.mockError)

			req := httptest.NewRequest("GET", "/user", nil)
			rec := httptest.NewRecorder()
			handler := listUsers(logger, sqlxDB)
			handler.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, rec.Code)
			}

			if tc.wantStatus == http.StatusOK {
				var gotUsers []userResponse
				if err := json.NewDecoder(rec.Body).Decode(&gotUsers); err != nil {
					t.Errorf("failed to decode response body: %v", err)
				}

				if len(gotUsers) != len(tc.wantUsers) {
					t.Errorf("want %d users, got %d", len(tc.wantUsers), len(gotUsers))
				}

				for i := range gotUsers {
					if gotUsers[i] != tc.wantUsers[i] {
						t.Errorf("want user %+v, got %+v", tc.wantUsers[i], gotUsers[i])
					}
				}
			}
		})
	}
}
