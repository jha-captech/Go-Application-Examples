package app

import (
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

func TestReadUser(t *testing.T) {
    testcases := map[string]struct {
        mockCalled    bool
        mockInputArgs []driver.Value
        mockRows      *sqlmock.Rows
        mockError     error
        id            string
        wantStatus    int
        wantUser      User
    }{
        "success": {
            mockCalled:    true,
            mockInputArgs: []driver.Value{1},
            mockRows: sqlmock.NewRows([]string{"id", "name", "email", "password"}).
                AddRow(1, "Alice", "alice@example.com", "supersecret"),
            mockError:  nil,
            id:         "1",
            wantStatus: http.StatusOK,
            wantUser:   User{ID: 1, Name: "Alice", Email: "alice@example.com", Password: "supersecret"},
        },
        "invalid_id": {
            mockCalled:    false,
            mockInputArgs: nil,
            mockRows:      nil,
            mockError:     nil,
            id:            "abc",
            wantStatus:    http.StatusBadRequest,
            wantUser:      User{},
        },
        "not_found": {
            mockCalled:    true,
            mockInputArgs: []driver.Value{2},
            mockRows:      sqlmock.NewRows([]string{"id", "name", "email", "password"}),
            mockError:     sql.ErrNoRows,
            id:            "2",
            wantStatus:    http.StatusNotFound,
            wantUser:      User{},
        },
        "db_error": {
            mockCalled:    true,
            mockInputArgs: []driver.Value{3},
            mockRows:      sqlmock.NewRows([]string{"id", "name", "email", "password"}),
            mockError:     errors.New("db error"),
            id:            "3",
            wantStatus:    http.StatusInternalServerError,
            wantUser:      User{},
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
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT id,
						name,
						email,
						password
					FROM users
					WHERE id = $1::int
					`)).
					WithArgs(tc.mockInputArgs...).
					WillReturnRows(tc.mockRows).
					WillReturnError(tc.mockError)
			}

            req := httptest.NewRequest("GET", "/user/"+tc.id, nil)
            req.SetPathValue("id", tc.id)
            rec := httptest.NewRecorder()
            handler := readUser(logger, db)
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