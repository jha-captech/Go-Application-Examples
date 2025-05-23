package app

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"log/slog"
	"net/http/httptest"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestCreateUser(t *testing.T) {
	testcases := map[string]struct {
		mockCalled    bool
		mockInputArgs []driver.Value
		mockOutput    *sqlmock.Rows
		mockError     error
		inputJSON     string
		wantStatus    int
		wantUser      User
	}{
		"success": {
			mockCalled:    true,
			mockInputArgs: []driver.Value{"Alice", "alice@example.com", "supersecret"},
			mockOutput:    sqlmock.NewRows([]string{"id"}).AddRow(1),
			mockError:     nil,
			inputJSON:     `{"name":"Alice","email":"alice@example.com","password":"supersecret"}`,
			wantStatus:    201,
			wantUser: User{
				ID:       1,
				Name:     "Alice",
				Email:    "alice@example.com",
				Password: "supersecret",
			},
		},
		"invalid_json": {
			mockCalled:    false,
			mockInputArgs: nil,
			mockOutput:    nil,
			mockError:     nil,
			inputJSON:     `{"name": "Bob", "email": "bob@example.com", "password": badjson}`,
			wantStatus:    400,
			wantUser:      User{},
		},
		"db_error": {
			mockCalled:    true,
			mockInputArgs: []driver.Value{"Bob", "bob@example.com", "badpass"},
			mockOutput:    sqlmock.NewRows([]string{"id"}),
			mockError:     sqlmock.ErrCancelled,
			inputJSON:     `{"name":"Bob","email":"bob@example.com","password":"badpass"}`,
			wantStatus:    500,
			wantUser:      User{},
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
				mock.
					ExpectQuery(regexp.QuoteMeta(`
						INSERT INTO users (name, email, password)
						VALUES ($1, $2, $3)
						RETURNING id
					`)).
					WithArgs(tc.mockInputArgs...).
					WillReturnRows(tc.mockOutput).
					WillReturnError(tc.mockError)
			}

			req := httptest.NewRequest("POST", "/users", bytes.NewBuffer([]byte(tc.inputJSON)))
			rec := httptest.NewRecorder()
			handler := createUser(logger, sqlxDB)
			handler.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, rec.Code)
			}

			if tc.wantStatus == 201 {
				var gotUser User
				if err := json.NewDecoder(rec.Body).Decode(&gotUser); err != nil {
					t.Errorf("failed to decode response body: %v", err)
				}
				if !reflect.DeepEqual(gotUser, tc.wantUser) {
					t.Errorf("want user %+v, got %+v", tc.wantUser, gotUser)
				}
			}
		})
	}
}
