package app

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestUpdateUser(t *testing.T) {
	testcases := map[string]struct {
		id         string
		body       any
		mockCalled bool
		mockArgs   []driver.Value
		mockRow    *sqlmock.Rows
		mockError  error
		wantStatus int
		wantUser   User
	}{
		"success": {
			id:         "1",
			body:       User{Name: "Alice", Email: "alice@new.com", Password: "pw"},
			mockCalled: true,
			mockArgs:   []driver.Value{"Alice", "alice@new.com", "pw", 1},
			mockRow: sqlmock.NewRows([]string{"id", "name", "email", "password"}).
				AddRow(1, "Alice", "alice@new.com", "pw"),
			mockError:  nil,
			wantStatus: http.StatusOK,
			wantUser:   User{ID: 1, Name: "Alice", Email: "alice@new.com", Password: "pw"},
		},
		"invalid_id": {
			id:         "abc",
			body:       User{Name: "Bob", Email: "bob@new.com", Password: "pw"},
			mockCalled: false,
			wantStatus: http.StatusBadRequest,
		},
		"bad_json": {
			id:         "2",
			body:       "{bad json}",
			mockCalled: false,
			wantStatus: http.StatusBadRequest,
		},
		"not_found": {
			id:         "3",
			body:       User{Name: "Carol", Email: "carol@new.com", Password: "pw"},
			mockCalled: true,
			mockArgs:   []driver.Value{"Carol", "carol@new.com", "pw", 3},
			mockRow:    sqlmock.NewRows([]string{"id", "name", "email", "password"}),
			mockError:  sql.ErrNoRows,
			wantStatus: http.StatusNotFound,
		},
		"db_error": {
			id:         "4",
			body:       User{Name: "Dave", Email: "dave@new.com", Password: "pw"},
			mockCalled: true,
			mockArgs:   []driver.Value{"Dave", "dave@new.com", "pw", 4},
			mockRow:    sqlmock.NewRows([]string{"id", "name", "email", "password"}),
			mockError:  errors.New("db error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			logger := slog.Default()
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unexpected error when opening stub db: %v", err)
			}
			defer db.Close()

			if tc.mockCalled {
				expect := mock.ExpectQuery(regexp.QuoteMeta(`
					UPDATE users
					SET name = $1, email = $2, password = $3
					WHERE id = $4
					RETURNING id, name, email, password
				`)).WithArgs(tc.mockArgs...)
				if tc.mockRow != nil {
					expect.WillReturnRows(tc.mockRow)
				}
				expect.WillReturnError(tc.mockError)
			}

			var reqBody []byte
			switch v := tc.body.(type) {
			case User:
				reqBody, _ = json.Marshal(v)
			case string:
				reqBody = []byte(v)
			default:
				reqBody = nil
			}

			req := httptest.NewRequest("PUT", "/user/"+tc.id, bytes.NewReader(reqBody))
			req.SetPathValue("id", tc.id)
			rec := httptest.NewRecorder()
			handler := updateUser(logger, db)
			handler.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, rec.Code)
			}

			if tc.wantStatus == http.StatusOK {
				var gotUser User
				if err := json.NewDecoder(rec.Body).Decode(&gotUser); err != nil {
					t.Errorf("failed to decode response body: %v", err)
				}
				if gotUser != tc.wantUser {
					t.Errorf("want user %+v, got %+v", tc.wantUser, gotUser)
				}
			}
		})
	}
}
