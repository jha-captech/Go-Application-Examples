package app

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestCreateUser(t *testing.T) {
	t.Parallel()
	type mockDB struct {
		mockCalled    bool
		mockInputArgs []driver.Value
		mockOutput    *sqlmock.Rows
		mockError     error
	}

	testcases := map[string]struct {
		mockDB
		inputJSON  string
		wantStatus int
		wantUser   userResponse
	}{
		"success": {
			mockDB: mockDB{
				mockCalled:    true,
				mockInputArgs: []driver.Value{"Alice", "alice@example.com", "supersecret"},
				mockOutput:    sqlmock.NewRows([]string{"id"}).AddRow(1),
				mockError:     nil,
			},
			inputJSON:  `{"name":"Alice","email":"alice@example.com","password":"supersecret"}`,
			wantStatus: 201,
			wantUser: userResponse{
				ID:    1,
				Name:  "Alice",
				Email: "alice@example.com",
			},
		},
		"invalid_json": {
			mockDB: mockDB{
				mockCalled:    false,
				mockInputArgs: nil,
				mockOutput:    nil,
				mockError:     nil,
			},
			inputJSON:  `{"name": "Bob", "email": "bob@example.com", "password": password123}`,
			wantStatus: 400,
			wantUser:   userResponse{},
		},
		"request_validation_error": {
			mockDB: mockDB{
				mockCalled:    false,
				mockInputArgs: nil,
				mockOutput:    nil,
				mockError:     nil,
			},
			inputJSON:  `{"name": "Bob", "email": "bob@example.com", "password": "pass"}`,
			wantStatus: 400,
			wantUser:   userResponse{},
		},
		"db_error": {
			mockDB: mockDB{
				mockCalled:    true,
				mockInputArgs: []driver.Value{"Bob", "bob@example.com", "badpass"},
				mockOutput:    sqlmock.NewRows([]string{"id"}),
				mockError:     sqlmock.ErrCancelled,
			},
			inputJSON:  `{"name":"Bob","email":"bob@example.com","password":"password123"}`,
			wantStatus: 500,
			wantUser:   userResponse{},
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

			req := httptest.NewRequest(
				http.MethodPost,
				"/users",
				bytes.NewBufferString(tc.inputJSON),
			)
			rec := httptest.NewRecorder()
			handler := createUser(logger, sqlxDB)
			handler.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, rec.Code)
			}

			if tc.wantStatus == 201 {
				var gotUser userResponse
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
